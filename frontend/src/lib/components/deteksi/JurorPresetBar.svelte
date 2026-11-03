<script lang="ts">
  export type Preset = { id: string; label: string; value: string };

  interface Props {
    presets: Preset[];
    onpick: (value: string) => void;
  }

  let { presets, onpick }: Props = $props();
</script>

<div class="presets">
  <p class="label">Contoh cepat</p>
  <ul>
    {#each presets as preset (preset.id)}
      <li>
        <button type="button" onclick={() => onpick(preset.value)}>{preset.label}</button>
      </li>
    {/each}
  </ul>
</div>

<style>
  .label {
    margin: 0 0 var(--space-2);
    font-family: var(--font-mono);
    font-size: var(--text-xs);
    color: var(--text-2);
    letter-spacing: 0.06em;
  }
  ul { display: flex; flex-wrap: wrap; gap: var(--space-2); list-style: none; margin: 0; padding: 0; }
  button {
    min-height: 36px;
    padding: 0 var(--space-3);
    background: var(--surface-2);
    color: var(--text-2);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    font: inherit;
    font-size: var(--text-sm);
    cursor: pointer;
    transition:
      background-color var(--dur-fast) var(--ease-out),
      color var(--dur-fast) var(--ease-out),
      border-color var(--dur-fast) var(--ease-out),
      transform var(--dur-fast) var(--ease-out),
      box-shadow var(--dur-fast) var(--ease-out);
  }
  button:hover {
    color: var(--text);
    background: var(--white);
    border-color: var(--border-strong);
    transform: translate(-2px, -2px);
    box-shadow: var(--shadow-hard-xs);
  }
  button:active { transform: translate(1px, 1px); box-shadow: none; }
  button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
