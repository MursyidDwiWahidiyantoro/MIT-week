<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import { resolve } from '$app/paths';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { platformName } from '$lib/format';
  import type { SourceMetadata } from '$lib/api/types';

  interface Props {
    metadata: SourceMetadata;
  }

  let { metadata }: Props = $props();

  const platform = $derived(platformName(metadata.platform) || t('lp.web'));
</script>

{#if metadata.status === 'OK'}
  <div class="preview">
    <p class="meta">
      <span class="platform">{platform}</span>
      {#if metadata.site_name && metadata.site_name !== platform}<span>· {metadata.site_name}</span>{/if}
      {#if metadata.author}<span>· {metadata.author}</span>{/if}
    </p>
    {#if metadata.title}<p class="title">{metadata.title}</p>{/if}
    {#if metadata.description && metadata.description !== metadata.title}
      <p class="desc">{metadata.description}</p>
    {/if}
    <p class="note">{t('lp.note')}</p>
  </div>
{:else}
  <div class="blocked" role="note">
    <p class="blocked-title">
      <Icon name="alert" size={18} />
      {metadata.status === 'BLOCKED' ? t('lp.blocked', { platform }) : t('lp.unavailable')}
    </p>
    <p class="desc">
      {metadata.status === 'BLOCKED' ? t('lp.blockedBody') : t('lp.unavailableBody')}
    </p>
    <!-- A GET form, like the home hero: navigates to /deteksi?tab=file. -->
    <form action={resolve('/deteksi')} method="GET">
      <input type="hidden" name="tab" value="file" />
      <button class="btn btn--primary" type="submit">
        <Icon name="image" size={18} /> {t('lp.upload')}
      </button>
    </form>
  </div>
{/if}

<style>
  .preview, .blocked { display: grid; gap: var(--space-2); }
  .meta { margin: 0; display: flex; flex-wrap: wrap; gap: 6px; font-size: var(--text-sm); color: var(--muted); }
  .platform { font-weight: 700; color: var(--ink); }
  .title { margin: 0; font-weight: 700; font-size: var(--text-lg); line-height: 1.35; overflow-wrap: anywhere; }
  .desc { margin: 0; color: var(--body-2); max-width: var(--measure); overflow-wrap: anywhere; white-space: pre-line; }
  .note { margin: var(--space-2) 0 0; font-size: var(--text-sm); color: var(--muted); }
  .blocked-title { margin: 0; display: flex; align-items: center; gap: var(--space-2); font-weight: 800; }
  .blocked form { justify-self: start; margin-top: var(--space-2); }
</style>
