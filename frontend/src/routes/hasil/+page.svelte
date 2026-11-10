<script lang="ts">
  import { resolve } from '$app/paths';
  import { result } from '$lib/result.svelte';
  import { statusLabel, formatScore } from '$lib/format';
  import { tokenFromHash, verifyToken, buildShareUrl, type VerifyResult } from '$lib/verify';
  import StatusBadge from '$lib/components/ui/StatusBadge.svelte';
  import Panel from '$lib/components/ui/Panel.svelte';
  import EvidenceAccordion from '$lib/components/hasil/EvidenceAccordion.svelte';
  import HighlightedText from '$lib/components/hasil/HighlightedText.svelte';
  import ReplyCard from '$lib/components/hasil/ReplyCard.svelte';
  import ProofCardExporter from '$lib/components/hasil/ProofCardExporter.svelte';
  import ShareButton from '$lib/components/share/ShareButton.svelte';
  import ReadAloud from '$lib/components/a11y/ReadAloud.svelte';
  import ElaViewer from '$lib/components/hasil/ElaViewer.svelte';
  import LinkPreview from '$lib/components/hasil/LinkPreview.svelte';
  import VideoAnalysis from '$lib/components/hasil/VideoAnalysis.svelte';
  import { autoRead, speak, speechAvailable } from '$lib/a11y.svelte';
  import { t, locale, type MessageKey } from '$lib/i18n/index.svelte';
  import { shortTitle, simpleSentence, explanation } from '$lib/i18n/verdict';
  import tacticLabelsId from '$lib/data/tactics.json';
  import tacticLabelsEn from '$lib/data/tactics.en.json';
  import Icon from '$lib/components/ui/Icon.svelte';
  import type { DossierParsed } from '$lib/api/schemas';

  let live = $state<DossierParsed | null>(result.value);
  let imageFile = $state<File | null>(result.imageFile);
  let shared = $state<VerifyResult | null>(null);
  let checking = $state(false);
  let verdictHeading: HTMLElement | null = $state(null);

  const shareOrigin = $derived.by(() =>
    typeof window === 'undefined' ? '' : window.location.origin
  );

  // A token in the fragment means this is a shared report, not a live session.
  $effect(() => {
    const token = tokenFromHash(window.location.hash);
    if (!token || live) return;
    const controller = new AbortController();
    checking = true;
    void verifyToken(token, controller.signal)
      .then((r) => (shared = r))
      .finally(() => (checking = false));
    return () => controller.abort();
  });

  // Move focus to the verdict once it exists, so a screen reader lands on the
  // answer instead of staying on whatever was focused during the scan.
  $effect(() => {
    if (live && verdictHeading) verdictHeading.focus({ preventScroll: true });
  });

  const basisLabel = (b: string): string => t(`basis.${b}` as MessageKey);

  const dossier = $derived(live);
  const shareUrl = $derived(
    live && shareOrigin ? buildShareUrl(shareOrigin, live.share_token) : ''
  );
  // A demo has no signed token, so its share and reply panels cannot work.
  const canShare = $derived(!!live && live.share_token.length > 0);
  const elaAnomalies = $derived(
    live && live.pillars.ela.available ? live.pillars.ela.anomalies : []
  );

  // The text tactic phrases were verified against, so the highlights land
  // where the user can see them: the link's caption, or what the model read
  // in the image.
  const highlight = $derived.by(() => {
    if (!live) return { text: '', label: '' };
    if (live.input_type === 'FILE') {
      return { text: live.ocr_text, label: t('res.ocrLabel') };
    }
    if (live.input_type === 'URL' && live.source_metadata) {
      const md = live.source_metadata;
      const caption = [md.title, md.description !== md.title ? md.description : '']
        .filter(Boolean)
        .join('\n');
      return { text: caption, label: t('res.captionLabel') };
    }
    return { text: '', label: '' };
  });

  const sentence = $derived(live ? simpleSentence(live) : '');
  const explained = $derived(live ? explanation(live) : '');
  const spoken = $derived(live ? `${sentence} ${explained}` : '');
  const sharedTitle = (p: { title?: string; status: string }) =>
    locale.value === 'id' && p.title ? p.title : statusLabel(p.status as never);

  // Optional auto-read (PRD F14): off by default, speaks once per result.
  let autoReadDone = false;
  $effect(() => {
    if (live && !autoReadDone && autoRead.value && speechAvailable()) {
      autoReadDone = true;
      speak(spoken);
    }
  });
</script>

<svelte:head><title>{t('res.title')}</title></svelte:head>

