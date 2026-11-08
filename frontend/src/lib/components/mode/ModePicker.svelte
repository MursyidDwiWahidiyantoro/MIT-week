<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import { mode, type Mode } from '$lib/mode.svelte';

  const options = [
    { value: 'explore', title: 'mode.explore', detail: 'picker.exploreDetail' },
    { value: 'essential', title: 'mode.essential', detail: 'picker.essentialDetail' }
  ] as const satisfies ReadonlyArray<{ value: Mode; title: string; detail: string }>;

  /**
   * This is an inline, non-modal choice sitting in the page flow, not a dialog.
   * It must not trap Tab or grab focus on load: the hero is the first thing a
   * visitor should meet, and a trap would make the rest of the page unreachable
   * by keyboard. Dismissal is the explicit button below, which is in the tab
   * order, so no key handler is needed.
   */
</script>

<section class="picker card" aria-labelledby="mode-picker-title">
  <h2 id="mode-picker-title">{t('picker.title')}</h2>
  <p class="intro">{t('picker.intro')}</p>
  <div class="options">
    {#each options as option (option.value)}
      <button class="option lift" type="button" onclick={() => mode.set(option.value)}>
        <strong>{t(option.title)}</strong>
        <span>{t(option.detail)}</span>
      </button>
    {/each}
  </div>
  <button class="later btn btn--quiet btn--sm" type="button" onclick={() => mode.dismiss()}>
    {t('picker.later')}
  </button>
</section>

<style>
  .picker { padding: clamp(1.25rem, 3vw, 2rem); margin-top: var(--space-6); }
  @media (min-width: 60rem) {
    .picker { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1.5fr); column-gap: clamp(2rem, 4vw, 4rem); align-items: center; }
    .picker h2, .picker .intro { grid-column: 1; }
    .picker .intro { margin-bottom: 0; }
    .picker .options { grid-column: 2; grid-row: 1 / span 2; margin-bottom: 0; }
    .picker .later { grid-column: 1; justify-self: start; margin-top: var(--space-4); }
  }
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
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-hard-xs);
    font: inherit;
    cursor: pointer;
  }
  .option strong { font-family: var(--font-display); font-size: 1.1em; }
  .option span { color: var(--body-2); font-size: var(--text-sm); }
</style>
