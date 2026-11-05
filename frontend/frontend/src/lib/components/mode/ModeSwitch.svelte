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
    gap: 4px;
    margin: 0;
    padding: 5px;
    background: var(--white);
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
    min-height: 44px;
    padding: 0 14px;
    background: var(--white);
    color: var(--ink);
    border: 0;
    border-radius: 11px;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: var(--text-sm);
    cursor: pointer;
    transition:
      background-color var(--dur-fast) var(--ease-out),
      color var(--dur-fast) var(--ease-out),
      transform var(--dur-fast) var(--ease-out);
  }
  button:hover { background: var(--indigo-tint); }
  button.selected:hover { background: var(--ink); color: var(--lime); }
  button:active { transform: scale(0.97); }
  button.selected { background: var(--ink); color: var(--lime); }
  button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
