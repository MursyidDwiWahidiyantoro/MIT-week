<script lang="ts">
  import { t, type MessageKey } from '$lib/i18n/index.svelte';
  import { platformName } from '$lib/format';
  import { readLink, type LinkProblem } from '$lib/link-input';
  import Icon from './Icon.svelte';

  interface Props {
    id: string;
    value?: string;
    /** Set by the form after a failed submit. Typing clears it. */
    problem?: LinkProblem | null;
    disabled?: boolean;
    autofocus?: boolean;
  }

  let {
    id,
    value = $bindable(''),
    problem = $bindable(null),
    disabled = false,
    autofocus = false
  }: Props = $props();

  let input: HTMLInputElement | undefined = $state();
  let pasteFailed = $state(false);

  // The reading is live, but an error only shows after a submit attempt, so a
  // person is not scolded for a half-typed link.
  const reading = $derived(readLink(value));
  const chip = $derived(
    reading.url
      ? reading.platform
        ? t('link.platform', { platform: platformName(reading.platform) })
        : t('link.web')
      : ''
  );
  const errorText = $derived(
    pasteFailed ? t('link.pasteFailed') : problem ? t(`link.err.${problem}` as MessageKey) : ''
  );

  $effect(() => {
    if (autofocus) input?.focus();
  });

  function take(text: string): void {
    const read = readLink(text);
    // Keep the extracted link only; the surrounding share text is noise.
    value = read.url ?? text.trim();
    problem = null;
    pasteFailed = false;
  }

  function onpaste(e: ClipboardEvent): void {
    const text = e.clipboardData?.getData('text');
    if (!text) return;
    e.preventDefault();
    take(text);
  }

  async function pasteButton(): Promise<void> {
    pasteFailed = false;
    try {
      const text = await navigator.clipboard.readText();
      if (!text.trim()) {
        problem = 'empty';
        return;
      }
      take(text);
      input?.focus();
    } catch {
      pasteFailed = true;
      input?.focus();
    }
  }

</script>

<div class="url-field">
  <label for={id}>{t('link.label')}</label>
  <div class="row-input" class:has-error={!!errorText}>
    <input
      {id}
      bind:this={input}
      bind:value
      class="input"
      type="text"
      inputmode="url"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      {disabled}
      placeholder={t('link.placeholder')}
      aria-invalid={errorText ? 'true' : undefined}
      aria-describedby="{id}-status"
      oninput={() => {
        problem = null;
        pasteFailed = false;
      }}
      {onpaste}
    />
    <button type="button" class="paste" {disabled} onclick={pasteButton}>
      <Icon name="copy" size={18} />
      <span>{t('link.paste')}</span>
    </button>
  </div>

  <div class="status" id="{id}-status" aria-live="polite">
    {#if errorText}
      <p class="msg msg--error"><Icon name="alert" size={16} /> <span>{errorText}</span></p>
    {:else if chip}
      <p class="msg msg--ok">
        <Icon name="check" size={16} />
        <span>{chip}{reading.extracted ? ` · ${t('link.extracted')}` : ''}</span>
      </p>
    {/if}
  </div>
</div>

<style>
  .url-field { display: grid; gap: var(--space-2); }
  label { font-weight: 700; font-size: var(--text-base); }
  .row-input { display: flex; gap: var(--space-2); align-items: stretch; }
  .row-input .input { flex: 1 1 auto; min-width: 0; min-height: 56px; }
  .row-input.has-error .input { border-color: var(--err); }
  .paste {
    flex: none;
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 56px;
    min-width: 44px;
    padding: 0 var(--space-4);
    background: var(--surface-2);
    color: var(--ink);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    font: inherit;
    font-weight: 700;
    font-size: var(--text-sm);
    cursor: pointer;
  }
  .paste:hover { background: var(--hover-overlay); }
  .paste:disabled { opacity: var(--disabled-opacity); cursor: not-allowed; }
  .status { min-height: 1.6em; }
  .msg { display: flex; align-items: flex-start; gap: var(--space-2); margin: 0; font-size: var(--text-sm); font-weight: 700; }
  .msg :global(svg) { flex: none; margin-top: 2px; }
  .msg--error { color: var(--err); }
  .msg--ok { color: var(--body-2); }
  @media (max-width: 30rem) {
    .paste span { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); }
    .paste { padding: 0; width: 56px; justify-content: center; }
  }
</style>
