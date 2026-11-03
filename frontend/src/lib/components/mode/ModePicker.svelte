<script lang="ts">
  import { mode, type Mode } from '$lib/mode.svelte';

  const options: Array<{ value: Mode; title: string; detail: string }> = [
    { value: 'explore', title: 'Jelajah', detail: 'Lebih banyak detail visual dan animasi.' },
    { value: 'essential', title: 'Esensial', detail: 'Lebih tenang, teks lebih besar, gerak minimal.' }
  ];

  /**
   * This is an inline, non-modal choice sitting in the page flow, not a dialog.
   * It must not trap Tab or grab focus on load: the hero is the first thing a
   * visitor should meet, and a trap would make the rest of the page unreachable
   * by keyboard. Dismissal is the explicit button below, which is in the tab
   * order, so no key handler is needed.
   */
</script>

<section class="picker card" aria-labelledby="mode-picker-title">
  <h2 id="mode-picker-title">Pilih tampilan yang paling nyaman</h2>
  <p class="intro">Keduanya menampilkan informasi dan fungsi yang sama.</p>
  <div class="options">
    {#each options as option (option.value)}
      <button class="option lift" type="button" onclick={() => mode.set(option.value)}>
        <strong>{option.title}</strong>
        <span>{option.detail}</span>
      </button>
    {/each}
  </div>
  <button class="later btn btn--quiet btn--sm" type="button" onclick={() => mode.dismiss()}>
    Pakai pengaturan saat ini
  </button>
</section>

<style>
  .picker { padding: clamp(1.25rem, 4vw, 1.75rem); max-width: 38rem; margin-top: var(--space-6); }
  h2 { font-size: var(--text-xl); margin: 0 0 var(--space-2); }
  .intro { margin: 0 0 var(--space-4); color: var(--body-2); }
  .options { display: grid; gap: var(--space-3); margin-bottom: var(--space-4); }
  @media (min-width: 40rem) {
    .options { grid-template-columns: 1fr 1fr; }
  }
  .option {
    display: grid;
    gap: var(--space-1);
    text-align: left;
    min-height: 44px;
    padding: var(--space-4);
    background: var(--indigo-tint);
    color: var(--ink);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-hard-xs);
    font: inherit;
    cursor: pointer;
  }
  .option strong { font-family: var(--font-display); font-size: 1.1em; }
  .option span { color: var(--body-2); font-size: var(--text-sm); }
</style>
