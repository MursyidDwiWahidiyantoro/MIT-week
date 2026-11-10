<script lang="ts">
  import { t, locale, type MessageKey } from '$lib/i18n/index.svelte';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import '../lib/styles/tokens.css';
  import '../app.css';
  import '../lib/styles/fonts.css';
  import ThemeSwitch from '$lib/components/ui/ThemeSwitch.svelte';
  import LangSwitch from '$lib/components/a11y/LangSwitch.svelte';
  import '../lib/styles/refinement.css';
  import PreviewBadge from '$lib/components/ui/PreviewBadge.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  let { children } = $props();
  // On phones the display controls live behind one toggle, so the
  // header stays two rows and never overlaps the navigation.
  let toolsOpen = $state(false);

  const LINKS: Array<{ href: string; label: MessageKey; match: string[] }> = [
    { href: '/deteksi', label: 'nav.check', match: ['/deteksi', '/hasil'] },
    { href: '/edukasi', label: 'nav.learn', match: ['/edukasi'] },
    { href: '/tentang', label: 'nav.about', match: ['/tentang'] }
  ];

  // Move focus to the main heading on NAVIGATION only. Focusing on the initial
  // mount would paint a focus ring on a heading the user never touched, which
  // reads as a glitch.
  let previousPath: string | null = null;
  $effect(() => {
    const path = page.url.pathname;
    if (previousPath === null) {
      previousPath = path; // first render: record, do not focus
      return;
    }
    if (path === previousPath) return;
    previousPath = path;

    const main = document.getElementById('main');
    const heading = main?.querySelector('h1');
    if (heading) {
      heading.setAttribute('tabindex', '-1');
      heading.focus({ preventScroll: true });
    }
  });
</script>

<a class="skip-link" href="#main">{t('layout.skip')}</a>

<PreviewBadge />

