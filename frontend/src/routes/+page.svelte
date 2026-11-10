<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { t, type MessageKey } from '$lib/i18n/index.svelte';
  import { readLink, type LinkProblem } from '$lib/link-input';
  import Icon from '$lib/components/ui/Icon.svelte';
  import UrlField from '$lib/components/ui/UrlField.svelte';

  // What one check covers, in the order a forwarded video is examined.
  const CHECKS: Array<{ icon: 'video' | 'search' | 'link' | 'image'; title: MessageKey; body: MessageKey }> = [
    { icon: 'video', title: 'home.checks.video.t', body: 'home.checks.video.b' },
    { icon: 'search', title: 'home.checks.fact.t', body: 'home.checks.fact.b' },
    { icon: 'link', title: 'home.checks.link.t', body: 'home.checks.link.b' },
    { icon: 'image', title: 'home.checks.image.t', body: 'home.checks.image.b' }
  ];

  let url = $state('');
  let problem = $state<LinkProblem | null>(null);

  // The check starts on /deteksi, which reads ?q= and runs it at once. The
  // link is cleaned here first, so a pasted share text never reaches the URL.
  function submit(event: SubmitEvent): void {
    event.preventDefault();
    const reading = readLink(url);
    problem = reading.problem;
    if (!reading.url) return;
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() is applied to the route; the query string is appended
    void goto(`${resolve('/deteksi')}?q=${encodeURIComponent(reading.url)}`);
  }
</script>

<svelte:head>
  <title>{t('home.title')}</title>
  <meta name="description" content={t('home.description')} />
</svelte:head>

<section class="hero" id="top">
  <div class="wrap hero-grid">
    <div class="hero-copy">
      <h1 class="hero-title">{t('home.h1')}</h1>
      <p class="hero-lede">{t('home.lede')}</p>

      <form class="checker glass" onsubmit={submit} novalidate aria-label={t('det.form')}>
        <UrlField id="hero-url" bind:value={url} bind:problem />
        <div class="checker-actions">
          <button class="btn btn--primary" type="submit">{t('det.start')}</button>
          <p class="checker-hint">
            {t('home.hint')}
            <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() plus a query string -->
            <a href={`${resolve('/deteksi')}?tab=file`}>{t('home.hintLink')}</a>
          </p>
        </div>
      </form>
    </div>

    <!-- The mechanism shown, not claimed: a forwarded fake seen through the
         verdict lens. The image is a labelled simulation from the demo set. -->
    <figure class="sample" aria-labelledby="sample-caption">
      <img
        src="/edukasi/quiz-diedit.jpg"
        alt={t('home.sample.alt')}
        width="360"
        height="480"
        fetchpriority="high"
        decoding="async"
      />
      <div class="lens glass">
        <p class="lens-verdict">
          <Icon name="alert" size={18} />
          <span>{t('home.sample.verdict')}</span>
        </p>
        <p class="lens-score tabular">{t('home.sample.score')}</p>
        <p class="lens-reason">{t('home.sample.reason')}</p>
      </div>
      <figcaption id="sample-caption">{t('home.sample.label')}</figcaption>
    </figure>
  </div>
</section>

