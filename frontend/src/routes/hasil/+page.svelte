<script lang="ts">
  import { goto } from '$app/navigation';
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
  import PromptBar from '$lib/components/ui/PromptBar.svelte';
  import { autoRead, speak, speechAvailable } from '$lib/a11y.svelte';
  import { t } from '$lib/i18n/index.svelte';
  import { simpleSentence, explanation, basisList } from '$lib/i18n/verdict';
  import { keyReasons } from '$lib/reasons';
  import { handOver } from '$lib/handoff';
  import { inview } from '$lib/motion';
  import type { LinkProblem } from '$lib/link-input';
  import tacticLabelsId from '$lib/data/tactics.json';
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

  // The person's own item, shown as their message: the uploaded image or the
  // video's first frame. Nothing external is loaded (CSP: self, data, blob).
  let imageUrl = $state<string | null>(null);
  $effect(() => {
    if (!imageFile) return;
    const objectUrl = URL.createObjectURL(imageFile);
    imageUrl = objectUrl;
    return () => URL.revokeObjectURL(objectUrl);
  });
  const picture = $derived.by(() => {
    if (imageUrl && dossier?.input_type === 'FILE') return imageUrl;
    const frame = dossier?.video?.frames?.[0]?.thumb ?? '';
    return /^(data:image\/|blob:|\/)/.test(frame) ? frame : null;
  });
  // Without a picture, the message names the link: its title, site or host.
  const linkName = $derived.by(() => {
    if (dossier?.input_type !== 'URL') return '';
    const md = dossier.source_metadata;
    if (md?.title) return md.title;
    if (md?.site_name) return md.site_name;
    try {
      const raw = dossier.pillars.link.final_url ?? '';
      return raw ? new URL(raw).hostname.replace(/^www\./, '') : '';
    } catch {
      return '';
    }
  });

  // Optional auto-read: off by default, speaks once per result.
  let autoReadDone = false;
  $effect(() => {
    if (dossier && !autoReadDone && autoRead.value && speechAvailable()) {
      autoReadDone = true;
      speak(`${sentence} ${explained}`);
    }
  });

  // The prompt stays at the bottom: the next check starts from here.
  let next = $state('');
  let nextProblem = $state<LinkProblem | null>(null);
  function onlink(link: string): void {
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() is applied to the route; the query string is appended
    void goto(`${resolve('/deteksi')}?q=${encodeURIComponent(link)}`);
  }
  function onfile(file: File): void {
    handOver(file);
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() is applied to the route; the query string is appended
    void goto(`${resolve('/deteksi')}?tab=file&from=prompt`);
  }
</script>

<svelte:head><title>{t('res.title')}</title></svelte:head>

