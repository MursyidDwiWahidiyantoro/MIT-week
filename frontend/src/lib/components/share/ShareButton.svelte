<script lang="ts">
  import Button from '$lib/components/ui/Button.svelte';
  import { buildShareUrl } from '$lib/verify';
  import { isMockToken } from '$lib/api/mock';
  import { Link2, Check, MessageCircle } from '@lucide/svelte';

  interface Props {
    token: string;
    title: string;
    shareText?: string;
  }

  let { token, title, shareText = 'Hasil pemeriksaan AMBA' }: Props = $props();

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
      error = 'Tidak bisa menyalin otomatis. Salin tautan di bawah ini.';
    }
  }

  function whatsapp(): void {
    const text = `${shareText}: ${title}\n${url}`;
    window.open(`https://wa.me/?text=${encodeURIComponent(text)}`, '_blank', 'noopener');
  }
</script>

<div class="share">
  <Button variant="quiet" onclick={copy}>
    {#if copied}
      <Check size={16} aria-hidden="true" /> Tautan disalin
    {:else}
      <Link2 size={16} aria-hidden="true" /> Salin tautan
    {/if}
  </Button>
  <Button variant="quiet" onclick={whatsapp}>
    <MessageCircle size={16} aria-hidden="true" /> Bagikan ke WhatsApp
  </Button>

  <label class="field">
    <span>Tautan laporan</span>
    <input readonly value={url} onfocus={(e) => e.currentTarget.select()} />
  </label>

  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if preview}
    <p class="note">
      Ini tautan contoh dari mode pratinjau. Tautan tidak ditandatangani server
      dan hanya bisa dibuka sebagai laporan contoh.
    </p>
  {:else}
    <p class="note">
      Tautan memuat tanda tangan digital. Jika isinya diubah, laporan akan ditandai tidak sah.
    </p>
  {/if}
</div>

<style>
  .share { display: flex; flex-wrap: wrap; gap: var(--space-3); align-items: center; }
  .field { flex-basis: 100%; display: grid; gap: var(--space-1); }
  .field span { font-size: var(--text-sm); color: var(--text-2); }
  input {
    min-height: 44px; padding: 0 var(--space-3);
    background: var(--surface-2); color: var(--text);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-md); font-family: var(--font-mono); font-size: var(--text-xs);
  }
  input:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .error { flex-basis: 100%; margin: 0; color: var(--err); font-size: var(--text-sm); }
  .note { flex-basis: 100%; margin: 0; color: var(--text-3); font-size: var(--text-sm); max-width: var(--measure); }
</style>
