<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { t, type MessageKey } from '$lib/i18n/index.svelte';
  import type { LinkProblem } from '$lib/link-input';
  import { inview, lines } from '$lib/motion';
  import { handOver } from '$lib/handoff';
  import Icon from '$lib/components/ui/Icon.svelte';
  import PromptBar from '$lib/components/ui/PromptBar.svelte';
  import evalRun from '$lib/data/eval.json';

  let url = $state('');
  let problem = $state<LinkProblem | null>(null);

  // The quick check in the hero runs on the dashboard: a link as ?q=, an
  // image handed over in memory.
  function onlink(link: string): void {
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() is applied to the route; the query string is appended
    void goto(`${resolve('/deteksi')}?q=${encodeURIComponent(link)}`);
  }
  function onfile(file: File): void {
    handOver(file);
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() is applied to the route; the query string is appended
    void goto(`${resolve('/deteksi')}?tab=file&from=prompt`);
  }

  const SOURCES: Array<{ icon: 'search' | 'video' | 'image' | 'link' | 'text'; label: MessageKey }> = [
    { icon: 'search', label: 'land.src.tbh' },
    { icon: 'video', label: 'land.src.frames' },
    { icon: 'image', label: 'land.src.ela' },
    { icon: 'link', label: 'land.src.link' },
    { icon: 'text', label: 'land.src.web' }
  ];

  // The line under each module: its threshold in scoring.go, or for the
  // fact check (which has no score) where the debunks come from.
  const MODULES: Array<{ icon: 'sparkle' | 'image' | 'search' | 'link'; title: MessageKey; body: MessageKey; line: string }> = [
    { icon: 'sparkle', title: 'risk.ai', body: 'app.mod.ai.d', line: t('risk.threshold', { n: 'video 55 · gambar 80' }) },
    { icon: 'image', title: 'risk.ela', body: 'app.mod.ela.d', line: t('risk.threshold', { n: 70 }) },
    { icon: 'search', title: 'risk.fact', body: 'app.mod.fact.d', line: 'TurnBackHoax.id + situs cek fakta' },
    { icon: 'link', title: 'risk.link', body: 'app.mod.link.d', line: t('risk.threshold', { n: '45 · berbahaya 90' }) }
  ];

  const FEATURES: Array<{ icon: 'panel' | 'search' | 'speaker' | 'lock'; title: MessageKey; body: MessageKey }> = [
    { icon: 'panel', title: 'land.dash.f1.t', body: 'land.dash.f1.b' },
    { icon: 'search', title: 'land.dash.f2.t', body: 'land.dash.f2.b' },
    { icon: 'speaker', title: 'land.dash.f3.t', body: 'land.dash.f3.b' },
    { icon: 'lock', title: 'land.dash.f4.t', body: 'land.dash.f4.b' }
  ];

  // Measured on the labelled set (Scratch eval, 10 Nov 2026); see eval.json.
  const NUMBERS: Array<{ key: 'ai' | 'asli' | 'hoaks'; label: MessageKey }> = [
    { key: 'ai', label: 'land.num.ai' },
    { key: 'asli', label: 'land.num.real' },
    { key: 'hoaks', label: 'land.num.hoax' }
  ];
</script>

<svelte:head>
  <title>{t('home.title')}</title>
  <meta name="description" content={t('home.description')} />
</svelte:head>