{#if !dossier}
  <section class="state">
    <h1 class="h1">{t('res.goneH1')}</h1>
    <p class="lede">{t('res.goneLede')}</p>
    <a class="btn btn--primary" href={resolve('/')}>{t('res.checkAgain')}</a>
  </section>
{:else}
  <div class="chat tone-{tone}" use:inview>
    <h1 class="title">{t('res.h1')}</h1>

    <!-- What was sent, as the person's message. -->
    <div class="turn turn--user rise" style:--i="0">
      <p class="sr-only">{t('chat.you')}</p>
      {#if picture}
        <img class="bubble-img" src={picture} alt={t('chat.imageAlt')} />
      {:else}
        <p class="bubble">
          <Icon name={dossier.input_type === 'URL' ? 'link' : 'image'} size={18} />
          <span>{linkName || t('chat.you')}</span>
        </p>
      {/if}
    </div>

    <!-- AMBA's answer. -->
    <div class="turn turn--amba">
      <span class="avatar rise" style:--i="1" aria-hidden="true"><Icon name="shield-check" size={18} /></span>
      <div class="answer">
        {#if dossier.is_demo}
          <p class="notice notice--demo rise" style:--i="1" role="note">{t('res.demoNotice')}</p>
        {/if}
        {#if dossier.degraded}
          <p class="notice rise" style:--i="1" role="note">{t('res.degraded')}</p>
        {/if}

        <section class="verdict rise" style:--i="2" aria-labelledby="verdict-title">
          <h2 id="verdict-title" class="verdict-heading" tabindex="-1" bind:this={verdictHeading} data-testid="verdict-badge">
            <StatusBadge status={dossier.verdict.status} score={dossier.verdict.score} />
          </h2>
          <p class="sentence" data-testid="verdict-sentence">{sentence}</p>
          {#if dossier.verdict.score !== null}
            <div class="score">
              <p class="score-num tabular" aria-hidden="true"><span>{dossier.verdict.score}</span><small>/100</small></p>
              <div class="meter-wrap">
                <div class="meter" role="img" aria-label={formatScore(dossier.verdict.score)}>
                  <span class="meter-fill" style:width="{dossier.verdict.score}%"></span>
                </div>
                <p class="score-cap" aria-hidden="true">{t('res.scoreCap')}</p>
              </div>
            </div>
          {/if}
        </section>

        <section class="why rise" style:--i="3" aria-labelledby="why-title">
          <h2 id="why-title" class="h3">{t('res.why')}</h2>
          <ul>
            {#each reasons as reason (reason)}
              <li><span class="dot" aria-hidden="true"></span><span>{reason}</span></li>
            {/each}
          </ul>
        </section>

        <div class="actions rise" style:--i="4" role="group" aria-label={t('res.actions')}>
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
          <a class="btn btn--tonal btn--sm" href={resolve('/')}><Icon name="plus" size={18} />{t('chat.again')}</a>
        </div>

        <h2 class="evidence-title rise" style:--i="5">{t('res.evidence')}</h2>

        <!-- Unread footage has nothing to show; the panel stays only for the platform's own AI label. -->
        {#if dossier.video && (dossier.video.status !== 'UNAVAILABLE' || dossier.video.platform_ai_label)}
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
                  labels={tacticLabelsId}
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
      </div>
    </div>

    <div class="dock">
      <PromptBar id="next-url" bind:value={next} bind:problem={nextProblem} {onlink} {onfile} />
    </div>
  </div>
{/if}

<style>
  .state { max-width: 46rem; padding-block: var(--space-6); display: grid; gap: var(--space-4); justify-items: start; }

  .chat { display: flex; flex-direction: column; gap: var(--space-5); }
  .tone-safe { --tone: var(--ok); }
  .tone-warn { --tone: var(--warn); }
  .tone-danger { --tone: var(--err); }
  .tone-neutral { --tone: var(--border-strong); }

  .title { font-size: var(--text-xl); font-weight: 400; color: var(--body-2); }

  /* ---- the person's message ---- */
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
    overflow-wrap: anywhere;
  }
  .bubble :global(svg) { flex: none; color: var(--indigo-text); }
  .bubble-img {
    display: block;
    max-width: min(60%, 16rem);
    max-height: 18rem;
    object-fit: cover;
    border-radius: 20px 20px 6px 20px;
  }

  /* ---- AMBA's answer ---- */
  .turn--amba { display: grid; grid-template-columns: 36px minmax(0, 1fr); gap: var(--space-3); align-items: start; }
  .avatar {
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: var(--grad);
    color: #ffffff;
  }
  .answer { display: grid; gap: var(--space-5); min-width: 0; padding-top: 2px; }
  .notice { margin: 0; padding: var(--space-3) var(--space-4); font-size: var(--text-sm); font-weight: 500; }

  .verdict { display: grid; gap: var(--space-4); }
  .verdict-heading { margin: 0; font-size: inherit; line-height: inherit; }
  .verdict-heading:focus { outline: none; }
  .verdict-heading:focus-visible { outline: var(--focus-ring); outline-offset: 6px; border-radius: 99px; }
  .sentence {
    margin: 0;
    max-width: 30ch;
    font-size: clamp(1.4rem, 1.1rem + 1.3vw, 2rem);
    font-weight: 400;
    line-height: 1.3;
    letter-spacing: -0.01em;
  }
  .score { display: flex; align-items: center; gap: var(--space-4); flex-wrap: wrap; }
  .score-num { margin: 0; display: flex; align-items: baseline; gap: 4px; line-height: 1; }
  .score-num span { font-size: clamp(2.4rem, 1.8rem + 2vw, 3.2rem); font-weight: 400; color: var(--tone); }
  .score-num small { font-size: var(--text-base); color: var(--body-2); }
  .meter-wrap { display: grid; gap: 6px; flex: 1 1 14rem; min-width: 0; }
  .meter { --meter-color: var(--tone); }
  .meter-fill { display: block; height: 100%; transform-origin: left center; }
  .score-cap { margin: 0; font-size: var(--text-sm); color: var(--body-2); }

  .why ul { list-style: none; margin: var(--space-3) 0 0; padding: 0; display: grid; gap: var(--space-3); }
  .why li { display: flex; gap: var(--space-3); align-items: baseline; }
  .dot { flex: none; width: 6px; height: 6px; border-radius: 50%; background: var(--tone); transform: translateY(-2px); }

  .actions { display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-2) var(--space-4); }
  .auto-read { display: inline-flex; align-items: center; gap: var(--space-2); min-height: 44px; font-size: var(--text-sm); color: var(--body-2); cursor: pointer; }
  .auto-read input { width: 20px; height: 20px; }

  .evidence-title { font-size: var(--text-lg); font-weight: 500; margin-top: var(--space-3); }
  .details { display: grid; gap: var(--space-5); }
  .block { display: grid; gap: var(--space-2); }
  .label { margin: 0; font-weight: 500; font-size: var(--text-sm); color: var(--body-2); }
  .explained { margin: 0 0 var(--space-4); max-width: var(--measure); color: var(--body-2); }
  .note { margin: var(--space-4) 0 0; color: var(--body-2); font-size: var(--text-sm); max-width: var(--measure); }
  .basis { margin: 0; color: var(--ink); }

  /* The score bar fills once the answer is in view. */
  :global([data-motion='in']) .meter-fill { animation: fill 900ms var(--ease-out) 350ms both; }
  @keyframes fill { from { transform: scaleX(0); } }

  .dock {
    position: sticky;
    bottom: 0;
    z-index: 5;
    margin-top: var(--space-4);
    padding: var(--space-3) 0 calc(var(--space-2) + env(safe-area-inset-bottom));
    background: linear-gradient(to top, var(--paper) 72%, transparent);
  }

  @media (max-width: 40rem) {
    .turn--amba { grid-template-columns: minmax(0, 1fr); }
    .avatar { width: 32px; height: 32px; }
  }
</style>
