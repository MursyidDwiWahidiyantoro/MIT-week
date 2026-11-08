<script lang="ts">
  import { resolve } from '$app/paths';
  import Icon from '$lib/components/ui/Icon.svelte';
  import type { SourceMetadata } from '$lib/api/types';

  interface Props {
    metadata: SourceMetadata;
  }

  let { metadata }: Props = $props();

  const PLATFORM: Record<string, string> = {
    youtube: 'YouTube',
    tiktok: 'TikTok',
    x: 'X (Twitter)',
    instagram: 'Instagram',
    facebook: 'Facebook',
    threads: 'Threads',
    web: 'Situs web'
  };
  const platform = $derived(PLATFORM[metadata.platform] ?? 'Situs web');
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
    <p class="note">Dibaca dari pratinjau publik tautan. Gambar dari situs lain tidak pernah dimuat.</p>
  </div>
{:else}
  <div class="blocked" role="note">
    <p class="blocked-title">
      <Icon name="alert" size={18} />
      {metadata.status === 'BLOCKED'
        ? `${platform} tidak mengizinkan pembacaan otomatis`
        : 'Isi halaman tautan tidak dapat dibaca'}
    </p>
    <p class="desc">
      {metadata.status === 'BLOCKED'
        ? 'Isi unggahannya belum ikut diperiksa. Unggah tangkapan layarnya supaya AMBA bisa membaca teks dan memeriksa gambarnya.'
        : 'Situsnya mungkin sedang tidak aktif atau bukan halaman web biasa. Tangkapan layar tetap bisa diperiksa.'}
    </p>
    <!-- A GET form, like the home hero: navigates to /deteksi?tab=file. -->
    <form action={resolve('/deteksi')} method="GET">
      <input type="hidden" name="tab" value="file" />
      <button class="btn btn--primary" type="submit">
        <Icon name="image" size={18} /> Unggah tangkapan layar
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
