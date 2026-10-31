<script lang="ts">
  import { mode, type Mode } from '$lib/mode.svelte';

  const options: Array<{ value: Mode; label: string }> = [
    { value: 'explore', label: 'Jelajah' },
    { value: 'essential', label: 'Esensial' }
  ];
</script>

<fieldset class="mode-switch">
  <legend>Tampilan</legend>
  {#each options as option (option.value)}
    <button
      type="button"
      class:selected={mode.resolved === option.value}
      aria-pressed={mode.resolved === option.value}
      onclick={() => mode.set(option.value)}
    >
      {option.label}
    </button>
  {/each}
</fieldset>

<style>
  .mode-switch {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    margin: 0;
    padding: var(--space-1);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
  }
  legend {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
  }
  button {
    min-height: 36px;
    padding: 0 var(--space-3);
    background: transparent;
    color: var(--text-2);
    border: 0;
    border-radius: var(--radius-sm);
    font: inherit;
    font-size: var(--text-sm);
    cursor: pointer;
  }
  button:hover { background: var(--hover-overlay); color: var(--text); }
  button.selected {
    background: var(--selected-bg);
    color: var(--text);
    font-weight: 600;
  }
  button:focus-visible {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }
</style>
