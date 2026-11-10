<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { AuditSession } from '$lib/audit.svelte';
  import { isMockMode } from '$lib/api/config';
  import { result } from '$lib/result.svelte';
  import { readSharedItem } from '$lib/share-target';
  import { readLink, type LinkProblem } from '$lib/link-input';
  import UrlField from '$lib/components/ui/UrlField.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import ModeSelectorTabs, { type InputKind } from '$lib/components/deteksi/ModeSelectorTabs.svelte';
  import DropzoneUploader from '$lib/components/deteksi/DropzoneUploader.svelte';
  import ScanningRadar from '$lib/components/deteksi/ScanningRadar.svelte';
  import demoData from '$lib/data/demo-cases.json';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { t, type MessageKey } from '$lib/i18n/index.svelte';

  const session = new AuditSession();

  let kind = $state<InputKind>('url');
  let url = $state('');
  let file = $state<File | null>(null);
  let problem = $state<LinkProblem | null>(null);
  let fileError = $state<string | null>(null);
  let sharedNotice = $state<string | null>(null);

  // The home page hands a pasted link over as ?q= and the check starts at
  // once, because submitting that form was the explicit action. The "upload a
  // screenshot" call to action on a blocked link opens the image tab with
  // ?tab=file.
  onMount(() => {
    const tab = page.url.searchParams.get('tab');
    if (tab === 'file' || tab === 'url') kind = tab;
    const q = page.url.searchParams.get('q')?.trim();
    if (q && !page.url.searchParams.get('share')) {
      kind = 'url';
      url = q;
      void tick().then(() => submit());
    }
  });

  // A share-target redirect carries a one-time cache id. Read and delete it,
  // then start the audit immediately because sharing was an explicit action.
  onMount(() => {
    const id = page.url.searchParams.get('share');
    if (id) void openShared(id);
  });

  // The contract parser (Zod) is not needed to paint this page. Warm it once
  // the browser is idle, so opening an example still needs no network.
  onMount(() => {
    const warm = () => void import('$lib/api/schemas').catch(() => {});
    const w = window as Window & { requestIdleCallback?: (cb: () => void, opts?: { timeout: number }) => number };
    if (w.requestIdleCallback) w.requestIdleCallback(warm, { timeout: 3000 });
    else setTimeout(warm, 300);
  });

  async function openShared(id: string): Promise<void> {
    sharedNotice = t('det.sharedOpening');
    const item = await readSharedItem(id);
    const link = item ? readLink(item.url ?? item.text ?? '').url : null;
    if (!item || (!item.file && !link)) {
      sharedNotice = null;
      problem = item ? 'notUrl' : null;
      fileError = item ? null : t('det.sharedMissing');
      return;
    }
    if (item.file) {
      kind = 'file';
      file = item.file;
    } else {
      kind = 'url';
      url = link ?? '';
    }
    sharedNotice = null;
    await tick();
    await submit();
  }

  // Three bundled example results. They make zero API calls, and they are
  // filtered here from the generated demo file, which stays untouched.
  const EXAMPLES: Array<{ id: string; label: MessageKey }> = [
    { id: 'demo-tiktok', label: 'ex.tiktok' },
    { id: 'demo-gambar-ai', label: 'ex.ai' },
    { id: 'demo-berita-ig', label: 'ex.edit' }
  ];
  type DemoCase = { id: string; image?: string; dossier: unknown };
  const demos = demoData.cases as DemoCase[];

  async function openExample(id: string): Promise<void> {
    const demo = demos.find((d) => d.id === id);
    if (!demo) return;
    let image: File | null = null;
    if (demo.image) {
      try {
        const res = await fetch(demo.image);
        if (res.ok) image = new File([await res.blob()], 'contoh-simulasi.jpg', { type: 'image/jpeg' });
      } catch {
        // The result still opens; only the image viewer is skipped.
      }
    }
    try {
      const { parseDossier } = await import('$lib/api/schemas');
      result.set(parseDossier(demo.dossier), image);
      void goto(resolve('/hasil'));
    } catch {
      // A bundled example that fails the contract is a build-time bug, not a
      // user-facing failure. Keep the form usable.
    }
  }

  async function submit(event?: SubmitEvent): Promise<void> {
    event?.preventDefault();
    if (session.running) return;
    fileError = null;
    let link: string | null = null;
    if (kind === 'url') {
      const reading = readLink(url);
      problem = reading.problem;
      if (!reading.url) return;
      link = reading.url;
      url = link;
    } else if (!file) {
      fileError = t('det.errFile');
      return;
    }
    await session.start(kind, { url: link ?? undefined, file: file ?? undefined });
    if (session.verdict) {
      // goto() keeps the in-memory result store; a location.href reload would
      // wipe it and land on "Hasil tidak tersimpan".
      await goto(resolve('/hasil'));
    }
  }
</script>

