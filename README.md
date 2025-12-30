# 🚀 Bingo Espacial Autónomo

¡Bienvenido a bordo, Capitán! Este es un sistema de Bingo Galáctico diseñado para funcionar de forma autónoma. El motor (Backend) gestiona el sorteo de forma segura mientras el radar (Frontend) muestra la acción en tiempo real.

## 🛠️ Tecnologías Espaciales

* **Motor (Backend):** [Go](https://go.dev/) - Encargado de la lógica del sorteo, validación de victorias y limpieza de datos.
* **Radar (Frontend):** [SvelteKit](https://kit.svelte.dev/) - Interfaz reactiva con animaciones de "Generador de Plasma".
* **Centro de Mando (DB):** [Supabase](https://supabase.com/) - Base de datos PostgreSQL con Realtime activado.

## 📡 Características Principales

- **Motor Autónomo:** El motor en Go detecta nuevas partidas y arranca el sorteo sin intervención humana.
- **Compactación de Datos:** Al terminar el juego, el sistema resume 75 filas de datos en un solo array para ahorrar espacio.
- **Latido de Seguridad:** El motor reporta su estado (`online/offline`) cada 10 segundos.
- **Generador de Plasma:** Algoritmo de selección aleatoria de cartones en el cliente.

## 🚀 Instalación y Despegue

1. **Configurar la Base de Datos:**
   - Crear tablas: `partidas`, `numeros_cantados` y `sistema`.
   - Activar **Realtime** en las publicaciones de Supabase.

2. **Backend (Motor):**
   ```bash
   cd motor-bingo
   go mod tidy
   go run main.go

3. Frontend:

   npm install
   npm run dev

📜 Licencia

Este proyecto está bajo la Licencia MIT. ¡Siéntete libre de usarlo para conquistar la galaxia!

