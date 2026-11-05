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
  import Icon from '$lib/components/ui/Icon.svelte';

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

<svelte:head>
  <title>Periksa konten — AMBA</title>
  <meta
    name="description"
    content="Tempel tautan, teks pesan, atau unggah tangkapan layar. AMBA menjelaskan tanda bahaya dengan kata sederhana."
  />
</svelte:head>

<section class="head">
  <p class="eyebrow">Pemeriksaan</p>
  <h1 class="h1">Periksa konten</h1>
  <p class="lede">
    Pilih jenis konten, lalu jalankan pemeriksaan. Hasilnya indikatif, bukan bukti pasti.
  </p>
</section>

{#if isMockMode}
  <p class="notice" role="note">
    <strong>Mode pratinjau.</strong> Pemeriksaan memakai aturan contoh di perangkat ini,
    bukan server AMBA. Hasilnya diberi tanda CONTOH.
  </p>
{/if}

<section class="console card" aria-label="Formulir pemeriksaan">
  <div class="console-top">
    <ModeSelectorTabs selected={kind} onselect={(k) => { kind = k; inputError = null; }} />
    <Button variant="quiet" onclick={pasteClipboard}>
      <Icon name="copy" size={18} /> Tempel dari clipboard
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
      <textarea
        class="textarea"
        id="teks"
        rows="6"
        bind:value={text}
        aria-invalid={inputError ? 'true' : undefined}
        placeholder="Tempel pesan yang Anda terima di sini."
      ></textarea>
      {#if inputError}<p class="error-text" role="alert">{inputError}</p>{/if}
    </div>
  {:else}
    <DropzoneUploader onselect={(f) => { file = f; inputError = null; }} />
    {#if inputError}<p class="error-text" role="alert">{inputError}</p>{/if}
  {/if}

  <div class="demos">
    <p class="demos-label">Coba contoh</p>
    <ul>
      {#each demos as demo (demo.id)}
        <li>
          <button class="demo-btn" type="button" onclick={() => openDemo(demo.dossier)}>
            {demo.label}
          </button>
        </li>
      {/each}
    </ul>
    <p class="demos-note">Contoh dimuat dari perangkat Anda, tanpa koneksi ke server.</p>
  </div>

  <div class="actions">
    <Button size="lg" onclick={submit} loading={session.running}>
      <Icon name="sparkle" size={18} />
      {session.running ? 'Memeriksa…' : 'Mulai periksa'}
    </Button>
    {#if session.running}
      <Button variant="quiet" onclick={() => session.abort()}>Batalkan</Button>
    {/if}
  </div>

  <p class="live" aria-live="polite">{session.announcement}</p>
</section>

{#if session.running || session.done}
  <ScanningRadar stages={session.stages} partials={session.partials} />
{/if}

{#if session.error}
  <div class="error-panel" role="alert">
    <p class="error-title">Pemeriksaan tidak selesai</p>
    <p>{session.error.message}</p>
    {#if session.error.retryAfter}
      <p class="hint">Tunggu {session.error.retryAfter} detik sebelum mencoba lagi.</p>
    {/if}
  </div>
{/if}

<style>
  .head { max-width: 54rem; margin-bottom: var(--space-6); }
  h1 { margin-bottom: var(--space-3); }
  .notice {
    margin: 0 0 var(--space-5);
    padding: var(--space-3) var(--space-4);
    background: var(--yellow);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-hard-xs);
    font-size: var(--text-sm);
    max-width: 46rem;
  }
  .console { padding: clamp(1.25rem, 3vw, 2rem); }
  .console-top {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }
  .console :global(.field) { margin-bottom: var(--space-5); }
  .console > :global(.field) { margin-bottom: var(--space-5); }
  .shared { margin: var(--space-3) 0 0; color: var(--body-2); font-size: var(--text-sm); }
  .field { display: grid; gap: var(--space-2); margin-bottom: var(--space-5); }
  label { font-weight: 700; }
  .actions { display: flex; gap: var(--space-3); margin: var(--space-5) 0 0; flex-wrap: wrap; }
  .live { min-height: 1.5em; color: var(--body-2); font-size: var(--text-sm); margin: var(--space-3) 0 0; }
  .error-panel {
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--coral-tint);
    box-shadow: var(--shadow-hard);
    padding: var(--space-5);
    margin-top: var(--space-5);
  }
  .error-panel p { margin: 0; }
  .error-title { font-family: var(--font-display); font-weight: 800; font-size: 1.15em; margin-bottom: var(--space-2) !important; }
  .hint { color: var(--body-2); font-size: var(--text-sm); margin-top: var(--space-2) !important; }

  .demos {
    margin-top: var(--space-5);
    padding-top: var(--space-4);
    border-top: var(--border-width) dashed var(--border);
  }
  .demos-label { margin: 0 0 var(--space-3); font-family: var(--font-display); font-weight: 800; font-size: var(--text-sm); text-transform: uppercase; letter-spacing: 0.06em; color: var(--indigo-text); }
  .demos ul { display: flex; flex-wrap: wrap; gap: var(--space-2); list-style: none; margin: 0; padding: 0; }
  .demo-btn {
    min-height: 44px;
    padding: 10px var(--space-4);
    background: var(--indigo-tint);
    color: var(--ink);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-hard-xs);
    font: inherit;
    font-size: var(--text-sm);
    font-weight: 700;
    cursor: pointer;
    text-align: left;
    transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
  }
  .demo-btn:hover { transform: translate(-2px, -2px); box-shadow: 5px 5px 0 var(--ink); }
  .demo-btn:active { transform: translate(2px, 2px); box-shadow: none; }
  .demo-btn:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .demos-note { margin: var(--space-3) 0 0; color: var(--muted); font-size: var(--text-sm); }

  .console + :global(.radar) { margin-top: var(--space-5); }

  /* Phone: the tab row gets the full width so three tabs share one line, and
   * the clipboard action drops below instead of squeezing them. */
  @media (max-width: 44rem) {
    .console-top { flex-direction: column; align-items: stretch; }
    .console-top :global(.tabs) { display: flex; width: 100%; }
    /* min-width:0 is required: flex-basis:0 alone still refuses to shrink
     * below the icon+label intrinsic width, which wrapped the third tab. */
    .console-top :global(.tabs button) { flex: 1 1 0; min-width: 0; padding: 0 8px; }
    .console-top :global(.btn) { width: 100%; }
    .actions { flex-direction: column; }
    .actions :global(.btn) { width: 100%; justify-content: center; }
  }
</style>