<section class="checks" aria-labelledby="checks-title">
  <div class="wrap checks-grid">
    <h2 class="checks-title" id="checks-title">{t('home.checks.title')}</h2>
    <dl class="check-list">
      {#each CHECKS as check (check.title)}
        <div class="check">
          <dt><Icon name={check.icon} size={22} /><span>{t(check.title)}</span></dt>
          <dd>{t(check.body)}</dd>
        </div>
      {/each}
    </dl>
  </div>
</section>

<section class="learn">
  <div class="wrap learn-row">
    <p class="learn-text">{t('home.learn.text')}</p>
    <a class="btn" href={resolve('/edukasi')}>
      <Icon name="book" size={18} />
      {t('home.learn.cta')}
    </a>
  </div>
</section>

<style>
  .hero {
    padding-block: clamp(2.25rem, 6vw, 5rem) clamp(2.5rem, 6vw, 4.5rem);
  }
  .hero-grid {
    display: grid;
    grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
    gap: clamp(2rem, 5vw, 4.5rem);
    align-items: center;
  }
  .hero-title {
    margin: 0 0 var(--space-4);
    font-size: var(--text-display);
    line-height: var(--leading-display);
    letter-spacing: -0.035em;
    max-width: 15ch;
  }
  .hero-lede {
    margin: 0;
    max-width: 44ch;
    font-size: var(--text-lg);
    color: var(--body-2);
  }

  .checker {
    display: grid;
    gap: var(--space-4);
    margin-top: var(--space-6);
    padding: clamp(1.1rem, 3vw, 1.6rem);
    max-width: 40rem;
  }
  .checker-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-3) var(--space-5);
  }
  .checker-hint {
    margin: 0;
    flex: 1 1 16rem;
    color: var(--body-2);
    font-size: var(--text-sm);
  }
  .checker-hint a { color: var(--indigo-text); font-weight: 700; }

  /* The example: a phone-sized screenshot, the lens overlapping its foot. */
  .sample {
    position: relative;
    justify-self: center;
    width: min(100%, 22.5rem);
    margin: 0 0 var(--space-6);
  }
  .sample img {
    display: block;
    width: 100%;
    height: auto;
    border-radius: var(--radius-lg);
    box-shadow: 0 0 0 1px var(--glass-edge), var(--shadow-2);
  }
  .lens {
    position: absolute;
    left: -9%;
    right: -9%;
    bottom: -1.5rem;
    padding: var(--space-4) var(--space-5);
    border-radius: var(--radius-md);
  }
  .lens p { margin: 0; }
  .lens-verdict {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--err);
    font-weight: 700;
  }
  .lens-score {
    margin-top: 2px !important;
    font-size: var(--text-sm);
    color: var(--body-2);
  }
  .lens-reason {
    margin-top: var(--space-2) !important;
    font-size: var(--text-sm);
    color: var(--ink);
  }
  figcaption {
    position: absolute;
    top: var(--space-3);
    right: var(--space-3);
    padding: 2px 10px;
    border-radius: var(--radius-pill);
    background: var(--glass-bg-strong);
    color: var(--ink);
    font-size: var(--text-xs);
    font-weight: 700;
    box-shadow: 0 0 0 1px var(--glass-edge);
  }

  /* What is checked: a definition list on hairlines, not a row of cards. */
  .checks {
    padding-block: clamp(2.5rem, 6vw, 5rem);
    border-top: 1px solid var(--glass-edge);
  }
  .checks-grid {
    display: grid;
    grid-template-columns: minmax(0, 4fr) minmax(0, 8fr);
    gap: var(--space-6) clamp(2rem, 5vw, 4.5rem);
  }
  .checks-title {
    font-size: var(--text-2xl);
    max-width: 14ch;
  }
  .check-list {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 clamp(1.5rem, 3vw, 3rem);
    margin: 0;
  }
  .check {
    padding-block: var(--space-4) var(--space-5);
    border-top: 1px solid var(--border);
  }
  .check dt {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-weight: 700;
    color: var(--ink);
  }
  .check dt :global(svg) { color: var(--indigo-text); flex: none; }
  .check dd {
    margin: var(--space-2) 0 0;
    color: var(--body-2);
    font-size: var(--text-sm);
    line-height: 1.6;
  }

  .learn { padding-block: 0 clamp(3rem, 7vw, 5.5rem); }
  .learn-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4) var(--space-6);
    padding-top: var(--space-5);
    border-top: 1px solid var(--glass-edge);
  }
  .learn-text {
    margin: 0;
    font-family: var(--font-display);
    font-size: var(--text-xl);
    font-weight: 600;
    letter-spacing: -0.015em;
    max-width: 36ch;
  }

  @media (max-width: 60rem) {
    .hero-grid { grid-template-columns: minmax(0, 1fr); }
    .sample { margin-top: var(--space-4); width: min(86%, 21rem); }
    .checks-grid { grid-template-columns: minmax(0, 1fr); }
  }
  @media (max-width: 40rem) {
    .check-list { grid-template-columns: minmax(0, 1fr); }
    .checker-actions .btn { width: 100%; }
    .lens { left: -6%; right: -6%; }
  }
</style>
