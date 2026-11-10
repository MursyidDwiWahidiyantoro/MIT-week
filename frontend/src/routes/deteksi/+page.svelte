<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { AuditSession } from '$lib/audit.svelte';
  import { isMockMode } from '$lib/api/config';
  import { readSharedItem } from '$lib/share-target';
  import { readLink, type LinkProblem } from '$lib/link-input';
  import { takeHandedFile, imageProblem } from '$lib/handoff';
  import { EXAMPLES, openExample } from '$lib/demos';
  import { sessionHistory } from '$lib/session-history.svelte';
  import { statusLabel } from '$lib/format';
  import { t, type MessageKey } from '$lib/i18n/index.svelte';
  import PromptBar from '$lib/components/ui/PromptBar.svelte';
  import ScanningRadar from '$lib/components/deteksi/ScanningRadar.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import type { Status } from '$lib/api/types';

  const session = new AuditSession();

  let tab = $state<'url' | 'file'>('url');
  let url = $state('');
  let problem = $state<LinkProblem | null>(null);
  let notice = $state<string | null>(null);
  let fileError = $state<string | null>(null);
  let dragging = $state(false);
  let picker: HTMLInputElement | undefined = $state();
  // What is being checked: shown in the input panel, kept to retry.
  type Sent = { kind: 'url'; url: string } | { kind: 'file'; file: File; preview: string };
  let sent = $state<Sent | null>(null);

  // A clock for the "n minutes ago" column.
  let now = $state(Date.now());
  onMount(() => {
    const id = setInterval(() => (now = Date.now()), 30_000);
    return () => clearInterval(id);
  });

  // A link arrives as ?q=, an image picked on another page in memory
  // (?from=prompt), a share as ?share=; each starts the check at once,
  // because sending it was the explicit action. ?tab=file opens the image tab.
  onMount(() => {
    if (page.url.searchParams.get('tab') === 'file') tab = 'file';
    if (page.url.searchParams.get('share')) return;
    const handed = page.url.searchParams.get('from') === 'prompt' ? takeHandedFile() : null;
    if (handed) {
      tab = 'file';
      void runFile(handed);
      return;
    }
    const q = page.url.searchParams.get('q')?.trim();
    if (q) void tick().then(() => runLink(q));
  });

  onMount(() => {
    const id = page.url.searchParams.get('share');
    if (id) void openShared(id);
  });

  // The contract parser (Zod) is not needed to paint this page; warm it idle.
  onMount(() => {
    const warm = () => void import('$lib/api/schemas').catch(() => {});
    const w = window as Window & { requestIdleCallback?: (cb: () => void, opts?: { timeout: number }) => number };
    if (w.requestIdleCallback) w.requestIdleCallback(warm, { timeout: 3000 });
    else setTimeout(warm, 300);
  });

  $effect(() => {
    const current = sent;
    return () => {
      if (current?.kind === 'file') URL.revokeObjectURL(current.preview);
    };
  });

  async function openShared(id: string): Promise<void> {
    notice = t('det.sharedOpening');
    const item = await readSharedItem(id);
    notice = null;
    const link = item ? readLink(item.url ?? item.text ?? '').url : null;
    if (!item || (!item.file && !link)) {
      notice = item ? t('link.err.notUrl') : t('det.sharedMissing');
      return;
    }
    if (item.file) await runFile(item.file);
    else if (link) await runLink(link);
  }

  async function runLink(raw: string): Promise<void> {
    if (session.running) return;
    const reading = readLink(raw);
    problem = reading.problem;
    if (!reading.url) return;
    tab = 'url';
    notice = null;
    url = reading.url;
    sent = { kind: 'url', url: reading.url };
    await session.start('url', { url: reading.url });
    if (session.verdict) await goto(resolve('/hasil'));
  }

  async function runFile(file: File): Promise<void> {
    if (session.running) return;
    const bad = imageProblem(file);
    if (bad) {
      fileError = t(bad === 'unsupported' ? 'drop.unsupported' : 'drop.tooLarge');
      return;
    }
    fileError = null;
    notice = null;
    sent = { kind: 'file', file, preview: URL.createObjectURL(file) };
    await session.start('file', { file });
    if (session.verdict) await goto(resolve('/hasil'));
  }

  function retry(): void {
    if (!sent) return;
    if (sent.kind === 'url') void runLink(sent.url);
    else void runFile(sent.file);
  }

  function pick(e: Event): void {
    const el = e.currentTarget as HTMLInputElement;
    const file = el.files?.[0];
    el.value = '';
    if (file) void runFile(file);
  }
  function drop(e: DragEvent): void {
    e.preventDefault();
    dragging = false;
    const file = e.dataTransfer?.files?.[0];
    if (file) void runFile(file);
  }

  function ago(at: number): string {
    const m = Math.floor((now - at) / 60000);
    if (m < 1) return t('app.time.just');
    if (m < 60) return t('app.time.min', { n: m });
    return t('app.time.hour', { n: Math.floor(m / 60) });
  }
  const tone = (s: Status): string => ({ SAFE: 'safe', SUSPICIOUS: 'warn', TAMPERED: 'danger', INSUFFICIENT: 'neutral' })[s];
