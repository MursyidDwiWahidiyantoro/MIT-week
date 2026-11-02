<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { AuditSession } from '$lib/audit.svelte';
  import { parseDossier } from '$lib/api/schemas';
  import { isMockMode } from '$lib/api/config';
  import { result } from '$lib/result.svelte';
  import { firstHttpUrl, readSharedItem } from '$lib/share-target';
  import Field from '$lib/components/ui/Field.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import ModeSelectorTabs, { type InputKind } from '$lib/components/deteksi/ModeSelectorTabs.svelte';
  import DropzoneUploader from '$lib/components/deteksi/DropzoneUploader.svelte';
  import ScanningRadar from '$lib/components/deteksi/ScanningRadar.svelte';
  import demoData from '$lib/data/demo-cases.json';
  import { ClipboardPaste } from '@lucide/svelte';

  const session = new AuditSession();

  let kind = $state<InputKind>('url');
  let url = $state('');
  let text = $state('');
  let file = $state<File | null>(null);
  let inputError = $state<string | null>(null);
  let sharedNotice = $state<string | null>(null);

  // The home hero hands a pasted link over as ?q=.
  $effect(() => {
    const q = page.url.searchParams.get('q');
    if (q && !url) url = q;
  });

  // A share-target redirect carries a one-time cache id. Read and delete it,
  // then start the audit immediately because sharing was an explicit action.
  onMount(() => {
    const id = page.url.searchParams.get('share');
    if (id) void openShared(id);
  });

  async function openShared(id: string): Promise<void> {
    sharedNotice = 'Membuka konten yang dibagikan…';
    const item = await readSharedItem(id);
    if (!item) {
      sharedNotice = null;
      inputError = 'Konten yang dibagikan tidak ditemukan atau sudah kedaluwarsa. Tempel manual di bawah.';
      return;
    }

    if (item.file) {
      kind = 'file';
      file = item.file;
    } else if (item.url) {
      kind = 'url';
      url = item.url;
    } else if (item.text) {
      const found = firstHttpUrl(item.text);
      if (found) {
        kind = 'url';
        url = found;
      } else {
        kind = 'text';
        text = item.text;
      }
    }

    sharedNotice = null;
    await tick();
    await submit();
  }

  async function pasteClipboard(): Promise<void> {
    inputError = null;
    try {
      if (navigator.clipboard?.read) {
        const items = await navigator.clipboard.read();
        for (const item of items) {
          const type = item.types.find((t) => t.startsWith('image/'));
          if (type) {
            const blob = await item.getType(type);
            const extension = type === 'image/png' ? 'png' : type === 'image/webp' ? 'webp' : 'jpg';
            kind = 'file';
            file = new File([blob], `tempelan.${extension}`, { type });
            return;
          }
        }
      }

      const value = (await navigator.clipboard.readText()).trim();
      if (!value) {
        inputError = 'Clipboard kosong. Salin tautan atau teks terlebih dahulu.';
        return;
      }
      const found = firstHttpUrl(value);
      if (found) {
        kind = 'url';
        url = found;
      } else {
        kind = 'text';
        text = value;
      }
    } catch {
      inputError = 'Tidak bisa membaca clipboard. Tempel manual dengan Ctrl+V atau tekan lama lalu pilih Tempel.';
    }
  }

  // Bundled example dossiers. Loading one makes zero network calls (PRD F8).
  const demos = demoData.cases.map((c) => ({ id: c.id, label: c.caption, dossier: c.dossier }));

  function openDemo(dossier: unknown): void {
    try {
      result.set(parseDossier(dossier), null);
      void goto(resolve('/hasil'));
    } catch {
      // A bundled dossier that fails the contract is a build-time bug, not a
      // user-facing failure. Say nothing and keep the form usable.
    }
  }

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
      // goto() keeps the in-memory result store; a location.href reload would
      // wipe it and land on "Hasil tidak tersimpan".
      await goto(resolve('/hasil'));
    }
  }
</script>

<svelte:head><title>Periksa konten — AMBA</title></svelte:head>

<h1>Periksa konten</h1>
<p class="lede">Pilih jenis konten, lalu jalankan pemeriksaan. Hasilnya indikatif, bukan bukti pasti.</p>

{#if isMockMode}
  <p class="preview-note" role="note">
    Mode pratinjau aktif: pemeriksaan memakai aturan contoh di perangkat ini,
    bukan server AMBA. Hasilnya diberi tanda CONTOH.
  </p>
{/if}

<div class="input-tools">
  <ModeSelectorTabs selected={kind} onselect={(k) => { kind = k; inputError = null; }} />
  <Button variant="quiet" onclick={pasteClipboard}>
    <ClipboardPaste size={16} aria-hidden="true" /> Tempel dari clipboard
  </Button>
</div>

{#if sharedNotice}
  <p class="shared" aria-live="polite">{sharedNotice}</p>
{/if}

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

<div class="demos">
  <p class="demos-label">Coba contoh</p>
  <ul>
    {#each demos as demo (demo.id)}
      <li>
        <button type="button" onclick={() => openDemo(demo.dossier)}>{demo.label}</button>
      </li>
    {/each}
  </ul>
  <p class="demos-note">Contoh dimuat dari perangkat Anda, tanpa koneksi ke server.</p>
</div>

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
  <ScanningRadar stages={session.stages} partials={session.partials} />
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
  .preview-note {
    margin: 0 0 var(--space-4); padding: var(--space-3);
    border: var(--border-width) solid var(--warn); border-radius: var(--radius-md);
    color: var(--warn); font-size: var(--text-sm); max-width: var(--measure);
  }
  .input-tools { display: flex; align-items: center; gap: var(--space-3); flex-wrap: wrap; }
  .shared { margin: var(--space-3) 0 0; color: var(--text-2); font-size: var(--text-sm); }
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

  .demos { margin: 0 0 var(--space-5); }
  .demos-label { margin: 0 0 var(--space-2); color: var(--text-2); font-size: var(--text-sm); }
  .demos ul { display: flex; flex-wrap: wrap; gap: var(--space-2); list-style: none; margin: 0; padding: 0; }
  .demos button {
    min-height: 44px; padding: 0 var(--space-4);
    background: var(--surface-2); color: var(--text);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-md); font: inherit; font-size: var(--text-sm); cursor: pointer;
    text-align: left;
  }
  .demos button:hover { background: var(--hover-overlay); }
  .demos button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .demos-note { margin: var(--space-2) 0 0; color: var(--text-3); font-size: var(--text-xs); }
</style>
