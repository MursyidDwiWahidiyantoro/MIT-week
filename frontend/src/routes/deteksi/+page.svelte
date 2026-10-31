<script lang="ts">
  import { page } from '$app/state';
  import { AuditSession } from '$lib/audit.svelte';
  import Field from '$lib/components/ui/Field.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import ModeSelectorTabs, { type InputKind } from '$lib/components/deteksi/ModeSelectorTabs.svelte';
  import DropzoneUploader from '$lib/components/deteksi/DropzoneUploader.svelte';
  import JurorPresetBar from '$lib/components/deteksi/JurorPresetBar.svelte';
  import ScanningRadar from '$lib/components/deteksi/ScanningRadar.svelte';

  const session = new AuditSession();

  let kind = $state<InputKind>('url');
  let url = $state('');
  let text = $state('');
  let file = $state<File | null>(null);
  let inputError = $state<string | null>(null);

  // The home hero hands a pasted link over as ?q=.
  $effect(() => {
    const q = page.url.searchParams.get('q');
    if (q && !url) url = q;
  });

  $effect(() => () => session.abort());

  const presets = [
    { id: 'hoax', label: 'Pesan berantai', value: 'Bantuan Rp 500.000 dibagikan hari ini saja. Sebarkan ke 10 grup!' },
    { id: 'phish', label: 'Tautan mencurigakan', value: 'Akun Anda akan diblokir. Verifikasi sekarang di tautan berikut.' },
    { id: 'netral', label: 'Pengumuman warga', value: 'Jadwal posyandu bulan ini: Sabtu, 8 pagi di balai warga.' }
  ];

  function validate(): string | null {
    if (kind === 'url') {
      const value = url.trim();
      if (!value) return 'Tempel tautan yang ingin diperiksa.';
      if (!/^https?:\/\//i.test(value)) return 'Tautan harus dimulai dengan http:// atau https://';
      if (value.length > 2048) return 'Tautan terlalu panjang.';
    }
    if (kind === 'text') {
      if (!text.trim()) return 'Tulis atau tempel pesan yang ingin diperiksa.';
      if (text.length > 5000) return 'Teks terlalu panjang. Maksimal 5000 karakter.';
    }
    if (kind === 'file' && !file) return 'Pilih gambar terlebih dahulu.';
    return null;
  }

  async function submit(): Promise<void> {
    inputError = validate();
    if (inputError) return;
    await session.start(kind, { url: url.trim(), text: text.trim(), file: file ?? undefined });
    if (session.verdict) {
      window.location.href = '/hasil';
    }
  }
</script>

<svelte:head><title>Periksa konten — AMBA</title></svelte:head>

<h1>Periksa konten</h1>
<p class="lede">Pilih jenis konten, lalu jalankan pemeriksaan. Hasilnya indikatif, bukan bukti pasti.</p>

<ModeSelectorTabs selected={kind} onselect={(k) => { kind = k; inputError = null; }} />

{#if kind === 'url'}
  <Field id="url" label="Tautan" type="url" bind:value={url} hint="Contoh: https://example.org/artikel" error={inputError ?? undefined} />
{:else if kind === 'text'}
  <div class="field">
    <label for="teks">Teks pesan</label>
    <textarea id="teks" rows="6" bind:value={text} aria-invalid={inputError ? 'true' : undefined}></textarea>
    {#if inputError}<p class="error" role="alert">{inputError}</p>{/if}
  </div>
{:else}
  <DropzoneUploader onselect={(f) => { file = f; inputError = null; }} />
  {#if inputError}<p class="error" role="alert">{inputError}</p>{/if}
{/if}

<JurorPresetBar {presets} onpick={(value) => { kind = 'text'; text = value; inputError = null; }} />

<div class="actions">
  <Button size="lg" onclick={submit} loading={session.running}>
    {session.running ? 'Memeriksa…' : 'Mulai periksa'}
  </Button>
  {#if session.running}
    <Button variant="quiet" onclick={() => session.abort()}>Batalkan</Button>
  {/if}
</div>

<p class="live" aria-live="polite">{session.announcement}</p>

{#if session.running || session.done}
  <ScanningRadar stages={session.stages} />
{/if}

{#if session.error}
  <div class="error-panel" role="alert">
    <p>{session.error.message}</p>
    {#if session.error.retryAfter}
      <p class="hint">Tunggu {session.error.retryAfter} detik sebelum mencoba lagi.</p>
    {/if}
  </div>
{/if}

<style>
  h1 { font-size: var(--text-3xl); margin: 0 0 var(--space-3); }
  .lede { color: var(--text-2); max-width: var(--measure); margin: 0 0 var(--space-6); }
  .field { display: grid; gap: var(--space-2); margin-top: var(--space-5); }
  label { font-weight: 600; font-size: var(--text-sm); }
  textarea {
    padding: var(--space-3); background: var(--surface-2); color: var(--text);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-md); font: inherit; resize: vertical;
  }
  textarea:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .error { margin: 0; color: var(--err); font-size: var(--text-sm); }
  .actions { display: flex; gap: var(--space-3); margin: var(--space-5) 0; flex-wrap: wrap; }
  .live { min-height: 1.5em; color: var(--text-2); font-size: var(--text-sm); margin: 0 0 var(--space-4); }
  .error-panel {
    border: var(--border-width) solid var(--err); border-radius: var(--radius-md);
    padding: var(--space-4); margin-top: var(--space-4);
  }
  .error-panel p { margin: 0; }
  .hint { color: var(--text-2); font-size: var(--text-sm); margin-top: var(--space-2) !important; }
</style>
