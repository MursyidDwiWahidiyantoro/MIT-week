<script lang="ts">
  import { textSize, type TextScale } from '$lib/a11y.svelte';

  const OPTIONS: Array<{ value: TextScale; label: string; scale: string }> = [
    { value: 100, label: 'A', scale: '0.9em' },
    { value: 125, label: 'A+', scale: '1.05em' },
    { value: 150, label: 'A++', scale: '1.25em' }
  ];
</script>

<div class="text-size">
  <span class="label" id="text-size-label">Ukuran teks</span>
  <div class="seg" role="group" aria-labelledby="text-size-label">
    {#each OPTIONS as option (option.value)}
      <button
        type="button"
        class:is-active={textSize.value === option.value}
        aria-pressed={textSize.value === option.value}
        aria-label={`Text size ${option.label}`}
        style:font-size={option.scale}
        onclick={() => textSize.set(option.value)}
      >
        {option.label}
      </button>
    {/each}
  </div>
</div>

<style>
  .text-size { display: inline-flex; align-items: center; gap: var(--space-2); }
  .label { font-size: 0.8em; font-weight: 700; color: var(--muted); }
  .seg {
    display: inline-flex;
    gap: 4px;
    padding: 5px;
    background: var(--white);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-md);
  }
  button {
    display: grid;
    place-items: center;
    min-width: 48px;
    min-height: 44px;
    padding: 0 8px;
    background: var(--white);
    color: var(--ink);
    border: 0;
    border-radius: 11px;
    font-family: var(--font-display);
    font-weight: 800;
    cursor: pointer;
  }
  button:hover { background: var(--indigo-tint); }
  button.is-active { background: var(--ink); color: var(--lime); }
  button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
