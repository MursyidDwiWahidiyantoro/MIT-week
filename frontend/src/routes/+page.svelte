<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { t } from '$lib/i18n/index.svelte';
  import { readLink, type LinkProblem } from '$lib/link-input';
  import Icon from '$lib/components/ui/Icon.svelte';
  import UrlField from '$lib/components/ui/UrlField.svelte';
  import HowItWorks from '$lib/components/home/HowItWorks.svelte';

  const TRUST = ['home.trust1', 'home.trust2', 'home.trust3'] as const;

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
  <div class="wrap hero-wrap">
    <h1 class="hero-title">
      {t('home.h1.a')} <span class="accent">{t('home.h1.mark')}</span> {t('home.h1.b')}
    </h1>
    <p class="hero-lede">{t('home.lede')}</p>

    <form class="hero-form card" onsubmit={submit} novalidate aria-label={t('det.form')}>
      <UrlField id="hero-url" bind:value={url} bind:problem />
      <button class="btn btn--primary" type="submit">
        {t('det.start')}
        <Icon name="arrow-right" size={18} />
      </button>
    </form>
    <p class="hero-hint">
      {t('home.hint')} <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() plus a query string -->
      <a href={`${resolve('/deteksi')}?tab=file`}>{t('home.hintLink')}</a>.
    </p>

    <ul class="trust">
      {#each TRUST as item (item)}
        <li>
          <span class="tick" aria-hidden="true"><Icon name="check" size={16} /></span>
          <span>{t(item)}</span>
        </li>
      {/each}
    </ul>
  </div>
</section>

<HowItWorks />

<section class="family section" id="family" aria-labelledby="family-title">
  <div class="wrap">
    <div class="family-panel card">
      <h2 class="h2" id="family-title">{t('home.familyTitle')}</h2>
      <p class="family-lede">{t('home.familyLede')}</p>
      <div class="cta-row">
        <a class="btn btn--primary" href={resolve('/deteksi')}>{t('home.familyCheck')}</a>
        <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() plus a same-page anchor -->
        <a class="btn btn--quiet" href={`${resolve('/edukasi')}#keluarga`}>{t('home.familyGuide')}</a>
      </div>
    </div>
  </div>
</section>

<style>
  .hero {
    padding-block: clamp(2.5rem, 7vw, 5.5rem) clamp(2rem, 5vw, 3.5rem);
  }
  .hero-wrap > * { max-width: 44rem; }
  .hero-wrap > .hero-title { max-width: 14ch; }
  .hero-title {
    margin: 0 0 var(--space-4);
    font-size: clamp(2.4rem, 1.7rem + 3.6vw, 4.6rem);
    line-height: 1.04;
    max-width: 14ch;
  }
  .accent { color: var(--indigo-text); }
  .hero-lede {
    font-size: var(--text-lg);
    line-height: var(--leading-normal);
    color: var(--body-2);
    max-width: 40ch;
    margin: 0;
  }

  .hero-form {
    display: grid;
    gap: var(--space-4);
    margin-top: var(--space-6);
    padding: clamp(1.1rem, 3vw, 1.75rem);
    max-width: 44rem;
  }
  .hero-form .btn { justify-self: start; }
  .hero-hint { margin: var(--space-3) 0 0; color: var(--body-2); font-size: var(--text-sm); }
  .hero-hint { line-height: 2.2; }

  .trust {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3) var(--space-5);
    margin: var(--space-5) 0 0;
    padding: 0;
    list-style: none;
    font-weight: 700;
    font-size: var(--text-sm);
  }
  .trust li { display: inline-flex; align-items: center; gap: var(--space-2); }
  .tick { display: grid; place-items: center; width: 24px; height: 24px; border-radius: 50%; background: var(--indigo-soft); color: var(--indigo-text); }

  .family-panel { padding: clamp(1.5rem, 4vw, 2.75rem); }
  .family-lede { margin: var(--space-3) 0 0; color: var(--body-2); max-width: 36em; }
  .cta-row { display: flex; flex-wrap: wrap; gap: var(--space-3); margin-top: var(--space-5); }

  @media (max-width: 40rem) {
    .hero-form .btn,
    .cta-row .btn { width: 100%; }
  }
</style>
