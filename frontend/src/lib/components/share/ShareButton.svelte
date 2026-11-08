<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { buildShareUrl } from '$lib/verify';
  import { isMockToken } from '$lib/api/mock';

  interface Props {
    token: string;
    title: string;
    shareText?: string;
  }

  let { token, title, shareText = '' }: Props = $props();

  let copied = $state(false);
  let error = $state<string | null>(null);

  // Derived, not stored: the origin never changes during a page's life.
  const url = $derived.by(() =>
    typeof window === 'undefined' ? '' : buildShareUrl(window.location.origin, token)
  );
  // Preview tokens are generated in the browser and are not signed. Say so
  // instead of borrowing the real report's "digital signature" wording.
  const preview = $derived(isMockToken(token));

  async function copy(): Promise<void> {
    error = null;
    try {
      await navigator.clipboard.writeText(url);
      copied = true;
      setTimeout(() => (copied = false), 2500);
    } catch {
      // Clipboard can be blocked; the field below is the manual fallback.
      error = t('share.copyFailed');
    }
  }

  function whatsapp(): void {
    const text = `${shareText || t('share.text')}: ${title}\n${url}`;
    window.open(`https://wa.me/?text=${encodeURIComponent(text)}`, '_blank', 'noopener');
  }
</script>

<div class="share">
  <Button variant="quiet" onclick={copy}>
    <Icon name={copied ? 'check' : 'link'} size={18} />
    {copied ? t('share.copied') : t('share.copy')}
  </Button>
  <Button variant="quiet" onclick={whatsapp}>
    <Icon name="whatsapp" size={18} /> {t('share.whatsapp')}
  </Button>

  <label class="field">
    <span>{t('share.field')}</span>
    <input class="input" readonly value={url} onfocus={(e) => e.currentTarget.select()} />
  </label>

  {#if error}<p class="error-text" role="alert">{error}</p>{/if}
  {#if preview}
    <p class="note">
      {t('share.previewNote')}
    </p>
  {:else}
    <p class="note">
      {t('share.signedNote')}
    </p>
  {/if}
</div>

<style>
  .share { display: flex; flex-wrap: wrap; gap: var(--space-3); align-items: center; }
  .field { flex-basis: 100%; display: grid; gap: var(--space-2); }
  .field span { font-weight: 700; }
  .note { flex-basis: 100%; margin: 0; color: var(--muted); font-size: var(--text-sm); max-width: var(--measure); }
</style>