<header class="site-header">
  <div class="wrap header-inner">
    <div class="lead">
      <a class="brand" href={resolve('/')} aria-label={t('layout.home')}>
        <span class="logo" aria-hidden="true"><Icon name="shield-check" size={22} /></span>
        <span class="wordmark">AMBA</span>
      </a>

      <nav aria-label={t('layout.nav')}>
        <ul>
          {#each LINKS as link (link.href)}
            <li>
              <a
                class="nav-link"
                href={resolve(link.href as '/' | '/deteksi' | '/edukasi' | '/tentang')}
                aria-current={link.match.includes(page.url.pathname) ? 'page' : undefined}
              >
                {t(link.label)}
              </a>
            </li>
          {/each}
        </ul>
      </nav>
    </div>

    <button
      class="tools-toggle"
      type="button"
      aria-expanded={toolsOpen}
      aria-controls="site-tools"
      aria-label={toolsOpen ? t('layout.displayClose') : t('layout.displayOpen')}
      onclick={() => (toolsOpen = !toolsOpen)}
    >
      <Icon name="type" size={20} />
      <span>{t('layout.display')}</span>
    </button>

    <div class="tools" id="site-tools" class:open={toolsOpen}>
      <LangSwitch />
      <ThemeSwitch />
    </div>
  </div>
</header>

<main id="main" tabindex="-1" class:is-home={page.url.pathname === '/'}>
  {#if page.url.pathname === '/'}
    {@render children()}
  {:else}
    <div class="wrap page-wrap">
      {@render children()}
    </div>
  {/if}
</main>

<footer class="site-footer">
  <div class="wrap footer-inner">
    <p class="footer-brand">AMBA</p>
    <p class="footer-note">
      {t('layout.footerNote1')} <strong>{t('layout.footerIndicative')}</strong>,
      {t('layout.footerNote2')}
    </p>
    <ul class="footer-links">
      <li><a href={resolve('/deteksi')}>{t('nav.check')}</a></li>
      <li><a href={resolve('/edukasi')}>{t('nav.learn')}</a></li>
      <li><a href={resolve('/tentang')}>{t('nav.about')}</a></li>
    </ul>
  </div>
  <!-- PRD §10: the legal disclaimer is kept verbatim, in Indonesian, in every language. -->
  <div class="wrap footer-legal">
    {#if locale.value === 'en'}<p class="footer-legal-note">{t('about.disclaimerNoteEn')}</p>{/if}
    <p class="footer-legal-text" lang="id">{t('disclaimer.full')}</p>
  </div>
</footer>

<style>
  /* The one glass surface every page shares: content scrolls under it. */
  .site-header {
    position: sticky;
    top: 0;
    z-index: 40;
    background: var(--glass-bg);
    -webkit-backdrop-filter: blur(var(--glass-blur)) saturate(1.35);
    backdrop-filter: blur(var(--glass-blur)) saturate(1.35);
    border-bottom: 1px solid var(--glass-edge);
    box-shadow: var(--glass-highlight);
  }
  .header-inner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3) var(--space-5);
    padding-block: 10px;
  }
  .lead {
    display: flex;
    align-items: center;
    gap: var(--space-5);
    min-width: 0;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    min-height: 44px;
    text-decoration: none;
    color: var(--ink);
    flex: none;
  }
  .brand:hover { color: var(--ink); }
  .logo {
    display: grid;
    place-items: center;
    width: 38px;
    height: 38px;
    background: var(--brand);
    color: var(--on-brand);
    border-radius: 11px;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.2), var(--shadow-1);
  }
  .wordmark {
    font-family: var(--font-display);
    font-weight: 650;
    font-size: 1.4rem;
    letter-spacing: -0.03em;
  }
  nav ul {
    display: flex;
    align-items: center;
    gap: 2px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .tools {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex: none;
  }
  .tools-toggle {
    display: none;
    align-items: center;
    gap: var(--space-2);
    min-height: 44px;
    padding: 0 14px;
    background: var(--glass-bg-strong);
    color: var(--ink);
    border: 1px solid var(--glass-edge);
    border-radius: var(--radius-sm);
    font: inherit;
    font-size: var(--text-sm);
    font-weight: 700;
    cursor: pointer;
  }
  .tools-toggle[aria-expanded='true'] { background: var(--white); box-shadow: var(--shadow-1); }

  main {
    padding-block: 0;
    min-height: calc(100dvh - 160px);
  }
  main:focus { outline: none; }
  .page-wrap {
    padding-block: clamp(1.75rem, 4vw, 3rem) clamp(3rem, 6vw, 4.5rem);
  }

  .site-footer {
    border-top: 1px solid var(--glass-edge);
    background: color-mix(in srgb, var(--paper) 55%, transparent);
    padding-block: clamp(2rem, 5vw, 3rem);
  }
  .footer-inner {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3) var(--space-6);
  }
  .footer-brand {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 650;
    font-size: 1.2rem;
    letter-spacing: -0.03em;
  }
  .footer-note {
    margin: 0;
    flex: 1 1 24rem;
    color: var(--body-2);
    font-size: var(--text-sm);
    max-width: 52ch;
  }
  .footer-links {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1) var(--space-4);
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .footer-links a {
    display: inline-flex;
    align-items: center;
    min-height: 44px;
    font-weight: 700;
    font-size: var(--text-sm);
    text-decoration: none;
    color: var(--body-2);
  }
  .footer-links a:hover { color: var(--ink); text-decoration: underline; }
  .footer-legal {
    margin-top: var(--space-5);
    padding-top: var(--space-4);
    border-top: 1px solid var(--glass-edge);
  }
  .footer-legal-note,
  .footer-legal-text {
    margin: 0;
    max-width: 96ch;
    color: var(--muted);
    font-size: var(--text-xs);
    line-height: 1.6;
  }
  .footer-legal-note { margin-bottom: var(--space-1); font-weight: 700; }

  /* Phones: row 1 brand + display toggle, row 2 the three destinations,
   * row 3 (opened on demand) the language and theme controls. */
  @media (max-width: 48rem) {
    .header-inner {
      display: grid;
      grid-template-columns: 1fr auto;
      align-items: center;
      gap: 8px 12px;
      padding-block: 8px;
    }
    .lead { display: contents; }
    .brand { grid-column: 1; grid-row: 1; }
    .logo { width: 36px; height: 36px; border-radius: 10px; }
    .wordmark { font-size: 1.3rem; }
    .tools-toggle {
      display: inline-flex;
      grid-column: 2;
      grid-row: 1;
    }
    .tools {
      display: none;
      grid-column: 1 / -1;
      grid-row: 3;
      flex-wrap: wrap;
      gap: 8px;
      justify-content: flex-start;
      padding-bottom: 4px;
    }
    .tools.open { display: flex; }
    nav {
      grid-column: 1 / -1;
      grid-row: 2;
      width: 100%;
    }
    nav ul {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 4px;
      width: 100%;
    }
    nav :global(.nav-link) {
      width: 100%;
      justify-content: center;
      padding: 0 4px;
      text-align: center;
      white-space: nowrap;
    }
  }
</style>
