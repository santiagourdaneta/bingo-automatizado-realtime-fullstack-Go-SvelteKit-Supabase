<script>
    import { onMount } from 'svelte';
    import { supabase } from '$lib/supabase';

    let estado = 'esperando'; 
    let seleccionados = new Set();
    let numerosCantados = [];
    let mensajeIA = "Bienvenido, Capitán. Elige tus 15 coordenadas espaciales.";
    let soyGanador = false;
    let mostrarPremio = false;
    let mostrarMenuFinal = false;
    let activandoLimpieza = false;
    let porcentajeVictorias = 0;
    let modoTristeza = false;

    async function cargarEstadisticas() {
        const { data } = await supabase.from('historial').select('*');
        if (data && data.length > 0) {
            const ganadas = data.filter(h => h.resultado === 'ganada').length;
            porcentajeVictorias = Math.round((ganadas / data.length) * 100);
        }
    }

    function toggleNumero(num) {
        if (seleccionados.has(num)) {
            seleccionados.delete(num);
        } else if (seleccionados.size < 15) {
            seleccionados.add(num);
        }
        seleccionados = seleccionados;
    }

    function generadorPlasma() {
        let nuevos = new Set();
        while(nuevos.size < 15) nuevos.add(Math.floor(Math.random() * 75) + 1);
        seleccionados = nuevos;
    }

  async function confirmarYJugar() {

    // Primero revisamos si ya hay una partida para no crear una fantasma
        const { data: existentes } = await supabase
            .from('partidas')
            .select('id')
            .or('estado.eq.pendiente,estado.eq.jugando');

        if (existentes && existentes.length > 0) {
            alert("¡Ya hay una partida en curso! Espera a que termine.");
            return;
        }

        // Si no hay ninguna, procedemos normalmente...
      
      if (seleccionados.size !== 15) return;
      
      // Convertimos el Set a un Array puro de números
      const arrayNumeros = Array.from(seleccionados).map(n => parseInt(n));

      console.log("Enviando cartón:", arrayNumeros); 

      const { data, error } = await supabase
          .from('partidas')
          .insert([
              { 
                  numeros_carton: arrayNumeros, 
                  estado: 'pendiente' 
              }
          ])
          .select();

      if (error) {
          console.error("Error de Supabase:", error.message);
          mensajeIA = "Error: " + error.message;
      } else {
          estado = 'jugando';
          mensajeIA = "🚀 ¡Coordenadas aceptadas! Motor en marcha.";
      }
  }

    function resetearEstadoParaNuevaPartida() {
        activandoLimpieza = true;
        seleccionados = new Set();
        numerosCantados = [];
        soyGanador = false;
        estado = 'esperando';
        mostrarPremio = false;
        setTimeout(() => { activandoLimpieza = false; }, 800);
    }

   let motorOnline = false; 

       onMount(async () => {

        // REVISAR SI YA HAY UNA PARTIDA EN CURSO
            const { data: partidaActiva } = await supabase
                .from('partidas')
                .select('*')
                .or('estado.eq.pendiente,estado.eq.jugando') // Busca partidas activas
                .limit(1)
                .single();

            if (partidaActiva) {
                // Si existe, nos sincronizamos con ella automáticamente
                seleccionados = new Set(partidaActiva.numeros_carton);
                estado = 'jugando';
                mensajeIA = "Reconectando con la misión en curso...";
            }

           // Verificar estado inicial del motor
           const { data, error } = await supabase
               .from('sistema')
               .select('motor_status')
               .eq('id', 1)
               .single();

           console.log("Dato del motor recibido:", data); 
           if (data) motorOnline = (data.motor_status === 'online');

           // Escuchar cambios del motor en tiempo real
           const motorChannel = supabase
               .channel('status-motor')
               .on('postgres_changes', { 
                   event: 'UPDATE', 
                   schema: 'public', 
                   table: 'sistema',
                   filter: 'id=eq.1' 
               }, payload => {
                   console.log("📡 Cambio de estado del motor:", payload.new.motor_status);
                   motorOnline = (payload.new.motor_status === 'online');
               })
               .subscribe();

           // Tu suscripción actual de bolas y partidas
              const juegoChannel = supabase
                  .channel('cambios-juego')
                  .on('postgres_changes', { 
                      event: 'INSERT', 
                      schema: 'public', 
                      table: 'numeros_cantados' 
                  }, payload => {
                      console.log("¡Bola recibida en tiempo real!", payload.new.valor);
                      // Usamos una función de actualización para asegurar que Svelte reaccione
                      numerosCantados = [...numerosCantados, Number(payload.new.valor)];
                  })
               .on('postgres_changes', { event: 'UPDATE', schema: 'public', table: 'partidas' }, payload => {
                   if (payload.new.estado === 'finalizada') {
                       soyGanador = (payload.new.resultado === 'ganada');
                       mostrarMenuFinal = true;
                       mensajeIA = soyGanador ? "¡Bingo confirmado!" : "Misión fallida.";
                   }
               })
             .subscribe((status) => {
                     console.log("Estado de la suscripción Realtime:", status);
                 });

           return () => {
               motorChannel.unsubscribe();
               juegoChannel.unsubscribe();
           };
       });
</script>