<!-- ============ hero ============ -->
<section class="hero" use:inview>
  <div class="hero-inner">
    <p class="badge rise" style:--i="0"><span class="badge-dot" aria-hidden="true"></span>{t('land.badge')}</p>
    <h1 class="hero-title">
      {#each lines(t('land.h1')) as line, i (i)}
        <span class="rise" class:gradient-text={i === 1} style:--i={i + 1}>{line}</span>
      {/each}
    </h1>
    <p class="hero-lede rise" style:--i="3">{t('land.lede')}</p>
    <div class="hero-ctas rise" style:--i="4">
      <a class="btn btn--primary" href={resolve('/deteksi')}>{t('land.cta')}<Icon name="arrow-right" size={18} /></a>
      <a class="btn" href="#cara-kerja">{t('land.cta2')}</a>
    </div>
    <div class="hero-prompt rise" style:--i="5">
      <PromptBar id="hero-url" bind:value={url} bind:problem {onlink} {onfile} />
    </div>
  </div>
  <!-- The glowing wave under the hero, in the product's gradient. -->
  <svg class="wave" viewBox="0 0 1440 420" preserveAspectRatio="none" aria-hidden="true" focusable="false">
    <defs>
      <linearGradient id="wg" x1="0" x2="1" y1="0" y2="0">
        <stop offset="0" stop-color="#4285f4" stop-opacity="0" />
        <stop offset="0.25" stop-color="#4285f4" />
        <stop offset="0.55" stop-color="#9b72cb" />
        <stop offset="0.8" stop-color="#d96570" />
        <stop offset="1" stop-color="#d96570" stop-opacity="0" />
      </linearGradient>
    </defs>
    <g class="wave-glow">
      <path d="M0 300 C 240 200, 420 380, 720 260 S 1180 120, 1440 240" />
      <path d="M0 330 C 260 250, 460 400, 760 300 S 1200 180, 1440 290" />
    </g>
    <g class="wave-line">
      <path d="M0 300 C 240 200, 420 380, 720 260 S 1180 120, 1440 240" />
      <path d="M0 330 C 260 250, 460 400, 760 300 S 1200 180, 1440 290" />
      <path d="M0 360 C 280 300, 500 420, 800 340 S 1220 240, 1440 330" />
    </g>
  </svg>
  <div class="sources" aria-label={t('land.sources')}>
    <span class="sources-label">{t('land.sources')}</span>
    <ul>
      {#each SOURCES as s (s.label)}
        <li><Icon name={s.icon} size={18} />{t(s.label)}</li>
      {/each}
    </ul>
  </div>
</section>

<!-- ============ how it works ============ -->
<section class="sec" id="cara-kerja" aria-labelledby="how-title" use:inview>
  <div class="wrap wrap--wide">
    <header class="sec-head rise">
      <h2 id="how-title">{t('land.how.title')}</h2>
      <p>{t('land.how.lede')}</p>
    </header>
    <ol class="steps">
      <li class="step rise" style:--i="1">
        <div class="vig" aria-hidden="true">
          <div class="v-prompt"><span class="v-plus">+</span><span class="v-url">https://vt.tiktok.com/…</span><span class="v-send"><Icon name="arrow-up" size={14} /></span></div>
          <div class="v-chip"><Icon name="check" size={12} /> TikTok</div>
        </div>
        <p class="step-n tabular">01</p>
        <h3>{t('land.how.1.t')}</h3>
        <p>{t('land.how.1.b')}</p>
      </li>
      <li class="step rise" style:--i="2">
        <div class="vig" aria-hidden="true">
          <div class="v-scan">
            <img src="/edukasi/quiz-diedit.jpg" alt="" width="72" height="90" loading="lazy" />
            <span class="br tl"></span><span class="br tr"></span><span class="br bl"></span><span class="br br2"></span>
            <span class="v-beam"></span>
          </div>
          <ul class="v-mods">
            <li class="on">Jejak AI</li><li class="on">Teks ditimpa</li><li class="run">Cek fakta</li><li>Tautan</li>
          </ul>
        </div>
        <p class="step-n tabular">02</p>
        <h3>{t('land.how.2.t')}</h3>
        <p>{t('land.how.2.b')}</p>
      </li>
      <li class="step rise" style:--i="3">
        <div class="vig" aria-hidden="true">
          <div class="v-verdict">
            <span class="v-pill">Terindikasi hoaks</span>
            <span class="v-score tabular">85<small>/100</small></span>
            <span class="v-meter"><span></span></span>
            <span class="v-line"></span><span class="v-line short"></span>
          </div>
        </div>
        <p class="step-n tabular">03</p>
        <h3>{t('land.how.3.t')}</h3>
        <p>{t('land.how.3.b')}</p>
      </li>
    </ol>
  </div>
</section>

<!-- ============ modules ============ -->
<section class="sec" id="modul" aria-labelledby="mod-title" use:inview>
  <div class="wrap wrap--wide mod-grid">
    <div class="mod-lead rise">
      <h2 id="mod-title">{t('land.mod.title')}</h2>
      <p>{t('land.mod.lede')}</p>
      <div class="mod-card" aria-hidden="true">
        <span class="mod-ring"><Icon name="shield-check" size={40} /></span>
        <span class="mod-count tabular">4</span>
        <span class="mod-caption">{t('app.modules')}</span>
      </div>
    </div>
    <ul class="mod-list">
      {#each MODULES as m, i (m.title)}
        <li class="mod rise" style:--i={i + 1}>
          <span class="mod-ic"><Icon name={m.icon} size={22} /></span>
          <h3>{t(m.title)}</h3>
          <p>{t(m.body)}</p>
          <p class="mod-line">{m.line}</p>
        </li>
      {/each}
    </ul>
  </div>
</section>

<!-- ============ numbers, measured ============ -->
{#if evalRun.total > 0}
  <section class="sec numbers" aria-labelledby="num-title" use:inview>
    <div class="wrap wrap--wide">
      <header class="sec-head center rise">
        <h2 id="num-title">{t('land.num.title')}</h2>
        <p>{t('land.num.lede', { n: evalRun.total })}</p>
      </header>
      <div class="beam-wrap">
        <span class="beam" aria-hidden="true"></span>
        <ul class="nums">
          {#each NUMBERS as n, i (n.key)}
            <li class="num rise" style:--i={i + 1}>
              <p class="num-value tabular">{evalRun[n.key].ok}<small>/{evalRun[n.key].n}</small></p>
              <p class="num-label">{t(n.label)}</p>
            </li>
          {/each}
        </ul>
      </div>
      <p class="caveat rise" style:--i="4">{t('land.num.caveat')}</p>
    </div>
  </section>
{/if}

<!-- ============ the dashboard, as it is ============ -->
<section class="sec" aria-labelledby="dash-title" use:inview>
  <div class="wrap wrap--wide">
    <header class="sec-head center rise">
      <h2 id="dash-title">{t('land.dash.title')}</h2>
      <p>{t('land.dash.lede')}</p>
    </header>
    <figure class="shot rise" style:--i="1">
      <img src="/landing/dasbor.jpg" alt={t('land.dash.alt')} width="1440" height="900" loading="lazy" decoding="async" />
    </figure>
    <ul class="feats">
      {#each FEATURES as f, i (f.title)}
        <li class="feat rise" style:--i={i + 2}>
          <span class="feat-ic"><Icon name={f.icon} size={20} /></span>
          <h3>{t(f.title)}</h3>
          <p>{t(f.body)}</p>
        </li>
      {/each}
    </ul>
  </div>
</section>

<!-- ============ closing band ============ -->
<section class="band-sec" use:inview>
  <div class="wrap wrap--wide">
    <div class="band rise">
      <div>
        <h2>{t('land.band.title')}</h2>
        <p>{t('land.band.lede')}</p>
      </div>
      <a class="btn btn--primary" href={resolve('/deteksi')}>{t('land.band.cta')}<Icon name="arrow-right" size={18} /></a>
    </div>
  </div>
</section>

<style>
  /* ---------- hero ---------- */
  .hero { position: relative; isolation: isolate; overflow: hidden; padding: clamp(3rem, 9vw, 6.5rem) 0 0; }
  .hero::before {
    content: '';
    position: absolute;
    inset: 30% -20% -10%;
    z-index: -2;
    background: radial-gradient(50% 45% at 50% 55%, var(--glow-a), transparent 72%);
    pointer-events: none;
  }
  .hero-inner { display: grid; justify-items: center; gap: var(--space-5); width: min(100%, 56rem); margin-inline: auto; padding-inline: var(--gutter); text-align: center; }
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    padding: 6px 14px;
    border: 1px solid var(--border-soft);
    border-radius: 999px;
    background: color-mix(in srgb, var(--card) 70%, transparent);
    color: var(--body-2);
    font-size: var(--text-xs);
    font-weight: 500;
  }
  .badge-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--grad); }
  .hero-title {
    display: grid;
    margin: 0;
    font-size: clamp(2.5rem, 1.4rem + 4.6vw, 4.6rem);
    font-weight: 500;
    line-height: 1.04;
    letter-spacing: -0.03em;
  }
  .hero-lede { margin: 0; max-width: 46rem; color: var(--body-2); font-size: var(--text-lg); line-height: 1.55; }
  .hero-ctas { display: flex; flex-wrap: wrap; justify-content: center; gap: var(--space-3); }
  .hero-prompt { width: min(100%, 44rem); margin-top: var(--space-2); text-align: left; }
  .wave { display: block; width: 100%; height: clamp(9rem, 24vw, 20rem); margin-top: -2rem; pointer-events: none; }
  .wave path { fill: none; stroke: url(#wg); }
  .wave-glow { filter: blur(22px); opacity: 0.55; }
  .wave-glow path { stroke-width: 46; }
  .wave-line path { stroke-width: 1.6; opacity: 0.9; }
  .wave-line path:nth-child(2) { opacity: 0.55; }
  .wave-line path:nth-child(3) { opacity: 0.3; }
  .wave g { animation: drift 16s var(--ease-in-out) infinite alternate; transform-origin: center; }
  .wave-line { animation-delay: -4s !important; }
  @keyframes drift {
    from { transform: translateX(-2%) scaleY(1); }
    to { transform: translateX(2%) scaleY(1.12); }
  }
  .sources { display: grid; justify-items: center; gap: var(--space-3); padding: var(--space-4) var(--gutter) var(--space-7); margin-top: -3rem; }
  .sources-label { color: var(--muted); font-size: var(--text-xs); }
  .sources ul { display: flex; flex-wrap: wrap; justify-content: center; gap: var(--space-3) var(--space-6); margin: 0; padding: 0; list-style: none; }
  .sources li { display: inline-flex; align-items: center; gap: 8px; color: var(--body-2); font-size: var(--text-sm); font-weight: 500; }
  .sources li :global(svg) { color: var(--muted); }

  /* ---------- sections ---------- */
  .sec { padding-block: clamp(3.5rem, 8vw, 6rem); }
  .sec-head { display: grid; gap: var(--space-3); max-width: 40rem; margin-bottom: var(--space-7); }
  .sec-head.center { margin-inline: auto; text-align: center; justify-items: center; }
  .sec-head h2 { font-size: clamp(1.9rem, 1.4rem + 1.8vw, 2.9rem); font-weight: 500; letter-spacing: -0.02em; }
  .sec-head p { margin: 0; color: var(--body-2); font-size: var(--text-lg); }
  #cara-kerja .sec-head, #cara-kerja .sec-head { margin-inline: auto; text-align: center; justify-items: center; }

  /* ---------- how it works: three cards with a piece of the product ---------- */
  .steps { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--space-4); margin: 0; padding: 0; list-style: none; }
  .step { display: grid; align-content: start; gap: var(--space-2); padding: var(--space-4) var(--space-4) var(--space-5); background: var(--card); border: 1px solid var(--border-soft); border-radius: 20px; }
  .step h3 { font-size: var(--text-lg); font-weight: 500; }
  .step p { margin: 0; color: var(--body-2); font-size: var(--text-sm); line-height: 1.6; }
  .step-n { color: var(--muted) !important; font-size: var(--text-xs) !important; margin-top: var(--space-3) !important; }
  .vig {
    position: relative;
    display: grid;
    place-items: center;
    gap: 10px;
    min-height: 11rem;
    padding: var(--space-4);
    border-radius: 14px;
    background:
      radial-gradient(70% 60% at 50% 100%, color-mix(in srgb, var(--grad-1) 22%, transparent), transparent 70%),
      var(--sheet);
    overflow: hidden;
  }
  .v-prompt { display: flex; align-items: center; gap: 8px; width: min(100%, 16rem); padding: 8px 8px 8px 12px; border-radius: 999px; background: var(--card); box-shadow: var(--shadow-prompt); font-size: 0.78rem; color: var(--body-2); }
  .v-plus { color: var(--muted); }
  .v-url { flex: 1; text-align: left; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
  .v-send { display: grid; place-items: center; width: 24px; height: 24px; border-radius: 50%; background: var(--brand); color: var(--on-brand); }
  .v-chip { display: inline-flex; align-items: center; gap: 4px; font-size: 0.72rem; color: var(--body-2); }
  .v-scan { position: relative; width: 72px; height: 90px; }
  .v-scan img { width: 72px; height: 90px; object-fit: cover; border-radius: 6px; display: block; }
  .br { position: absolute; width: 12px; height: 12px; border: 0 solid var(--indigo); }
  .tl { left: -6px; top: -6px; border-width: 2px 0 0 2px; }
  .tr { right: -6px; top: -6px; border-width: 2px 2px 0 0; }
  .bl { left: -6px; bottom: -6px; border-width: 0 0 2px 2px; }
  .br2 { right: -6px; bottom: -6px; border-width: 0 2px 2px 0; }
  .v-beam { position: absolute; left: -6px; right: -6px; top: 0; height: 2px; background: var(--indigo); box-shadow: 0 0 12px 2px color-mix(in srgb, var(--indigo) 60%, transparent); animation: beam 2.6s var(--ease-in-out) infinite alternate; }
  @keyframes beam { from { top: 0; } to { top: calc(100% - 2px); } }
  .v-mods { display: flex; flex-wrap: wrap; justify-content: center; gap: 6px; margin: 0; padding: 0; list-style: none; font-size: 0.7rem; }
  .v-mods li { padding: 3px 8px; border-radius: 999px; background: var(--card); color: var(--muted); }
  .v-mods li.on { color: var(--ink); background: var(--indigo-soft); }
  .v-mods li.run { color: var(--indigo-text); }
  .v-verdict { display: grid; gap: 8px; width: min(100%, 15rem); padding: 14px; border-radius: 12px; background: var(--card); text-align: left; }
  .v-pill { justify-self: start; padding: 3px 10px; border-radius: 999px; background: var(--coral); color: var(--on-fill); font-size: 0.72rem; font-weight: 500; }
  .v-score { font-size: 1.6rem; color: var(--err); line-height: 1; }
  .v-score small { font-size: 0.75rem; color: var(--muted); }
  .v-meter { height: 6px; border-radius: 999px; background: var(--surface-2); overflow: hidden; }
  .v-meter span { display: block; width: 85%; height: 100%; background: var(--err); border-radius: inherit; }
  .v-line { height: 6px; border-radius: 3px; background: var(--surface-2); }
  .v-line.short { width: 60%; }

  /* ---------- modules ---------- */
  .mod-grid { display: grid; grid-template-columns: minmax(0, 5fr) minmax(0, 7fr); gap: var(--space-7); align-items: start; }
  .mod-lead { display: grid; gap: var(--space-4); }
  .mod-lead h2 { font-size: clamp(1.9rem, 1.4rem + 1.8vw, 2.9rem); font-weight: 500; letter-spacing: -0.02em; }
  .mod-lead > p { margin: 0; color: var(--body-2); font-size: var(--text-lg); }
  .mod-card {
    position: relative;
    display: grid;
    justify-items: start;
    gap: 6px;
    margin-top: var(--space-4);
    padding: var(--space-6);
    border: 1px solid var(--border-soft);
    border-radius: 24px;
    background:
      radial-gradient(80% 90% at 0% 0%, color-mix(in srgb, var(--grad-1) 30%, transparent), transparent 60%),
      radial-gradient(70% 80% at 100% 100%, color-mix(in srgb, var(--grad-3) 22%, transparent), transparent 60%),
      var(--card);
    overflow: hidden;
  }
  .mod-ring { display: grid; place-items: center; width: 72px; height: 72px; border-radius: 22px; background: var(--grad); color: #ffffff; }
  .mod-count { margin-top: var(--space-4); font-size: 3.5rem; font-weight: 400; line-height: 1; }
  .mod-caption { color: var(--body-2); }
  .mod-list { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-6) var(--space-5); margin: 0; padding: 0; list-style: none; }
  .mod { display: grid; align-content: start; gap: var(--space-2); }
  .mod-ic { color: var(--indigo-text); }
  .mod h3 { font-size: var(--text-lg); font-weight: 500; }
  .mod p { margin: 0; color: var(--body-2); font-size: var(--text-sm); line-height: 1.6; }
  .mod-line { color: var(--muted) !important; font-size: var(--text-xs) !important; }

  /* ---------- numbers, with the light beam ---------- */
  .beam-wrap { position: relative; }
  .beam {
    position: absolute;
    left: -10%;
    right: -10%;
    top: 50%;
    height: 2px;
    background: linear-gradient(90deg, transparent, var(--grad-1) 20%, #a8c7fa 50%, var(--grad-2) 80%, transparent);
    box-shadow: 0 0 24px 6px color-mix(in srgb, var(--grad-1) 45%, transparent);
    pointer-events: none;
  }
  .nums { position: relative; display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--space-4); margin: 0; padding: 0; list-style: none; }
  .num { display: grid; gap: var(--space-2); padding: var(--space-6) var(--space-5); background: color-mix(in srgb, var(--card) 92%, transparent); border: 1px solid var(--border-soft); border-radius: 20px; -webkit-backdrop-filter: blur(6px); backdrop-filter: blur(6px); }
  .num-value { margin: 0; font-size: clamp(2.6rem, 1.8rem + 2.6vw, 3.8rem); font-weight: 400; line-height: 1; }
  .num-value small { font-size: 0.45em; color: var(--muted); margin-left: 4px; }
  .num-label { margin: 0; color: var(--body-2); }
  .caveat { margin: var(--space-5) auto 0; max-width: 46rem; text-align: center; color: var(--muted); font-size: var(--text-sm); }

  /* ---------- the dashboard shot ---------- */
  .shot {
    margin: 0;
    padding: 10px;
    border: 1px solid var(--border-soft);
    border-radius: 22px;
    background: var(--card);
    box-shadow: 0 30px 80px -30px color-mix(in srgb, var(--grad-1) 55%, transparent);
  }
  .shot img { display: block; width: 100%; height: auto; border-radius: 14px; }
  .feats { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: var(--space-5); margin: var(--space-7) 0 0; padding: 0; list-style: none; }
  .feat { display: grid; align-content: start; gap: var(--space-2); }
  .feat-ic { color: var(--indigo-text); }
  .feat h3 { font-size: var(--text-base); font-weight: 500; }
  .feat p { margin: 0; color: var(--body-2); font-size: var(--text-sm); line-height: 1.6; }

  /* ---------- closing band ---------- */
  .band-sec { padding-block: var(--space-6) var(--space-4); }
  .band {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-5);
    padding: clamp(1.75rem, 4vw, 3rem);
    border: 1px solid var(--border-soft);
    border-radius: 24px;
    background:
      radial-gradient(60% 120% at 100% 50%, color-mix(in srgb, var(--grad-2) 26%, transparent), transparent 70%),
      var(--card);
  }
  .band h2 { font-size: clamp(1.5rem, 1.2rem + 1.2vw, 2.2rem); font-weight: 500; }
  .band p { margin: var(--space-2) 0 0; color: var(--body-2); }

  @media (max-width: 60rem) {
    .steps { grid-template-columns: minmax(0, 1fr); }
    .mod-grid { grid-template-columns: minmax(0, 1fr); }
    .feats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  }
  @media (max-width: 40rem) {
    .mod-list { grid-template-columns: minmax(0, 1fr); }
    .nums { grid-template-columns: minmax(0, 1fr); }
    .feats { grid-template-columns: minmax(0, 1fr); }
    .beam { top: 33%; }
  }
  @media (prefers-reduced-motion: reduce) {
    .wave g,
    .v-beam { animation: none; }
  }
</style>
