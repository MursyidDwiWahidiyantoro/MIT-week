<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { AuditSession } from '$lib/audit.svelte';
  import { isMockMode } from '$lib/api/config';
  import { readSharedItem } from '$lib/share-target';
  import { readLink, type LinkProblem } from '$lib/link-input';
  import { takeHandedFile } from '$lib/handoff';
  import { EXAMPLES, openExample } from '$lib/demos';
  import { t } from '$lib/i18n/index.svelte';
  import PromptBar from '$lib/components/ui/PromptBar.svelte';
  import ScanningRadar from '$lib/components/deteksi/ScanningRadar.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  const session = new AuditSession();

  let url = $state('');
  let problem = $state<LinkProblem | null>(null);
  let sharedNotice = $state<string | null>(null);
  let notice = $state<string | null>(null);
  // What the person sent: shown as their bubble, and kept to retry.
  type Sent = { kind: 'url'; url: string } | { kind: 'file'; file: File; preview: string };
  let sent = $state<Sent | null>(null);

  // The home page hands a link over as ?q= and an image in memory
  // (?from=prompt); either way the check starts at once, because sending it
  // was the explicit action. ?tab=file is kept for older links: it opens the
  // empty page, where the + button picks an image.
  onMount(() => {
    if (page.url.searchParams.get('share')) return;
    const handed = page.url.searchParams.get('from') === 'prompt' ? takeHandedFile() : null;
    if (handed) {
      void runFile(handed);
      return;
    }
    const q = page.url.searchParams.get('q')?.trim();
    if (q) void tick().then(() => runLink(q));
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

  // The image preview lives as long as the bubble that shows it.
  $effect(() => {
    const current = sent;
    return () => {
      if (current?.kind === 'file') URL.revokeObjectURL(current.preview);
    };
  });

  async function openShared(id: string): Promise<void> {
    sharedNotice = t('det.sharedOpening');
    const item = await readSharedItem(id);
    sharedNotice = null;
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
    notice = null;
    url = '';
    sent = { kind: 'url', url: reading.url };
    await session.start('url', { url: reading.url });
    // goto() keeps the in-memory result store; a full reload would wipe it.
    if (session.verdict) await goto(resolve('/hasil'));
  }

  async function runFile(file: File): Promise<void> {
    if (session.running) return;
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
</script>

<svelte:head>
  <title>{t('det.title')}</title>
  <meta name="description" content={t('det.description')} />
</svelte:head>

<div class="chat">
  {#if !sent}
    <section class="intro">
      <h1 class="h1">{t('det.h1')}</h1>
      <p class="lede">{t('det.lede')}</p>
      {#if isMockMode}
        <p class="notice" role="note"><strong>{t('det.previewTitle')}</strong> {t('det.previewBody')}</p>
      {/if}
      {#if sharedNotice}<p class="info" aria-live="polite">{sharedNotice}</p>{/if}
      {#if notice}<p class="info info--error" role="alert"><Icon name="alert" size={16} /> {notice}</p>{/if}
      <div class="chips" role="group" aria-label={t('det.demos')}>
        <span class="chips-label">{t('det.demos')}</span>
        {#each EXAMPLES as ex (ex.id)}
          <button class="chip-btn" type="button" onclick={() => openExample(ex.id)}>
            <Icon name={ex.icon} size={18} />
            <span>{t(ex.label)}</span>
          </button>
        {/each}
      </div>
    </section>
  {:else}
    <h1 class="sr-only">{t('det.h1')}</h1>
    <div class="turn turn--user">
      <p class="sr-only">{t('chat.you')}</p>
      {#if sent.kind === 'url'}
        <p class="bubble"><Icon name="link" size={18} /><span class="link-text">{sent.url}</span></p>
      {:else}
        <img class="bubble-img" src={sent.preview} alt={t('chat.imageAlt')} />
      {/if}
    </div>

    <div class="turn turn--amba">
      <span class="avatar" class:thinking={session.running} aria-hidden="true">
        <Icon name="shield-check" size={18} />
      </span>
      <div class="answer">
        {#if session.running}
          <p class="status-line"><span class="shimmer">{t('chat.thinking')}</span></p>
        {/if}
        {#if session.error}
          <div class="error-panel" role="alert">
            <p class="error-title"><Icon name="alert" size={18} /> {t('det.errTitle')}</p>
            <p>{session.error.message}</p>
            {#if session.error.retryAfter}
              <p class="hint">{t('det.wait', { n: session.error.retryAfter })}</p>
            {/if}
            <button class="btn btn--tonal btn--sm" type="button" onclick={retry}>
              <Icon name="refresh" size={16} />{t('res.checkAgain')}
            </button>
          </div>
        {/if}
        <details class="steps" open>
          <summary>{t('chat.steps')}</summary>
          <ScanningRadar stages={session.stages} partials={session.partials} running={session.running} />
        </details>
        <p class="sr-only" aria-live="polite">{session.announcement}</p>
      </div>
    </div>
  {/if}

  <div class="dock">
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
  </div>
</div>

<style>
  .chat {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    min-height: calc(100dvh - 9rem);
  }
  .intro { display: grid; gap: var(--space-3); padding-top: clamp(0.5rem, 4vh, 2.5rem); }
  .notice { margin: var(--space-2) 0 0; padding: var(--space-3) var(--space-4); font-size: var(--text-sm); }
  .info { display: flex; gap: var(--space-2); align-items: flex-start; margin: 0; color: var(--body-2); font-size: var(--text-sm); }
  .info--error { color: var(--err); font-weight: 500; }
  .chips { display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-2); margin-top: var(--space-3); }
  .chips-label { width: 100%; color: var(--muted); font-size: var(--text-sm); }
  .chips :global(svg) { color: var(--indigo-text); }

  /* ---- the conversation ---- */
  .turn--user { display: flex; justify-content: flex-end; }
  .bubble {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    max-width: min(85%, 34rem);
    margin: 0;
    padding: 12px 18px;
    background: var(--card);
    border-radius: 24px 24px 6px 24px;
    font-size: var(--text-sm);
  }
  .bubble :global(svg) { flex: none; color: var(--indigo-text); }
  .link-text { min-width: 0; overflow-wrap: anywhere; }
  .bubble-img {
    display: block;
    max-width: min(60%, 16rem);
    max-height: 18rem;
    object-fit: cover;
    border-radius: 20px 20px 6px 20px;
  }

  .turn--amba { display: grid; grid-template-columns: 36px minmax(0, 1fr); gap: var(--space-3); align-items: start; }
  /* AMBA's mark; while it works, a gradient ring turns around it. */
  .avatar {
    position: relative;
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: var(--grad);
    color: #ffffff;
  }
  .avatar.thinking::before {
    content: '';
    position: absolute;
    inset: -4px;
    border-radius: 50%;
    background: conic-gradient(from 0deg, var(--grad-1), var(--grad-2), var(--grad-3), transparent 70%, var(--grad-1));
    -webkit-mask: radial-gradient(farthest-side, transparent calc(100% - 3px), #000 calc(100% - 2px));
    mask: radial-gradient(farthest-side, transparent calc(100% - 3px), #000 calc(100% - 2px));
    animation: turn 1.1s linear infinite;
  }
  @keyframes turn { to { transform: rotate(360deg); } }
  .answer { display: grid; gap: var(--space-3); min-width: 0; padding-top: 4px; }
  .status-line { margin: 0; font-size: var(--text-lg); }

  .steps {
    padding: var(--space-3) var(--space-4) var(--space-4);
    background: var(--card);
    border-radius: var(--radius-lg);
  }
  .steps summary {
    min-height: 40px;
    display: flex;
    align-items: center;
    color: var(--body-2);
    font-size: var(--text-sm);
    font-weight: 500;
    cursor: pointer;
  }
  .steps[open] summary { margin-bottom: var(--space-2); }

  .error-panel {
    display: grid;
    gap: var(--space-2);
    justify-items: start;
    padding: var(--space-4) var(--space-5);
    background: var(--coral-tint);
    border-radius: var(--radius-lg);
  }
  .error-panel p { margin: 0; }
  .error-title { display: flex; align-items: center; gap: var(--space-2); font-weight: 500; color: var(--err); }
  .hint { color: var(--body-2); font-size: var(--text-sm); }

  /* The prompt stays at the bottom, with the page fading out behind it. */
  .dock {
    position: sticky;
    bottom: 0;
    z-index: 5;
    margin-top: auto;
    padding: var(--space-3) 0 calc(var(--space-2) + env(safe-area-inset-bottom));
    background: linear-gradient(to top, var(--paper) 72%, transparent);
  }

  @media (prefers-reduced-motion: reduce) {
    .avatar.thinking::before { animation: none; }
  }
</style>
