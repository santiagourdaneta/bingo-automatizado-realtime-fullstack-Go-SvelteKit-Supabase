<script>
	export let numeros = [];
	export let marcados = [];
</script>

<div class="grid">
	{#each numeros as num}
		<div class="celda" class:activa={marcados.includes(num)}>
			{num}
		</div>
	{/each}
</div>

<style>
	.grid {
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 8px;
		max-width: 350px;
		margin: 20px auto;
	}
	.celda {
		background: #1e293b;
		padding: 20px;
		text-align: center;
		border-radius: 8px;
		font-weight: bold;
		border: 2px solid #334155;
		transition: all 0.3s ease;
	}
	.activa {
		background: #f59e0b;
		color: #000;
		border-color: #fff;
		transform: scale(1.05);
		box-shadow: 0 0 15px rgba(245, 158, 11, 0.5);
	}
</style>