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
  /* Mirrors TextSize: 44px targets, brand fill for the active option. */
  .lang {
    display: inline-flex;
    gap: 4px;
    padding: 5px;
    background: var(--white);
    border: var(--border-width) solid var(--border);
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
    font-size: 0.9em;
    cursor: pointer;
    transition: background-color var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out);
  }
  button:hover { background: var(--indigo-tint); }
  button.is-active,
  button.is-active:hover { background: var(--brand); color: var(--on-brand); }
  button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
