<script>
    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher();

    let seleccionados = new Set();
    let derrotas = 0; 
    let mostrarMeteorito = false;

    // Función para manejar la selección
    function toggleNumero(num) {
        if (seleccionados.has(num)) {
            seleccionados.delete(num);
        } else if (seleccionados.size < 15) {
            seleccionados.add(num);
        }
        seleccionados = seleccionados; // Disparar reactividad
    }

    // Generador de Plasma (Aleatorio)
    function generadorPlasma() {
        let nuevos = new Set();
        while(nuevos.size < 15) {
            nuevos.add(Math.floor(Math.random() * 75) + 1);
        }
        seleccionados = nuevos;
    }

    function confirmarYJugar() {
        if (seleccionados.size === 15) {
            dispatch('jugar', { numeros: Array.from(seleccionados) });
        }
    }
</script>

<main class:triste={derrotas >= 3}>
    
    {#if mostrarMeteorito}
        <div class="meteorito" on:click={() => mostrarMeteorito = false}>
            ☄️ <span class="txt-premio">¡TOCA EL PREMIO!</span>
        </div>
    {/if}

    <h2>Selecciona tus 15 Números de la Suerte</h2>
    
    <div class="grid-75">
        {#each Array(75) as _, i}
            {@const num = i + 1}
            <button 
                class="celda" 
                class:dorado={seleccionados.has(num)}
                on:click={() => toggleNumero(num)}
            >
                {num}
            </button>
        {/each}
    </div>

    <div class="controles">
        <button class="btn-plasma" on:click={generadorPlasma}>GENERADOR DE PLASMA 🧪</button>
        <button 
            class="btn-jugar" 
            disabled={seleccionados.size < 15}
            on:click={confirmarYJugar}
        >
            ¡INICIAR DESPEGUE! 🚀 ({seleccionados.size}/15)
        </button>
    </div>
</main>

<style>
    /* Fondo que cambia con la tristeza */
    main {
        transition: background 2s ease;
        padding: 20px;
        min-height: 100vh;
    }
    
    main.triste {
        background: radial-gradient(circle, #0f172a, #1e293b);
    }

    .grid-75 {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
        gap: 8px;
        max-width: 600px;
        margin: 20px auto;
    }

    .celda {
        aspect-ratio: 1;
        background: #334155;
        border: 1px solid #475569;
        color: white;
        border-radius: 4px;
        cursor: pointer;
        font-weight: bold;
        transition: all 0.2s;
    }

    /* Efecto Dorado de Selección */
    .celda.dorado {
        background: #f59e0b;
        color: #000;
        filter: drop-shadow(0 0 10px gold);
        transform: scale(1.1);
        border: 2px solid white;
    }

    /* El Meteorito Animado */
    .meteorito {
        position: fixed;
        top: -100px;
        left: 50%;
        font-size: 4rem;
        cursor: pointer;
        z-index: 1000;
        animation: caida 4s forwards cubic-bezier(0.25, 0.46, 0.45, 0.94);
    }

    @keyframes caida {
        to { transform: translateY(110vh) rotate(360deg); }
    }

    .btn-plasma { background: #8b5cf6; margin-right: 10px; }
    .btn-jugar { background: #10b981; }
    .btn-jugar:disabled { background: #475569; opacity: 0.5; }

    .controles { margin-top: 30px; }
    
    button {
        padding: 12px 24px;
        border-radius: 8px;
        border: none;
        color: white;
        font-weight: bold;
        cursor: pointer;
    }
</style>