<script lang="ts">
  import Icon from '$lib/components/ui/Icon.svelte';

  export type InputKind = 'url' | 'file' | 'text';

  interface Props {
    selected: InputKind;
    onselect: (kind: InputKind) => void;
  }

  let { selected, onselect }: Props = $props();

  const tabs: Array<{ value: InputKind; label: string; icon: 'link' | 'image' | 'text' }> = [
    { value: 'url', label: 'Tautan', icon: 'link' },
    { value: 'file', label: 'Gambar', icon: 'image' },
    { value: 'text', label: 'Teks', icon: 'text' }
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
        e.preventDefault();
        const i = tabs.findIndex((t) => t.value === selected);
        const next =
          e.key === 'ArrowRight'
            ? (i + 1) % tabs.length
            : (i - 1 + tabs.length) % tabs.length;
        onselect(tabs[next].value);
        // Roving tabindex: focus must follow selection, or the keyboard user
        // stays on a tab that is no longer the active one.
        const list = (e.currentTarget as HTMLElement).parentElement;
        list?.querySelectorAll<HTMLElement>('[role="tab"]')[next]?.focus();
      }}
    >
      <Icon name={tab.icon} size={18} />
      {tab.label}
    </button>
  {/each}
</div>

<style>
  .tabs {
    display: inline-flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    padding: var(--space-1);
    background: var(--white);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-hard-xs);
  }
  button {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 48px;
    padding: 0 18px;
    background: var(--white);
    color: var(--ink);
    border: 0;
    border-radius: var(--radius-md);
    font: inherit;
    font-weight: 700;
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