{#if checking}
  <section class="state">
    <p class="eyebrow">{t('res.verifyEyebrow')}</p>
    <h1 class="h1">{t('res.checkingH1')}</h1>
    <p class="lede" aria-live="polite">{t('res.checkingLede')}</p>
  </section>
{:else if shared && !shared.valid && shared.reason === 'invalid'}
  <section class="state">
    <p class="eyebrow">{t('res.verifyEyebrow')}</p>
    <h1 class="h1">{t('res.invalidH1')}</h1>
    <p class="lede">{t('res.invalidLede')}</p>
    <a class="btn btn--primary" href={resolve('/deteksi')}>{t('res.checkYourself')}</a>
  </section>
{:else if shared && !shared.valid && shared.reason === 'unreachable'}
  <section class="state">
    <p class="eyebrow">{t('res.verifyEyebrow')}</p>
    <h1 class="h1">{t('res.unreachableH1')}</h1>
    <p class="lede">{t('res.unreachableLede')}</p>
    <a class="btn btn--primary" href={resolve('/deteksi')}>{t('res.checkYourself')}</a>
  </section>
{:else if shared && !shared.valid}
  <section class="state">
    <p class="eyebrow">{t('res.verifyEyebrow')}</p>
    <h1 class="h1">{t('res.busyH1')}</h1>
    <p class="lede">
      {t('res.busyLede')}
      {#if shared.reason === 'unavailable' && shared.retryAfter}
        {t('res.retryIn', { n: shared.retryAfter })}
      {:else}
        {t('res.retryLater')}
      {/if}
    </p>
    <button class="btn btn--primary" type="button" onclick={() => window.location.reload()}>{t('res.retry')}</button>
  </section>
{:else if shared && shared.valid && shared.payload.preview}
  <section class="state">
    <p class="eyebrow">{t('res.verifyEyebrow')}</p>
    <h1 class="h1">{t('res.previewH1')}</h1>
    <p class="verified verified--preview" role="status">{t('res.previewBadge')}</p>
    <p class="sentence">{sharedTitle(shared.payload)}</p>
    <p class="score-line">{t('res.scoreLine', { score: formatScore(shared.payload.score) })}</p>
    <p class="note">{t('res.previewNote')}</p>
  </section>
{:else if shared && shared.valid}
  <section class="state">
    <p class="eyebrow">{t('res.verifyEyebrow')}</p>
    <h1 class="h1">{t('res.verifiedH1')}</h1>
    <p class="verified" role="status"><Icon name="shield-check" size={18} /> {t('res.verifiedBadge')}</p>
    <p class="sentence">{sharedTitle(shared.payload)}</p>
    <p class="score-line">{t('res.scoreLine', { score: formatScore(shared.payload.score) })}</p>
    <p class="note">{t('res.verifiedNote')}</p>
  </section>
{:else if !dossier}
  <section class="state">
    <p class="eyebrow">{t('res.resultEyebrow')}</p>
    <h1 class="h1">{t('res.goneH1')}</h1>
    <p class="lede">{t('res.goneLede')}</p>
    <a class="btn btn--primary" href={resolve('/deteksi')}>{t('res.checkAgain')}</a>
  </section>
{:else}
  <section class="result-head">
    <h1 class="h1">{t('res.h1')}</h1>
  </section>

  {#if dossier.is_demo}
    <p class="notice notice--demo" role="note">
      {t('res.demoNotice')}
    </p>
  {/if}

  {#if dossier.degraded}
    <p class="notice" role="note">
      {t('res.degraded')}
    </p>
  {/if}

  <section class="verdict card lift" aria-labelledby="verdict-title">
    <h2 id="verdict-title" class="verdict-heading" tabindex="-1" bind:this={verdictHeading} data-info-key="verdict.score">
      <StatusBadge status={dossier.verdict.status} score={dossier.verdict.score} />
    </h2>
    <p class="sentence" data-info-key="verdict.sentence">{sentence}</p>
    <p class="explanation" data-info-key="verdict.explanation">{explained}</p>
    <div class="read-aloud" data-info-key="readaloud">
      <ReadAloud text={spoken} />
      {#if speechAvailable()}
        <label class="auto-read">
          <input
            type="checkbox"
            checked={autoRead.value}
            onchange={(e) => autoRead.set((e.currentTarget as HTMLInputElement).checked)}
          />
          {t('res.autoRead')}
        </label>
      {/if}
    </div>
  </section>

  {#if dossier.input_type === 'URL' && dossier.source_metadata && !(dossier.source_metadata.status === 'BLOCKED' && dossier.video?.status === 'OK')}
    <Panel title={t('res.panelLink')}>
      <div data-info-key="source.metadata">
        <LinkPreview metadata={dossier.source_metadata} />
      </div>
    </Panel>
  {/if}

  {#if dossier.video}
    <VideoAnalysis video={dossier.video} ai={dossier.pillars.ai} />
  {/if}

  {#if imageFile && dossier.input_type === 'FILE'}
    <Panel title={t('res.panelImage')}>
      <ElaViewer file={imageFile} anomalies={elaAnomalies} />
    </Panel>
  {/if}

  {#if dossier.tactics.length > 0}
    <Panel title={t('res.panelTactics')}>
      <div data-info-key="tactics">
        <HighlightedText
          text={highlight.text}
          sourceLabel={highlight.label}
          tactics={dossier.tactics}
          labels={locale.value === 'en' ? tacticLabelsEn : tacticLabelsId}
        />
      </div>
    </Panel>
  {/if}

  <Panel title={t('res.panelDetails')}>
    <EvidenceAccordion
      pillars={dossier.pillars}
      inputType={dossier.input_type}
      fromVideo={dossier.video?.status === 'OK'}
    />
  </Panel>

  <Panel title={t('res.panelLimits')} tone="soft">
    <dl data-info-key="verdict.limits">
      <dt>{t('res.basis')}</dt>
      <dd>{dossier.verdict.basis.map(basisLabel).join(', ') || '—'}</dd>
      <dt>{t('res.notChecked')}</dt>
      <dd>
        {dossier.verdict.not_checked.map(basisLabel).join(', ') || t('res.allRan')}
      </dd>
    </dl>
    <p class="note">
      {t('res.limitsNote')}
    </p>
  </Panel>

  <Panel title={t('res.panelShare')}>
    <div data-info-key="share">
      {#if canShare}
        <ShareButton token={dossier.share_token} title={shortTitle(dossier)} />
      {:else}
        <p class="note" role="note">
          {t('res.demoNoShare')}
        </p>
      {/if}
    </div>
  </Panel>

  <Panel title={t('res.panelReply')}>
    <div data-info-key="reply">
      <ReplyCard token={dossier.share_token} status={dossier.verdict.status} demo={dossier.is_demo} />
    </div>
  </Panel>

  <Panel title={t('res.panelProof')}>
    <div data-info-key="proofcard">
      <ProofCardExporter {dossier} {shareUrl} />
    </div>
  </Panel>
{/if}

<style>
  .state { max-width: 46rem; padding-block: var(--space-6); }
  .state .lede { margin-bottom: var(--space-5); }
  .state :global(.btn) { margin-top: var(--space-2); }

  .result-head { margin-bottom: var(--space-4); }
  .result-head .h1 { margin-top: var(--space-1); }

  .notice {
    margin: 0 0 var(--space-4);
    padding: var(--space-3) var(--space-4);
    background: var(--yellow);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-hard-xs);
    font-weight: 700;
    font-size: var(--text-sm);
  }
  .notice--demo { background: var(--coral); }

  .verified {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    margin: var(--space-2) 0 var(--space-4);
    padding: 6px 16px;
    background: var(--safe);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-pill);
    box-shadow: var(--shadow-hard-xs);
    font-weight: 700;
    font-size: var(--text-sm);
  }
  .verified--preview { background: var(--yellow); }

  .verdict { padding: clamp(1.25rem, 3vw, 2rem); margin-bottom: var(--space-5); }
  .verdict-heading { margin: 0; font-size: inherit; line-height: inherit; }
  .verdict-heading:focus { outline: none; }
  .verdict-heading:focus-visible { outline: var(--focus-ring); outline-offset: 6px; border-radius: 99px; }
  .sentence {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: clamp(1.35rem, 3vw, 1.9rem);
    line-height: 1.2;
    letter-spacing: -0.02em;
    margin: var(--space-5) 0 var(--space-3);
    max-width: 28ch;
  }
  .explanation { color: var(--body-2); max-width: var(--measure); margin: 0; }
  .read-aloud { margin-top: var(--space-4); display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-3) var(--space-5); }
  .auto-read { display: inline-flex; align-items: center; gap: var(--space-2); min-height: 44px; font-size: var(--text-sm); color: var(--body-2); cursor: pointer; }
  .auto-read input { width: 20px; height: 20px; accent-color: var(--ink); }
  .score-line { margin: var(--space-2) 0 0; font-weight: 700; }
  .note { margin: var(--space-4) 0 0; color: var(--body-2); font-size: var(--text-sm); max-width: var(--measure); }

  dl { display: grid; grid-template-columns: max-content 1fr; gap: var(--space-2) var(--space-5); margin: 0; }
  dt { font-weight: 700; font-size: var(--text-sm); }
  dd { margin: 0; }
  @media (max-width: 40rem) {
    dl { grid-template-columns: 1fr; gap: var(--space-1); }
    dd { margin-bottom: var(--space-3); }
  }
</style>
