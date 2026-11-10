<script lang="ts">
  import { resolve } from '$app/paths';
  import { result } from '$lib/result.svelte';
  import { formatScore } from '$lib/format';
  import StatusBadge from '$lib/components/ui/StatusBadge.svelte';
  import Panel from '$lib/components/ui/Panel.svelte';
  import EvidenceAccordion from '$lib/components/hasil/EvidenceAccordion.svelte';
  import HighlightedText from '$lib/components/hasil/HighlightedText.svelte';
  import ReadAloud from '$lib/components/a11y/ReadAloud.svelte';
  import ElaViewer from '$lib/components/hasil/ElaViewer.svelte';
  import LinkPreview from '$lib/components/hasil/LinkPreview.svelte';
  import VideoAnalysis from '$lib/components/hasil/VideoAnalysis.svelte';
  import { autoRead, speak, speechAvailable } from '$lib/a11y.svelte';
  import { t, locale } from '$lib/i18n/index.svelte';
  import { simpleSentence, explanation, basisList } from '$lib/i18n/verdict';
  import { keyReasons } from '$lib/reasons';
  import tacticLabelsId from '$lib/data/tactics.json';
  import tacticLabelsEn from '$lib/data/tactics.en.json';
  import Icon from '$lib/components/ui/Icon.svelte';
  import type { DossierParsed } from '$lib/api/schemas';

  const dossier: DossierParsed | null = result.value;
  const imageFile: File | null = result.imageFile;
  let verdictHeading: HTMLElement | null = $state(null);

  // Move focus to the verdict once it exists, so a screen reader lands on the
  // answer instead of staying on whatever was focused during the scan.
  $effect(() => {
    if (dossier && verdictHeading) verdictHeading.focus({ preventScroll: true });
  });

  const elaAnomalies = $derived(
    dossier && dossier.pillars.ela.available ? dossier.pillars.ela.anomalies : []
  );

  // The text tactic phrases were verified against, so the highlights land
  // where the user can see them: the link's caption, or what the model read
  // in the image.
  const highlight = $derived.by(() => {
    if (!dossier) return { text: '', label: '' };
    if (dossier.input_type === 'FILE') return { text: dossier.ocr_text, label: t('res.ocrLabel') };
    if (dossier.input_type === 'URL' && dossier.source_metadata) {
      const md = dossier.source_metadata;
      const caption = [md.title, md.description !== md.title ? md.description : '']
        .filter(Boolean)
        .join('\n');
      return { text: caption, label: t('res.captionLabel') };
    }
    return { text: '', label: '' };
  });

  const sentence = $derived(dossier ? simpleSentence(dossier) : '');
  const explained = $derived(dossier ? explanation(dossier) : '');
  const basis = $derived(dossier ? basisList(dossier) : '');
  // Only a neutral (old) INSUFFICIENT payload needs the server's own words.
  const showExplained = $derived(dossier?.verdict.status === 'INSUFFICIENT' && explained !== '');
  const reasons = $derived(dossier ? keyReasons(dossier) : []);
  const tone = $derived(
    dossier
      ? ({ SAFE: 'safe', SUSPICIOUS: 'warn', TAMPERED: 'danger', INSUFFICIENT: 'neutral' } as const)[dossier.verdict.status]
      : 'neutral'
  );
  const showLinkPreview = $derived(
    !!dossier &&
      dossier.input_type === 'URL' &&
      !!dossier.source_metadata &&
      !(dossier.source_metadata.status === 'BLOCKED' && dossier.video?.status === 'OK')
  );

  // Optional auto-read: off by default, speaks once per result.
  let autoReadDone = false;
  $effect(() => {
    if (dossier && !autoReadDone && autoRead.value && speechAvailable()) {
      autoReadDone = true;
      speak(`${sentence} ${explained}`);
    }
  });
</script>

<svelte:head><title>{t('res.title')}</title></svelte:head>

