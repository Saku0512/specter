<script lang="ts">
	import { language, setLanguage, type Language } from '$lib/language';

	const options: { value: Language; label: string }[] = [
		{ value: 'ja', label: '日本語' },
		{ value: 'en', label: 'English' }
	];
</script>

<div class="language-toggle" aria-label="Language">
	{#each options as option}
		<button
			type="button"
			class:active={$language === option.value}
			aria-pressed={$language === option.value}
			onclick={() => setLanguage(option.value)}
		>
			{option.label}
		</button>
	{/each}
</div>

<style>
	.language-toggle {
		display: inline-grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		min-width: 9.4rem;
		padding: 0.2rem;
		border: 1px solid rgba(145, 184, 220, 0.18);
		border-radius: 999px;
		background: rgba(9, 17, 31, 0.38);
		backdrop-filter: blur(12px);
	}

	button {
		min-height: 2.2rem;
		padding: 0.45rem 0.72rem;
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: #c7d7ec;
		font: inherit;
		font-size: 0.82rem;
		font-weight: 700;
		cursor: pointer;
		transition:
			background 160ms ease,
			color 160ms ease;
	}

	button.active {
		background: linear-gradient(135deg, #8af1ff 0%, #b9ffd2 100%);
		color: #07111f;
	}
</style>
