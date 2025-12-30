package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

var (
	sURL    string
	sKey    string
	// LIMITADOR GLOBAL: 5 peticiones por segundo, ráfaga de 10.
	limiter = rate.NewLimiter(5, 10)
)

func main() {

	godotenv.Load()
	sURL = os.Getenv("SUPABASE_URL")
	sKey = os.Getenv("SUPABASE_KEY")
	rand.Seed(time.Now().UnixNano())

	// Goroutine del Latido (Heartbeat)
	    go func() {
	        for {
	            actualizarStatusMotor("online")
	            time.Sleep(10 * time.Second)
	        }
	    }()

	    // Goroutine para apagar limpiamente (Ctrl+C)
	    c := make(chan os.Signal, 1)
	    signal.Notify(c, os.Interrupt)
	    go func() {
	        <-c
	        fmt.Println("\n🛑 Apagando motor...")
	        actualizarStatusMotor("offline")
	        os.Exit(0)
	    }()

	fmt.Println("🛰️  Motor Seguro iniciado. Rate Limit Global Activo.")

	for {
		partidaID := buscarPartidaEsperando()
		if partidaID != "" {
			fmt.Printf("🚀 Partida %s detectada. Iniciando sorteo...\n", partidaID)
			procesarSorteo(partidaID)
		}
		// Respiro del procesador para evitar uso innecesario de CPU
		time.Sleep(3 * time.Second)
	}
}

// Función de apoyo para actualizar el status
func actualizarStatusMotor(status string) {

    url := sURL + "/rest/v1/sistema?id=eq.1"
    body := []byte(`{"motor_status": "` + status + `", "last_ping": "now()"}`)
    
    req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
    req.Header.Set("apikey", sKey)
    req.Header.Set("Authorization", "Bearer "+sKey)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err == nil {
        defer resp.Body.Close()
    }
}

// FUNCIÓN MAESTRA: Maneja Seguridad (Rate Limit) y Cabeceras (Headers)
func ejecutarPeticionSupabase(metodo, endpoint string, data interface{}) {
    // Definimos un tiempo máximo de 10 segundos para toda la operación
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // El limiter respeta el tiempo límite
    err := limiter.Wait(ctx)
    if err != nil {
        fmt.Printf("❌ Espera de límite de tasa cancelada: %v\n", err)
        return
    }

    url := sURL + endpoint
    var bodyReader io.Reader
    if data != nil {
        jsonData, _ := json.Marshal(data)
        bodyReader = bytes.NewBuffer(jsonData)
    }

    req, _ := http.NewRequestWithContext(ctx, metodo, url, bodyReader)

    req.Header.Set("apikey", sKey)
    req.Header.Set("Authorization", "Bearer "+sKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Prefer", "return=minimal")

    client := &http.Client{}
    resp, err := client.Do(req)
    
    if err != nil {
        fmt.Printf("❌ Error al ejecutar %s: %v\n", metodo, err)
        return
    }
    defer resp.Body.Close()

    fmt.Printf("✅ %s en %s finalizado (Status: %d)\n", metodo, endpoint, resp.StatusCode)
}

func buscarPartidaEsperando() string {
    limiter.Wait(context.Background())

    url := sURL + "/rest/v1/partidas?estado=eq.pendiente&limit=1"
    
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("apikey", sKey)
    req.Header.Set("Authorization", "Bearer "+sKey)

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil || resp.StatusCode != 200 {
        return ""
    }
    defer resp.Body.Close()

    var resultados []map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&resultados)

    if len(resultados) > 0 {
        // Retorna el ID de la primera partida que encuentre
        return resultados[0]["id"].(string)
    }
    return ""
}

func procesarSorteo(pID string) {
	ganadorEncontrado := false
	actualizarEstado(pID, "jugando")
	bolasCantadas := make(map[int]bool)
	var listaOrdenadaBolas []int // Aquí guardaremos la secuencia para el historial

	carton := obtenerCartonUsuario(pID)

	for i := 1; i <= 35; i++ {
		var bola int
		for {
			n := rand.Intn(35) + 1
			if !bolasCantadas[n] {
				bola = n
				bolasCantadas[n] = true
				listaOrdenadaBolas = append(listaOrdenadaBolas, n) // Guardamos la bola
				break
			}
		}

		insertarBola(pID, bola)
		fmt.Printf("🔮 Bola %d: [%d]\n", i, bola)

		if comprobarVictoria(carton, bolasCantadas) {
			ganadorEncontrado = true
			break
		}
		time.Sleep(1 * time.Second) // 1s para que sea fluido
	}

	resultadoFinal := "perdida"
	if ganadorEncontrado {
		resultadoFinal = "ganada"
		fmt.Println("🏆 ¡HAY UN GANADOR!")
	}

	// --- LOGICA DE COMPACTACIÓN Y LIMPIEZA ---
	fmt.Println("🧹 Compactando datos y limpiando tabla temporal...")
	
	// 1. Guardamos el array completo en la tabla 'partidas'
	actualizarDatosFinales(pID, resultadoFinal, listaOrdenadaBolas)

	// 2. Borramos las filas de 'numeros_cantados' para liberar espacio
	limpiarNumerosTemporales(pID)

	actualizarEstado(pID, "finalizada")
}

func actualizarDatosFinales(pID string, resultado string, bolas []int) {
    endpoint := "/rest/v1/partidas?id=eq." + pID
    data := map[string]interface{}{
        "resultado":       resultado,
        "historial_bolas": bolas,
    }
    // Añadimos un log para saber qué está pasando
    fmt.Println("📡 Enviando historial a Supabase...")
    ejecutarPeticionSupabase("PATCH", endpoint, data)
}

func limpiarNumerosTemporales(pID string) {
    endpoint := "/rest/v1/numeros_cantados?partida_id=eq." + pID
    fmt.Println("🗑️ Borrando bolas temporales...")
    ejecutarPeticionSupabase("DELETE", endpoint, nil)
}
func actualizarEstado(id, nuevoEstado string) {
	endpoint := "/rest/v1/partidas?id=eq." + id
	ejecutarPeticionSupabase("PATCH", endpoint, map[string]string{"estado": nuevoEstado})
}

func insertarBola(pID string, valor int) {
	endpoint := "/rest/v1/numeros_cantados"
	ejecutarPeticionSupabase("POST", endpoint, map[string]interface{}{"partida_id": pID, "valor": valor})
}

// Esta función descarga los 15 números que el usuario eligió en Svelte
func obtenerCartonUsuario(pID string) []int {
	url := sURL + "/rest/v1/partidas?id=eq." + pID + "&select=numeros_carton"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", sKey)
	req.Header.Set("Authorization", "Bearer "+sKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return []int{}
	}
	defer resp.Body.Close()

	var resultados []struct {
		NumerosCarton []int `json:"numeros_carton"`
	}
	json.NewDecoder(resp.Body).Decode(&resultados)

	if len(resultados) > 0 {
		return resultados[0].NumerosCarton
	}
	return []int{}
}

// Esta función compara el cartón con las bolas que han salido
func comprobarVictoria(carton []int, bolasCantadas map[int]bool) bool {
	if len(carton) == 0 {
		return false
	}
	for _, num := range carton {
		if !bolasCantadas[num] {
			// Si falta aunque sea un número, no hay bingo todavía
			return false
		}
	}
	// Si pasó por todos y todos están en bolasCantadas... ¡BINGO!
	return true
}