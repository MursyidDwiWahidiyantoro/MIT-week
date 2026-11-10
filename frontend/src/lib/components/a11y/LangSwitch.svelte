<script lang="ts">
  import { locale, t, type Locale } from '$lib/i18n/index.svelte';

  const OPTIONS: Array<{ value: Locale; label: string; name: string }> = [
    { value: 'id', label: 'ID', name: 'Bahasa Indonesia' },
    { value: 'en', label: 'EN', name: 'English' }
  ];
</script>

<div class="lang" role="group" aria-label={t('lang.group')}>
  {#each OPTIONS as option (option.value)}
    <button
      type="button"
      lang={option.value}
      class:is-active={locale.value === option.value}
      aria-pressed={locale.value === option.value}
      onclick={() => locale.set(option.value)}
    >
      <!-- The visible code stays part of the accessible name (WCAG 2.5.3). -->
      {option.label}<span class="sr-only"> {option.name}</span>
    </button>
  {/each}
</div>

<style>
  /* Same glass track as the theme switch: a raised neutral segment marks the
   * active language; the accent stays on the primary action. */
  .lang {
    display: inline-flex;
    gap: 2px;
    padding: 3px;
    background: var(--glass-bg);
    border: 1px solid var(--glass-edge);
    border-radius: var(--radius-md);
    box-shadow: var(--glass-highlight);
  }
  button {
    display: grid;
    place-items: center;
    min-width: 44px;
    min-height: 40px;
    padding: 0 8px;
    background: transparent;
    color: var(--body-2);
    border: 0;
    border-radius: calc(var(--radius-md) - 4px);
    font-family: var(--font-sans);
    font-weight: 700;
    font-size: var(--text-sm);
    cursor: pointer;
    transition: background-color var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out);
  }
  button:hover { background: var(--hover-overlay); color: var(--ink); }
  button.is-active,
  button.is-active:hover { background: var(--white); color: var(--ink); box-shadow: 0 0 0 1px var(--glass-edge), var(--shadow-1); }
  button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