<svelte:head>
  <title>{t('det.title')}</title>
  <meta name="description" content={t('det.description')} />
</svelte:head>

<section class="head">
  <h1 class="h1">{t('det.h1')}</h1>
  <p class="lede">{t('det.lede')}</p>
</section>

{#if isMockMode}
  <p class="notice" role="note">
    <strong>{t('det.previewTitle')}</strong> {t('det.previewBody')}
  </p>
{/if}

<form class="console card" aria-label={t('det.form')} onsubmit={submit} novalidate>
  <ModeSelectorTabs selected={kind} onselect={(k) => { kind = k; problem = null; fileError = null; }} />

  {#if sharedNotice}
    <p class="shared" aria-live="polite">{sharedNotice}</p>
  {/if}

  <div class="input-area">
    {#if kind === 'url'}
      <UrlField id="url" bind:value={url} bind:problem disabled={session.running} />
    {:else}
      <DropzoneUploader onselect={(f) => { file = f; fileError = null; }} />
      {#if fileError}
        <p class="file-error" role="alert"><Icon name="alert" size={16} /> {fileError}</p>
      {/if}
    {/if}
  </div>

  <div class="actions">
    <Button type="submit" size="lg" loading={session.running}>
      {session.running ? t('det.running') : t('det.start')}
      {#if !session.running}<Icon name="arrow-right" size={18} />{/if}
    </Button>
    {#if session.running}
      <Button variant="quiet" onclick={() => session.abort()}>{t('det.cancel')}</Button>
    {/if}
  </div>

  <p class="live" aria-live="polite">{session.announcement}</p>

  {#if !session.running}
    <div class="examples">
      <span class="examples-label">{t('det.demos')}</span>
      <ul>
        {#each EXAMPLES as ex (ex.id)}
          <li><button class="example" type="button" onclick={() => openExample(ex.id)}>{t(ex.label)}</button></li>
        {/each}
      </ul>
    </div>
  {/if}
</form>

{#if session.running || session.done}
  <ScanningRadar stages={session.stages} partials={session.partials} running={session.running} />
{/if}

{#if session.error}
  <div class="error-panel" role="alert">
    <p class="error-title"><Icon name="alert" size={20} /> {t('det.errTitle')}</p>
    <p>{session.error.message}</p>
    {#if session.error.retryAfter}
      <p class="hint">{t('det.wait', { n: session.error.retryAfter })}</p>
    {/if}
  </div>
{/if}

<style>
  .head { max-width: 40rem; margin-bottom: var(--space-5); }
  h1 { margin-bottom: var(--space-3); }
  .notice {
    margin: 0 0 var(--space-5);
    padding: var(--space-3) var(--space-4);
    background: var(--yellow);
    color: var(--on-fill);
    border-radius: var(--radius-md);
    font-size: var(--text-sm);
    max-width: 46rem;
  }
  .console { max-width: 46rem; padding: clamp(1.25rem, 3vw, 2rem); display: grid; gap: var(--space-5); }
  .input-area { display: grid; gap: var(--space-3); }
  .shared { margin: 0; color: var(--body-2); font-size: var(--text-sm); }
  .file-error { display: flex; gap: var(--space-2); align-items: flex-start; margin: 0; color: var(--err); font-weight: 700; font-size: var(--text-sm); }
  .actions { display: flex; gap: var(--space-3); flex-wrap: wrap; }
  .live { min-height: 1.4em; margin: calc(var(--space-3) * -1) 0 0; color: var(--body-2); font-size: var(--text-sm); }
  .examples {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2) var(--space-3);
    padding-top: var(--space-4);
    border-top: 1px solid var(--border);
  }
  .examples-label { font-size: var(--text-sm); color: var(--muted); font-weight: 700; }
  .examples ul { display: flex; flex-wrap: wrap; gap: var(--space-2); list-style: none; margin: 0; padding: 0; }
  .example {
    min-height: 44px;
    padding: 0 var(--space-4);
    background: transparent;
    color: var(--ink);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-pill);
    font: inherit;
    font-size: var(--text-sm);
    font-weight: 700;
    cursor: pointer;
  }
  .example:hover { background: var(--hover-overlay); }

  .error-panel {
    max-width: 46rem;
    border: 1px solid var(--err);
    border-radius: var(--radius-lg);
    background: var(--coral-tint);
    padding: var(--space-5);
    margin-top: var(--space-5);
  }
  .error-panel p { margin: 0; }
  .error-title { display: flex; align-items: center; gap: var(--space-2); font-family: var(--font-display); font-weight: 800; font-size: 1.15em; margin-bottom: var(--space-2) !important; color: var(--err); }
  .hint { color: var(--body-2); font-size: var(--text-sm); margin-top: var(--space-2) !important; }

  @media (max-width: 44rem) {
    .actions :global(.btn) { width: 100%; }
  }
</style>
