<script lang="ts">
  export type InputKind = 'url' | 'file' | 'text';

  interface Props {
    selected: InputKind;
    onselect: (kind: InputKind) => void;
  }

  let { selected, onselect }: Props = $props();

  const tabs: Array<{ value: InputKind; label: string }> = [
    { value: 'url', label: 'Tautan' },
    { value: 'file', label: 'Gambar' },
    { value: 'text', label: 'Teks' }
  ];
</script>

<div class="tabs" role="tablist" aria-label="Jenis konten">
  {#each tabs as tab (tab.value)}
    <button
      type="button"
      role="tab"
      aria-selected={selected === tab.value}
      tabindex={selected === tab.value ? 0 : -1}
      class:selected={selected === tab.value}
      onclick={() => onselect(tab.value)}
      onkeydown={(e) => {
        if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return;
        const i = tabs.findIndex((t) => t.value === selected);
        const next = e.key === 'ArrowRight' ? (i + 1) % tabs.length : (i - 1 + tabs.length) % tabs.length;
        onselect(tabs[next].value);
      }}
    >
      {tab.label}
    </button>
  {/each}
</div>

<style>
  .tabs {
    display: inline-flex;
    gap: var(--space-1);
    padding: var(--space-1);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
  }
  button {
    min-height: 40px;
    padding: 0 var(--space-4);
    background: transparent;
    color: var(--text-2);
    border: 0;
    border-radius: var(--radius-sm);
    font: inherit;
    cursor: pointer;
  }
  button:hover { color: var(--text); background: var(--hover-overlay); }
  button.selected { background: var(--selected-bg); color: var(--text); font-weight: 600; }
  button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