<main class:fondo-triste={modoTristeza}>
    <header>
        <h1>🚀 Bingo IA Autónomo</h1>
        <div class="status-badge" class:active={estado === 'jugando'}>
            ESTADO: {estado.toUpperCase()}
        </div>
        <p class="ia-msg">{mensajeIA}</p>
    </header>

    {#if soyGanador}
        <div class="win-banner">¡ERES EL GANADOR! 🏆</div>
    {/if}

    {#if mostrarPremio}
        <div class="premio-modal">
            <div class="modal-content">
                <h2>🎁 ¡REGALO ESPACIAL!</h2>
                <button class="btn-next" on:click={resetearEstadoParaNuevaPartida}>ACEPTAR</button>
            </div>
        </div>
    {/if}

    <section class="game-area efecto-limpieza" class:barrido={activandoLimpieza}>
        {#if estado === 'esperando'}
            <div class="selector-container">
                <div class="grid-75">
                    {#each Array(75) as _, i}
                        <button class="celda" class:dorado={seleccionados.has(i + 1)} on:click={() => toggleNumero(i + 1)}>
                            {i + 1}
                        </button>
                    {/each}
                </div>
                <button class="btn-plasma" on:click={generadorPlasma}>GENERADOR DE PLASMA 🧪</button>
                <button 
                    class="btn-jugar" 
                    disabled={!motorOnline || seleccionados.size < 15} 
                    on:click={confirmarYJugar}
                    style={!motorOnline ? "filter: grayscale(1); cursor: not-allowed;" : ""}
                >
                    {#if !motorOnline}
                        📡 BUSCANDO SEÑAL DEL MOTOR...
                    {:else}
                        ¡INICIAR DESPEGUE! ({seleccionados.size}/15)
                    {/if}
                </button>
            </div>
        {:else}
            <div class="grid-carton">
                {#each Array.from(seleccionados) as num}
                    <div class="bola-carton" class:marcada={numerosCantados.includes(num)}>{num}</div>
                {/each}
            </div>
        {/if}
    </section>

    {#if mostrarMenuFinal}
        <section class="final-actions">
            <button class="btn-next" on:click={resetearEstadoParaNuevaPartida}>¡SIGUIENTE VUELO! 🚀</button>
        </section>
    {/if}

    <section class="visual-stats">
        <div class="chart-circle" style="--porcentaje: {porcentajeVictorias}%">
            <span>{porcentajeVictorias}% Victoria</span>
        </div>
        <div class="contador-bolas">
            🚀 Progreso del Sorteo: <strong>{numerosCantados.length}</strong> / 35
        </div>
        <div class="balls-history">
            {#each numerosCantados.slice().reverse() as num}
                <span class="ball">{num}</span>
            {/each}
        </div>
    </section>
</main>

<style>
    :global(body) { background: #0f172a; color: white; font-family: 'Inter', sans-serif; margin: 0; }

    main {
        
        transform-origin: top center; /* Para que no queden huecos blancos arriba */
        
        max-width: 650px; 
        margin: 0 auto;
        width: 100%;
    }

    h1 { font-family: 'Orbitron', sans-serif; color: #60a5fa; }
    
    .status-badge { background: #1e293b; padding: 5px 15px; border-radius: 20px; display: inline-block; font-size: 0.8rem; border: 1px solid #3b82f6; }
    .status-badge.active { background: #1d4ed8; box-shadow: 0 0 10px #3b82f6; }

    .grid-75 {
        display: grid;
        /* Reducimos el tamaño de las celdas proporcionalmente */
        grid-template-columns: repeat(auto-fit, minmax(20px, 1fr)); 
        gap: 2px; /* Espacio más apretado */
    }

    .celda {
        padding: 4px 0; /* Botones más bajos */
        font-size: 0.6rem; /* Texto legible pero pequeño */
    }

    .celda.dorado { background: #eab308; color: black; font-weight: bold; box-shadow: 0 0 10px #eab308; }

    .grid-carton { display: grid; grid-template-columns: repeat(5, 1fr); gap: 10px; max-width: 500px; margin: 0 auto; }
    .bola-carton { background: #1e293b; padding: 15px; border-radius: 50%; border: 2px solid #334155; font-family: 'Orbitron'; }
    .bola-carton.marcada { background: #10b981; border-color: #34d399; transform: scale(1.1); transition: 0.3s; }

    .btn-next, .btn-jugar, .btn-plasma { 
        background: #10b981; border: none; color: white; padding: 15px 30px; 
        border-radius: 50px; font-family: 'Orbitron'; cursor: pointer; margin: 10px;
        box-shadow: 0 0 15px rgba(16, 185, 129, 0.4);
    }
    .btn-plasma { background: #8b5cf6; box-shadow: 0 0 15px rgba(139, 92, 246, 0.4); }

    .chart-circle { 
        width: 120px; height: 120px; border-radius: 50%; margin: 20px auto;
        background: conic-gradient(#10b981 calc(var(--porcentaje) * 1%), #334155 0);
        display: flex; align-items: center; justify-content: center; position: relative;
    }
    .chart-circle::after { content: ""; position: absolute; width: 90px; height: 90px; background: #0f172a; border-radius: 50%; }
    .chart-circle span { position: relative; z-index: 10; font-weight: bold; }

    .balls-history { display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; margin-top: 20px; }
    .ball { width: 35px; height: 35px; background: #3b82f6; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 0.8rem; font-weight: bold; }

    .efecto-limpieza { position: relative; overflow: hidden; }
    .efecto-limpieza::after { content: ""; position: absolute; top: 0; left: -150%; width: 100%; height: 100%; background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent); transition: left 0.8s; }
    .barrido::after { left: 150%; }
</style>