{#if !dossier}
  <section class="state">
    <h1 class="h1">{t('res.goneH1')}</h1>
    <p class="lede">{t('res.goneLede')}</p>
    <a class="btn btn--primary" href={resolve('/deteksi')}>{t('res.checkAgain')}</a>
  </section>
{:else}
  <div class="column">
  <section class="result-head">
    <h1 class="title">{t('res.h1')}</h1>
  </section>

  {#if dossier.is_demo}
    <p class="notice notice--demo" role="note">{t('res.demoNotice')}</p>
  {/if}
  {#if dossier.degraded}
    <p class="notice" role="note">{t('res.degraded')}</p>
  {/if}

  <section class="verdict tone-{tone}" aria-labelledby="verdict-title">
    <h2 id="verdict-title" class="verdict-heading" tabindex="-1" bind:this={verdictHeading} data-testid="verdict-badge">
      <StatusBadge status={dossier.verdict.status} score={dossier.verdict.score} />
    </h2>
    <div class="conclusion">
      <p class="sentence" data-testid="verdict-sentence">{sentence}</p>
      {#if dossier.verdict.score !== null}
        <div class="score">
          <p class="score-num" aria-hidden="true"><span>{dossier.verdict.score}</span><small>/100</small></p>
          <div class="meter" role="img" aria-label={formatScore(dossier.verdict.score)}>
            <span class="meter-fill" style:width="{dossier.verdict.score}%"></span>
          </div>
          <p class="score-cap" aria-hidden="true">{t('res.scoreCap')}</p>
        </div>
      {/if}
    </div>
    <div class="read-aloud">
      <ReadAloud text={`${sentence} ${explained}`} />
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

  <section class="why" aria-labelledby="why-title">
    <h2 id="why-title" class="h3">{t('res.why')}</h2>
    <ul>
      {#each reasons as reason (reason)}
        <li><span class="bullet" aria-hidden="true"><Icon name="chevron" size={14} /></span><span lang={locale.value === 'en' ? undefined : 'id'}>{reason}</span></li>
      {/each}
    </ul>
  </section>

  {#if dossier.video}
    <VideoAnalysis video={dossier.video} ai={dossier.pillars.ai} />
  {/if}

  {#if imageFile && dossier.input_type === 'FILE'}
    <Panel title={t('res.panelImage')}>
      <ElaViewer file={imageFile} anomalies={elaAnomalies} />
    </Panel>
  {/if}

  <Panel title={t('res.panelDetails')}>
    <div class="details">
      {#if showLinkPreview && dossier.source_metadata}
        <div class="block" data-testid="source-metadata">
          <h3 class="label">{t('res.panelLink')}</h3>
          <LinkPreview metadata={dossier.source_metadata} />
        </div>
      {/if}
      {#if dossier.tactics.length > 0}
        <div class="block">
          <h3 class="label">{t('res.panelTactics')}</h3>
          <HighlightedText
            text={highlight.text}
            sourceLabel={highlight.label}
            tactics={dossier.tactics}
            labels={locale.value === 'en' ? tacticLabelsEn : tacticLabelsId}
          />
        </div>
      {/if}
      <EvidenceAccordion
        pillars={dossier.pillars}
        inputType={dossier.input_type}
        fromVideo={dossier.video?.status === 'OK'}
      />
    </div>
  </Panel>

  <Panel title={t('res.panelLimits')} tone="soft">
    {#if showExplained}<p class="explained">{explained}</p>{/if}
    {#if basis}
      <p class="basis"><strong>{t('res.basis')}:</strong> {basis}.</p>
    {/if}
    <p class="note">{t('res.limitsNote')}</p>
  </Panel>

  <p class="again"><a class="btn btn--primary" href={resolve('/deteksi')}>{t('res.checkAgain')}</a></p>
  </div>
{/if}

<style>
  .column { max-width: 60rem; }
  .title { font-size: var(--text-2xl); line-height: 1.1; margin: 0; }
  .state { max-width: 46rem; padding-block: var(--space-6); }
  .state .lede { margin-bottom: var(--space-5); }

  .result-head { margin-bottom: var(--space-4); }
  .notice {
    margin: 0 0 var(--space-4);
    padding: var(--space-3) var(--space-4);
    background: var(--yellow);
    color: var(--on-fill);
    border-radius: var(--radius-md);
    font-weight: 700;
    font-size: var(--text-sm);
    max-width: 46rem;
  }
  .notice--demo { background: var(--coral); }

  /* One column on the reading measure: hero, reasons, evidence, limits. */

  .verdict {
    padding: clamp(1.25rem, 3vw, 2rem);
    margin-bottom: var(--space-5);
    border: 1px solid var(--border);
    border-left: 6px solid var(--tone, var(--border-strong));
    border-radius: var(--radius-lg);
    background: var(--tone-bg, var(--white));
  }
  .tone-safe { --tone: var(--ok); --tone-bg: var(--white); }
  .tone-warn { --tone: var(--yellow); --tone-bg: var(--yellow-tint); }
  .tone-danger { --tone: var(--coral); --tone-bg: var(--coral-tint); }
  .tone-neutral { --tone: var(--border-strong); --tone-bg: var(--white); }

  .verdict-heading { margin: 0; font-size: inherit; line-height: inherit; }
  .verdict-heading:focus { outline: none; }
  .verdict-heading:focus-visible { outline: var(--focus-ring); outline-offset: 6px; border-radius: 99px; }
  .conclusion { display: grid; gap: var(--space-5); align-items: center; margin: var(--space-5) 0 var(--space-4); }
  @media (min-width: 48rem) { .conclusion { grid-template-columns: minmax(0, 1fr) 14rem; gap: var(--space-7); } }
  .sentence {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: clamp(1.6rem, 1.2rem + 1.9vw, 2.4rem);
    line-height: 1.18;
    letter-spacing: -0.02em;
    margin: 0;
    max-width: 24ch;
  }
  .score { display: grid; gap: var(--space-2); }
  .score-num { margin: 0; font-family: var(--font-display); font-weight: 800; line-height: 1; display: flex; align-items: baseline; gap: 4px; }
  .score-num span { font-size: clamp(2.8rem, 2rem + 3vw, 4rem); letter-spacing: -0.03em; }
  .score-num small { font-size: var(--text-lg); color: var(--body-2); font-weight: 700; }
  .score-cap { margin: 0; font-size: var(--text-sm); color: var(--body-2); }
  .meter { height: 12px; border-radius: var(--radius-pill); background: var(--border); overflow: hidden; }
  .meter-fill { display: block; height: 100%; background: var(--tone, var(--brand)); border-radius: inherit; }
  .read-aloud { display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-3) var(--space-5); }
  .auto-read { display: inline-flex; align-items: center; gap: var(--space-2); min-height: 44px; font-size: var(--text-sm); color: var(--body-2); cursor: pointer; }
  .auto-read input { width: 20px; height: 20px; accent-color: var(--brand); }

  .why { margin-bottom: var(--space-5); }
  .why ul { list-style: none; margin: var(--space-3) 0 0; padding: 0; display: grid; gap: var(--space-3); }
  .why li { display: flex; gap: var(--space-3); align-items: flex-start; font-size: var(--text-base); }
  .bullet {
    flex: none;
    display: grid;
    place-items: center;
    width: 24px;
    height: 24px;
    margin-top: 2px;
    border-radius: 50%;
    background: var(--indigo-soft);
    color: var(--indigo-text);
    transform: rotate(-90deg);
  }

  .details { display: grid; gap: var(--space-5); }
  .block { display: grid; gap: var(--space-2); }
  .label {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: var(--text-sm);
    letter-spacing: 0.04em;
    color: var(--indigo-text);
  }

  .explained { margin: 0 0 var(--space-4); max-width: var(--measure); color: var(--body-2); }
  .note { margin: var(--space-4) 0 0; color: var(--body-2); font-size: var(--text-sm); max-width: var(--measure); }
  .basis { margin: 0; color: var(--ink); }
  .again { margin: var(--space-6) 0 0; }
  @media (max-width: 40rem) {
    .again .btn { width: 100%; }
  }
</style>
