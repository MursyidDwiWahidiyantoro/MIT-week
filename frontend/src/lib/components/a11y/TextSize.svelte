<script lang="ts">
  import { textSize, type TextScale } from '$lib/a11y.svelte';

  const OPTIONS: TextScale[] = [100, 125, 150];
</script>

<div class="text-size" role="group" aria-label="Ukuran teks">
  <span class="label">Ukuran teks</span>
  {#each OPTIONS as option (option)}
    <button
      type="button"
      class:selected={textSize.value === option}
      aria-pressed={textSize.value === option}
      onclick={() => textSize.set(option)}
    >
      {option}%
    </button>
  {/each}
</div>

<style>
  .text-size { display: inline-flex; align-items: center; gap: var(--space-1); }
  .label { font-size: var(--text-xs); color: var(--text-3); margin-right: var(--space-1); }
  button {
    min-height: 36px; min-width: 44px; padding: 0 var(--space-2);
    background: transparent; color: var(--text-2);
    border: var(--border-width) solid var(--border); border-radius: var(--radius-sm);
    font: inherit; font-size: var(--text-xs); cursor: pointer;
  }
  button:hover { color: var(--text); background: var(--hover-overlay); }
  button.selected { background: var(--selected-bg); color: var(--text); border-color: var(--border-strong); font-weight: 600; }
  button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