</script>

<svelte:head>
  <title>{t('det.title')}</title>
  <meta name="description" content={t('det.description')} />
</svelte:head>

<div class="dash">
  <div class="title-row">
    <h1 class="title">{t('det.h1')}</h1>
    <p class="hint">{t('app.inputHint')}</p>
  </div>

  {#if isMockMode}
    <p class="notice" role="note"><strong>{t('det.previewTitle')}</strong> {t('det.previewBody')}</p>
  {/if}

  <section class="panel input" aria-label={t('det.form')}>
    <div class="tabs-line" role="tablist" aria-label={t('app.tabs.label')}>
      <button type="button" role="tab" aria-selected={tab === 'url'} onclick={() => (tab = 'url')} disabled={session.running}>
        <Icon name="link" size={18} />{t('app.tabs.link')}
      </button>
      <button type="button" role="tab" aria-selected={tab === 'file'} onclick={() => (tab = 'file')} disabled={session.running}>
        <Icon name="image" size={18} />{t('app.tabs.image')}
      </button>
    </div>

    {#if tab === 'url'}
      <PromptBar
        id="url"
        bind:value={url}
        bind:problem
        busy={session.running}
        autofocus={!sent}
        onlink={(link) => void runLink(link)}
        onfile={(file) => void runFile(file)}
        onstop={() => session.abort()}
      />
    {:else}
      <div
        class="drop"
        class:over={dragging}
        role="group"
        aria-label={t('app.drop')}
        ondragover={(e) => { e.preventDefault(); dragging = true; }}
        ondragleave={() => (dragging = false)}
        ondrop={drop}
      >
        <Icon name="image" size={28} />
        <p>{dragging ? t('app.dropping') : t('app.drop')}</p>
        <button class="btn btn--tonal btn--sm" type="button" disabled={session.running} onclick={() => picker?.click()}>{t('app.dropButton')}</button>
        <input bind:this={picker} class="sr-only" type="file" accept="image/jpeg,image/png,image/webp" tabindex="-1" aria-hidden="true" onchange={pick} />
      </div>
      {#if fileError}<p class="err" role="alert"><Icon name="alert" size={16} /> {fileError}</p>{/if}
    {/if}

    {#if notice}<p class="info" role="status">{notice}</p>{/if}

    {#if sent}
      <div class="current">
        <span class="cur-label">{t('app.checking')}</span>
        {#if sent.kind === 'url'}
          <span class="cur-item"><Icon name="link" size={16} /><span class="cur-text">{sent.url}</span></span>
        {:else}
          <span class="cur-item"><img src={sent.preview} alt={t('chat.imageAlt')} /><span class="cur-text">{sent.file.name}</span></span>
        {/if}
        {#if session.running}
          <button class="btn btn--sm" type="button" onclick={() => session.abort()}><Icon name="stop" size={16} />{t('det.cancel')}</button>
        {/if}
      </div>
    {/if}

    {#if session.error}
      <div class="error-panel" role="alert">
        <p class="error-title"><Icon name="alert" size={18} /> {t('det.errTitle')}</p>
        <p>{session.error.message}</p>
        {#if session.error.retryAfter}<p class="muted">{t('det.wait', { n: session.error.retryAfter })}</p>{/if}
        <button class="btn btn--tonal btn--sm" type="button" onclick={retry}><Icon name="refresh" size={16} />{t('res.checkAgain')}</button>
      </div>
    {/if}
    <p class="sr-only" aria-live="polite">{session.announcement}</p>
  </section>

  <section class="sec" aria-labelledby="mod-title">
    <div class="sec-head">
      <h2 id="mod-title">{t('app.modules')}</h2>
    </div>
    <ScanningRadar stages={session.stages} partials={session.partials} running={session.running} />
  </section>

  <section class="sec" aria-labelledby="hist-title">
    <div class="sec-head">
      <h2 id="hist-title">{t('app.history')}</h2>
      <p class="fresh">{t('app.history.note')}</p>
      {#if sessionHistory.list.length}
        <button class="btn btn--sm btn--quiet push" type="button" onclick={() => sessionHistory.clear()}>{t('app.history.clear')}</button>
      {/if}
    </div>
    {#if sessionHistory.list.length === 0}
      <p class="empty">{t('app.history.empty')}</p>
    {:else}
      <table class="tbl">
        <thead>
          <tr>
            <th scope="col">{t('app.history.item')}</th>
            <th scope="col" class="hide-s">{t('app.history.type')}</th>
            <th scope="col">{t('app.history.verdict')}</th>
            <th scope="col" class="num">{t('app.history.score')}</th>
            <th scope="col" class="hide-s">{t('app.history.time')}</th>
            <th scope="col"><span class="sr-only">{t('app.history.recheck')}</span></th>
          </tr>
        </thead>
        <tbody>
          {#each sessionHistory.list as e (e.id)}
            <tr>
              <td class="item">
                {#if e.demo}<span class="tag">{t('app.history.demo')}</span>{/if}
                <span class="item-label">{e.label}</span>
              </td>
              <td class="hide-s">{t(`app.type.${e.kind}` as MessageKey)}</td>
              <td><span class="vpill tone-{tone(e.status)}">{statusLabel(e.status)}</span></td>
              <td class="num tabular">{#if e.score === null}<span class="muted">—</span>{:else}{e.score}<small>/100</small>{/if}</td>
              <td class="hide-s time">{ago(e.at)}</td>
              <td class="act">
                {#if e.url}
                  <button class="btn btn--sm btn--quiet" type="button" disabled={session.running} onclick={() => runLink(e.url)}><Icon name="refresh" size={16} /><span class="hide-s">{t('app.history.recheck')}</span></button>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
    <div class="examples" role="group" aria-label={t('det.demos')}>
      <span class="ex-label">{t('det.demos')}</span>
      {#each EXAMPLES as ex (ex.id)}
        <button class="chip-btn" type="button" onclick={() => openExample(ex.id)}><Icon name={ex.icon} size={16} /><span>{t(ex.label)}</span></button>
      {/each}
    </div>
  </section>
</div>

<style>
  .dash { display: grid; grid-template-columns: minmax(0, 1fr); gap: var(--space-6); max-width: 76rem; }
  .title-row { display: flex; flex-wrap: wrap; align-items: baseline; justify-content: space-between; gap: 6px var(--space-5); }
  .title { font-size: var(--text-xl); font-weight: 500; }
  .hint { margin: 0; max-width: 60ch; color: var(--muted); font-size: var(--text-sm); }
  .notice { margin: 0; padding: var(--space-3) var(--space-4); font-size: var(--text-sm); }

  .panel { display: grid; grid-template-columns: minmax(0, 1fr); gap: var(--space-4); padding: var(--space-5); background: var(--card); border-radius: 16px; min-width: 0; }
  /* Underline tabs: a 2px indicator as wide as the label, on a hairline. */
  .tabs-line { display: flex; gap: var(--space-5); border-bottom: 1px solid var(--border-soft); }
  .tabs-line button {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 44px;
    margin-bottom: -1px;
    padding: 0 2px;
    border: 0;
    border-bottom: 2px solid transparent;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: var(--text-sm);
    font-weight: 500;
    cursor: pointer;
  }
  .tabs-line button[aria-selected='true'] { color: var(--ink); border-bottom-color: var(--indigo); }
  .tabs-line button:disabled { cursor: not-allowed; }

  .drop {
    display: grid;
    justify-items: center;
    gap: var(--space-3);
    padding: var(--space-6) var(--space-4);
    border: 1.5px dashed var(--border);
    border-radius: 16px;
    color: var(--muted);
    text-align: center;
    transition: border-color var(--dur-fast) var(--ease-out), background-color var(--dur-fast) var(--ease-out);
  }
  .drop.over { border-color: var(--indigo); background: var(--indigo-tint); color: var(--ink); }
  .drop p { margin: 0; color: var(--body-2); }
  .err { display: flex; gap: 6px; margin: 0; color: var(--err); font-size: var(--text-sm); font-weight: 500; }
  .info { margin: 0; color: var(--body-2); font-size: var(--text-sm); }

  .current {
    min-width: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2) var(--space-3);
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-soft);
  }
  .cur-label { color: var(--muted); font-size: var(--text-sm); }
  .cur-item { display: inline-flex; align-items: center; gap: 8px; min-width: 0; flex: 1 1 16rem; }
  .cur-item :global(svg) { flex: none; color: var(--indigo-text); }
  .cur-item img { width: 36px; height: 36px; object-fit: cover; border-radius: 8px; flex: none; }
  .cur-text { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: var(--text-sm); }

  .error-panel { display: grid; gap: var(--space-2); justify-items: start; padding: var(--space-4); background: var(--coral-tint); border-radius: 12px; }
  .error-panel p { margin: 0; }
  .error-title { display: flex; align-items: center; gap: 6px; color: var(--err); font-weight: 500; }

  .sec { display: grid; gap: var(--space-3); }
  .sec-head { display: flex; flex-wrap: wrap; align-items: baseline; gap: 4px var(--space-3); }
  .sec-head h2 { font-size: var(--text-base); font-weight: 500; }
  .fresh { margin: 0; color: var(--muted); font-size: var(--text-sm); }
  .push { margin-left: auto; }
  .empty { margin: 0; padding: var(--space-5); border: 1px dashed var(--border-soft); border-radius: 12px; color: var(--muted); font-size: var(--text-sm); }

  /* The session list: rules and alignment, no zebra, no header band. */
  .tbl { width: 100%; border-collapse: collapse; font-size: var(--text-sm); }
  .tbl th {
    padding: 0 12px 8px;
    color: var(--muted);
    font-weight: 400;
    font-size: var(--text-xs);
    text-align: left;
  }
  .tbl td { padding: 0 12px; height: 52px; border-top: 1px solid var(--border-soft); color: var(--body-2); }
  .tbl .num { text-align: right; }
  .tbl td.num { color: var(--ink); }
  .tbl small { color: var(--muted); margin-left: 1px; }
  .item { max-width: 0; width: 44%; }
  .item-label { display: inline-block; max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; vertical-align: middle; color: var(--ink); }
  .tag { display: inline-block; margin-right: 8px; padding: 1px 8px; border-radius: 999px; background: var(--surface-2); color: var(--body-2); font-size: var(--text-xs); vertical-align: middle; }
  .vpill { display: inline-block; padding: 3px 10px; border-radius: 999px; font-size: var(--text-xs); font-weight: 500; white-space: nowrap; }
  .tone-safe { background: var(--safe); color: var(--on-fill); }
  .tone-warn { background: var(--yellow); color: var(--on-fill); }
  .tone-danger { background: var(--coral); color: var(--on-fill); }
  .tone-neutral { background: var(--surface-2); color: var(--ink); }
  .time { white-space: nowrap; }
  .act { text-align: right; white-space: nowrap; }

  .examples { display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-2); margin-top: var(--space-3); }
  .ex-label { color: var(--muted); font-size: var(--text-sm); margin-right: 4px; }
  .examples :global(svg) { color: var(--indigo-text); }

  @media (max-width: 40rem) {
    .hide-s { display: none; }
    .item { width: auto; }
    .panel { padding: var(--space-4); }
  }
</style>
