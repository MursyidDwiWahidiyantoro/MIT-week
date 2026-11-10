<script lang="ts">
  import { onMount } from 'svelte';
  import { resolve } from '$app/paths';
  import { result } from '$lib/result.svelte';
  import { formatScore, platformName, formatClock } from '$lib/format';
  import StatusBadge from '$lib/components/ui/StatusBadge.svelte';
  import Panel from '$lib/components/ui/Panel.svelte';
  import EvidenceAccordion from '$lib/components/hasil/EvidenceAccordion.svelte';
  import HighlightedText from '$lib/components/hasil/HighlightedText.svelte';
  import ReadAloud from '$lib/components/a11y/ReadAloud.svelte';
  import ElaViewer from '$lib/components/hasil/ElaViewer.svelte';
  import LinkPreview from '$lib/components/hasil/LinkPreview.svelte';
  import VideoAnalysis from '$lib/components/hasil/VideoAnalysis.svelte';
  import RiskMap from '$lib/components/hasil/RiskMap.svelte';
  import { autoRead, speak, speechAvailable } from '$lib/a11y.svelte';
  import { t, type MessageKey } from '$lib/i18n/index.svelte';
  import { simpleSentence, explanation, basisList } from '$lib/i18n/verdict';
  import { keyReasons } from '$lib/reasons';
  import { inview } from '$lib/motion';
  import { sessionHistory } from '$lib/session-history.svelte';
  import tacticLabelsId from '$lib/data/tactics.json';
  import Icon from '$lib/components/ui/Icon.svelte';
  import type { DossierParsed } from '$lib/api/schemas';

  const dossier: DossierParsed | null = result.value;
  const imageFile: File | null = result.imageFile;
  let verdictHeading: HTMLElement | null = $state(null);

  // The session's list on the dashboard gains this check (memory only).
  onMount(() => {
    if (dossier) sessionHistory.add(dossier, imageFile?.name ?? null);
  });

  // Move focus to the verdict once it exists, so a screen reader lands on the
  // answer instead of staying on whatever was focused during the scan.
  $effect(() => {
    if (dossier && verdictHeading) verdictHeading.focus({ preventScroll: true });
  });

  const elaAnomalies = $derived(dossier && dossier.pillars.ela.available ? dossier.pillars.ela.anomalies : []);

  // The text tactic phrases were verified against: the link's caption, or
  // what the model read in the image.
  const highlight = $derived.by(() => {
    if (!dossier) return { text: '', label: '' };
    if (dossier.input_type === 'FILE') return { text: dossier.ocr_text, label: t('res.ocrLabel') };
    if (dossier.input_type === 'URL' && dossier.source_metadata) {
      const md = dossier.source_metadata;
      const caption = [md.title, md.description !== md.title ? md.description : ''].filter(Boolean).join('\n');
      return { text: caption, label: t('res.captionLabel') };
    }
    return { text: '', label: '' };
  });

  const sentence = $derived(dossier ? simpleSentence(dossier) : '');
  const explained = $derived(dossier ? explanation(dossier) : '');
  const basis = $derived(dossier ? basisList(dossier) : '');
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
  const video = $derived(dossier?.video?.status === 'OK' ? dossier.video : null);

  // The item as it was checked: the uploaded image or the video's first frame.
  // Nothing external is loaded (CSP: self, data, blob).
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

  // The title row names the item: the post's title, its site, or the upload.
  const itemName = $derived.by(() => {
    if (!dossier) return '';
    if (dossier.input_type === 'FILE') return imageFile?.name || t('res.item.image');
    const md = dossier.source_metadata;
    if (md?.title) return md.title;
    try {
      return new URL(dossier.source_reference).hostname.replace(/^www\./, '');
    } catch {
      return dossier.source_reference;
    }
  });
  const meta = $derived.by(() => {
    if (!dossier) return '';
    const parts: string[] = [];
    parts.push(t(dossier.input_type === 'FILE' ? 'app.type.image' : video ? 'app.type.video' : 'app.type.link'));
    if (video?.platform) parts.push(platformName(video.platform));
    else if (dossier.source_metadata?.site_name) parts.push(dossier.source_metadata.site_name);
    if (video?.duration_sec) parts.push(formatClock(video.duration_sec));
    try {
      parts.push(new Date(dossier.timestamp).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }));
    } catch {
      // A malformed timestamp only drops the time.
    }
    return parts.join(' · ');
  });

  // The KPI strip: figures as type, one tone of colour for what crossed a line.
  type Kpi = { label: MessageKey; value: string; unit?: string; over?: boolean };
  const kpis = $derived.by((): Kpi[] => {
    if (!dossier) return [];
    const p = dossier.pillars;
    const list: Kpi[] = [];
    const aiLine = video ? 55 : 80;
    list.push(p.ai.available ? { label: 'kpi.ai', value: String(p.ai.score), unit: '/100', over: p.ai.score >= aiLine } : { label: 'kpi.ai', value: '—' });
    if (video) list.push({ label: 'kpi.manip', value: String(video.manipulation.score), unit: '/100', over: video.manipulation.score >= 70 });
    else list.push(p.ela.available ? { label: 'kpi.ela', value: String(p.ela.score), unit: '/100', over: p.ela.score >= 70 } : { label: 'kpi.ela', value: '—' });
    list.push({
      label: 'kpi.fact',
      value: p.fact.available ? t(`risk.fact.${p.fact.status}` as MessageKey) : '—',
      over: p.fact.available && p.fact.status === 'CONTRADICTS'
    });
    list.push(p.link.available ? { label: 'kpi.link', value: String(p.link.score), unit: '/100', over: p.link.score >= 45 } : { label: 'kpi.link', value: '—' });
    return list;
  });

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
    <h1 class="title">{t('res.goneH1')}</h1>
    <p class="lede">{t('res.goneLede')}</p>
    <a class="btn btn--primary" href={resolve('/deteksi')}>{t('res.checkAgain')}</a>
  </section>
{:else}
  <div class="dash tone-{tone}" use:inview>
    <div class="title-row">
      <div class="titles">
        <h1 class="title">{itemName}</h1>
        <p class="meta">{meta}</p>
      </div>
      <div class="actions">
        <ReadAloud text={`${sentence} ${explained}`} />
        <a class="btn btn--primary btn--sm" href={resolve('/deteksi')}><Icon name="plus" size={18} />{t('shell.newCheck')}</a>
      </div>
    </div>

    {#if dossier.is_demo}<p class="notice notice--demo" role="note">{t('res.demoNotice')}</p>{/if}
    {#if dossier.degraded}<p class="notice" role="note">{t('res.degraded')}</p>{/if}

    <!-- KPI strip: the verdict and the score first, then one cell per check. -->
    <section class="kpis" aria-label={t('res.h1')}>
      <div class="kpi kpi--verdict">
        <p class="k-label">{t('kpi.verdict')}</p>
        <h2 id="verdict-title" class="verdict-heading" tabindex="-1" bind:this={verdictHeading} data-testid="verdict-badge">
          <StatusBadge status={dossier.verdict.status} />
        </h2>
      </div>
      <div class="kpi">
        <p class="k-label">{t('kpi.score')}</p>
        <p class="k-value k-score tabular" aria-label={dossier.verdict.score === null ? undefined : formatScore(dossier.verdict.score)}>
          {#if dossier.verdict.score === null}—{:else}{dossier.verdict.score}<small>/100</small>{/if}
        </p>
      </div>
      {#each kpis as k (k.label)}
        <div class="kpi" class:over={k.over}>
          <p class="k-label">{t(k.label)}</p>
          <p class="k-value tabular" class:k-word={!k.unit && k.value !== '—'}>{k.value}{#if k.unit}<small>{k.unit}</small>{/if}</p>
        </div>
      {/each}
    </section>

    <div class="grid">
      <div class="col-main">
        <section class="card lead rise" style:--i="0" aria-labelledby="lead-title">
          <h2 id="lead-title" class="k-label">{t('res.lead')}</h2>
          <p class="sentence" data-testid="verdict-sentence">{sentence}</p>
          {#if dossier.verdict.score !== null}
            <div class="meter" role="img" aria-label={formatScore(dossier.verdict.score)}>
              <span class="meter-fill" style:width="{dossier.verdict.score}%"></span>
            </div>
            <p class="score-cap">{t('res.scoreCap')}</p>
          {/if}
          <h3 class="why-title">{t('res.why')}</h3>
          <ul class="why">
            {#each reasons as reason (reason)}
              <li><span class="dot" aria-hidden="true"></span><span>{reason}</span></li>
            {/each}
          </ul>
          {#if speechAvailable()}
            <label class="auto-read">
              <input type="checkbox" checked={autoRead.value} onchange={(e) => autoRead.set((e.currentTarget as HTMLInputElement).checked)} />
              {t('res.autoRead')}
            </label>
          {/if}
        </section>

        <section class="card rise" style:--i="1">
          <RiskMap {dossier} />
        </section>
      </div>

      <aside class="col-side rise" style:--i="2">
        <figure class="card media">
          {#if picture}
            <img src={picture} alt={t('chat.imageAlt')} />
          {:else}
            <div class="media-empty"><Icon name={dossier.input_type === 'URL' ? 'link' : 'image'} size={32} /><span>{itemName}</span></div>
          {/if}
          <figcaption>{meta}</figcaption>
        </figure>
        {#if showLinkPreview && dossier.source_metadata}
          <div class="card" data-testid="source-metadata">
            <h3 class="k-label">{t('res.panelLink')}</h3>
            <LinkPreview metadata={dossier.source_metadata} />
          </div>
        {/if}
      </aside>
    </div>

    <h2 class="sec-title">{t('res.evidence')}</h2>
    <div class="evidence">
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
          {#if dossier.tactics.length > 0}
            <div class="block">
              <h3 class="label">{t('res.panelTactics')}</h3>
              <HighlightedText text={highlight.text} sourceLabel={highlight.label} tactics={dossier.tactics} labels={tacticLabelsId} />
            </div>
          {/if}
          <EvidenceAccordion pillars={dossier.pillars} inputType={dossier.input_type} fromVideo={dossier.video?.status === 'OK'} />
        </div>
      </Panel>
      <Panel title={t('res.panelLimits')} tone="soft">
        {#if showExplained}<p class="explained">{explained}</p>{/if}
        {#if basis}<p class="basis"><strong>{t('res.basis')}:</strong> {basis}.</p>{/if}
        <p class="note">{t('res.limitsNote')}</p>
      </Panel>
    </div>
  </div>
{/if}

<style>
  .state { display: grid; gap: var(--space-4); justify-items: start; max-width: 40rem; padding-block: var(--space-6); }
  .dash { display: grid; grid-template-columns: minmax(0, 1fr); gap: var(--space-5); max-width: 76rem; }
  .tone-safe { --tone: var(--ok); }
  .tone-warn { --tone: var(--warn); }
  .tone-danger { --tone: var(--err); }
  .tone-neutral { --tone: var(--ink); }

  .title-row { display: flex; flex-wrap: wrap; align-items: flex-start; justify-content: space-between; gap: var(--space-3) var(--space-5); }
  .titles { display: grid; gap: 4px; min-width: 0; flex: 1 1 24rem; }
  .title { font-size: var(--text-xl); font-weight: 500; overflow-wrap: anywhere; }
  .meta { margin: 0; color: var(--muted); font-size: var(--text-sm); }
  .actions { display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-2); }
  .notice { margin: 0; padding: var(--space-3) var(--space-4); font-size: var(--text-sm); font-weight: 500; }

  /* KPIs are type in one strip split by hairlines, not tiles. */
  .kpis {
    display: grid;
    grid-template-columns: minmax(12rem, 1.4fr) repeat(5, minmax(0, 1fr));
    border: 1px solid var(--border-soft);
    border-radius: 16px;
    overflow: hidden;
  }
  .kpi { display: grid; align-content: start; gap: 8px; padding: 16px 18px; border-left: 1px solid var(--border-soft); min-width: 0; }
  .kpi:first-child { border-left: 0; }
  .k-label { margin: 0; color: var(--muted); font-size: var(--text-sm); font-weight: 500; }
  .k-value { margin: 0; font-size: 1.5rem; font-weight: 400; color: var(--ink); line-height: 1.15; }
  .k-value small { margin-left: 2px; color: var(--muted); font-size: var(--text-sm); }
  .k-word { font-size: var(--text-base); font-weight: 500; }
  .k-score { color: var(--tone); font-size: 1.75rem; }
  .over .k-value { color: var(--err); }
  .verdict-heading { margin: 0; font-size: inherit; line-height: inherit; }
  .verdict-heading:focus { outline: none; }
  .verdict-heading:focus-visible { outline: var(--focus-ring); outline-offset: 4px; border-radius: 12px; }

  .grid { display: grid; grid-template-columns: minmax(0, 2fr) minmax(16rem, 1fr); gap: var(--space-4); align-items: start; }
  .col-main, .col-side { display: grid; gap: var(--space-4); min-width: 0; }
  .card { margin: 0; padding: var(--space-5); background: var(--card); border-radius: 16px; min-width: 0; }
  .lead { display: grid; gap: var(--space-3); }
  .sentence { margin: 0; max-width: 34ch; font-size: clamp(1.3rem, 1.05rem + 1vw, 1.75rem); font-weight: 400; line-height: 1.3; }
  .meter { --meter-color: var(--tone); }
  .meter-fill { display: block; height: 100%; transform-origin: left center; }
  .score-cap { margin: 0; color: var(--muted); font-size: var(--text-xs); }
  .why-title { margin-top: var(--space-2); font-size: var(--text-base); font-weight: 500; }
  .why { list-style: none; margin: 0; padding: 0; display: grid; gap: var(--space-2); }
  .why li { display: flex; gap: var(--space-3); align-items: baseline; color: var(--body-2); }
  .dot { flex: none; width: 6px; height: 6px; border-radius: 50%; background: var(--tone); transform: translateY(-2px); }
  .auto-read { display: inline-flex; align-items: center; gap: var(--space-2); min-height: 44px; font-size: var(--text-sm); color: var(--body-2); cursor: pointer; }
  .auto-read input { width: 20px; height: 20px; }

  .media { display: grid; gap: var(--space-3); padding: var(--space-3); }
  .media img { display: block; width: 100%; max-height: 22rem; object-fit: contain; border-radius: 12px; background: var(--sheet); }
  .media-empty { display: grid; place-items: center; gap: var(--space-2); min-height: 10rem; color: var(--muted); text-align: center; font-size: var(--text-sm); overflow-wrap: anywhere; padding: var(--space-4); }
  .media figcaption { color: var(--muted); font-size: var(--text-xs); padding: 0 4px 4px; }

  .sec-title { margin-top: var(--space-3); font-size: var(--text-base); font-weight: 500; }
  .evidence { display: grid; gap: var(--space-4); }
  .details { display: grid; gap: var(--space-5); }
  .block { display: grid; gap: var(--space-2); }
  .label { margin: 0; font-weight: 500; font-size: var(--text-sm); color: var(--body-2); }
  .explained { margin: 0 0 var(--space-4); max-width: var(--measure); color: var(--body-2); }
  .note { margin: var(--space-4) 0 0; color: var(--body-2); font-size: var(--text-sm); max-width: var(--measure); }
  .basis { margin: 0; color: var(--ink); }

  :global([data-motion='in']) .meter-fill { animation: fill 900ms var(--ease-out) 300ms both; }
  @keyframes fill { from { transform: scaleX(0); } }

  @media (max-width: 75rem) {
    .kpis { grid-template-columns: repeat(3, minmax(0, 1fr)); }
    .kpi:nth-child(4) { border-left: 0; }
    .kpi:nth-child(n + 4) { border-top: 1px solid var(--border-soft); }
  }
  @media (max-width: 56rem) {
    .grid { grid-template-columns: minmax(0, 1fr); }
    .col-side { order: -1; }
  }
  @media (max-width: 36rem) {
    .kpis { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .kpi:nth-child(odd) { border-left: 0; }
    .kpi:nth-child(4) { border-left: 1px solid var(--border-soft); }
    .kpi:nth-child(n + 3) { border-top: 1px solid var(--border-soft); }
    .kpi--verdict { grid-column: 1 / -1; }
    .kpi--verdict + .kpi { border-left: 0; }
    .card { padding: var(--space-4); }
  }
</style>
