<script lang="ts">
  import { t, locale, type MessageKey } from '$lib/i18n/index.svelte';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import '../lib/styles/tokens.css';
  import '../app.css';
  import '../lib/styles/fonts.css';
  import { mode } from '$lib/mode.svelte';
  import { textSize } from '$lib/a11y.svelte';
  import ThemeSwitch from '$lib/components/ui/ThemeSwitch.svelte';
  import TextSize from '$lib/components/a11y/TextSize.svelte';
  import ModeSwitch from '$lib/components/mode/ModeSwitch.svelte';
  import LangSwitch from '$lib/components/a11y/LangSwitch.svelte';
  import '../lib/styles/refinement.css';
  import PreviewBadge from '$lib/components/ui/PreviewBadge.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  let { children } = $props();
  // On phones the four display controls live behind one toggle, so the
  // header stays two rows and never overlaps the navigation.
  let toolsOpen = $state(false);

  const LINKS: Array<{ href: string; label: MessageKey; match: string[] }> = [
    { href: '/deteksi', label: 'nav.check', match: ['/deteksi', '/hasil'] },
    { href: '/cek-cepat', label: 'nav.quick', match: ['/cek-cepat'] },
    { href: '/edukasi', label: 'nav.learn', match: ['/edukasi'] },
    { href: '/tentang', label: 'nav.about', match: ['/tentang'] }
  ];

  // Read the stored preference once the component is live.
  $effect(() => {
    mode.init();
    textSize.init();
  });

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
    <a class="brand" href={resolve('/')} aria-label={t('layout.home')}>
      <span class="logo" aria-hidden="true"><Icon name="shield-check" size={24} /></span>
      <span class="wordmark">AMBA</span>
    </a>

    <nav aria-label={t('layout.nav')}>
      <ul>
        {#each LINKS as link (link.href)}
          <li>
            <a
              class="nav-link"
              href={resolve(link.href as '/' | '/deteksi' | '/cek-cepat' | '/edukasi' | '/tentang')}
              aria-current={link.match.includes(page.url.pathname) ? 'page' : undefined}
            >
              {t(link.label)}
            </a>
          </li>
        {/each}
      </ul>
    </nav>

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
      <TextSize />
      <LangSwitch />
      <ModeSwitch />
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
      <li><a href={resolve('/cek-cepat')}>{t('nav.quick')}</a></li>
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
  .site-header {
    background: var(--paper);
    border-bottom: var(--border-width) solid var(--border);
    position: sticky;
    top: 0;
    z-index: 40;
  }
  .header-inner {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3) var(--space-5);
    padding-block: 14px;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    text-decoration: none;
    color: var(--ink);
    flex: none;
    transition: transform var(--dur-fast) var(--ease-out);
  }
  .brand:hover { transform: translate(-2px, -2px); }
  .brand:hover .logo { box-shadow: var(--shadow-hard-sm); }
  .brand:active { transform: translate(1px, 1px); }
  .logo {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    background: var(--indigo);
    color: var(--on-brand);
    border: var(--border-width) solid var(--border);
    border-radius: 14px;
    box-shadow: var(--shadow-hard-xs);
    transition: box-shadow var(--dur-fast) var(--ease-out);
  }
  .wordmark {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.55em;
    letter-spacing: -0.03em;
  }
  nav ul {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2) var(--space-3);
    margin: 0;
    padding: 0;
    list-style: none;
  }
  nav { font-size: var(--text-base); }
  .tools {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex: none;
  }
  .tools-toggle {
    display: none;
    align-items: center;
    gap: var(--space-2);
    min-height: 44px;
    padding: 0 14px;
    background: var(--white);
    color: var(--ink);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    font: inherit;
    font-weight: 700;
    cursor: pointer;
  }
  .tools-toggle[aria-expanded='true'] { background: var(--brand); color: var(--on-brand); }
  .tools-toggle:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }

  main {
    padding-block: 0;
    min-height: calc(100dvh - 160px);
  }
  main:focus { outline: none; }
  .page-wrap {
    padding-block: clamp(1.5rem, 4vw, 2.5rem) clamp(2.5rem, 6vw, 4rem);
  }

  .site-footer {
    background: var(--white);
    border-top: var(--border-width) solid var(--border);
    padding-block: clamp(2rem, 5vw, 3rem);
  }
  .footer-inner {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3) var(--space-5);
  }
  .footer-brand {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.3em;
    letter-spacing: -0.03em;
  }
  .footer-note {
    margin: 0;
    flex: 1 1 24rem;
    color: var(--body-2);
    font-size: var(--text-sm);
    max-width: 46ch;
  }
  .footer-links {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2) var(--space-4);
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
    color: var(--ink);
    transition: color var(--dur-fast) var(--ease-out), transform var(--dur-fast) var(--ease-out);
  }
  .footer-links a:hover { color: var(--indigo-text); text-decoration: underline; transform: translateY(-2px); }
  .footer-legal {
    margin-top: var(--space-5);
    padding-top: var(--space-4);
    border-top: 1px solid var(--border);
  }
  .footer-legal-note,
  .footer-legal-text {
    margin: 0;
    max-width: 96ch;
    color: var(--body-2);
    font-size: var(--text-xs);
    line-height: 1.55;
  }
  .footer-legal-note { margin-bottom: var(--space-1); font-weight: 700; }

  /* Mobile: 2 deliberate rows (Row 1: brand + tools; Row 2: 4 equal nav links)
   * Keeps mobile header compact (~80px instead of 180px) and clean. */
  @media (max-width: 48rem) {
    .header-inner {
      display: grid;
      grid-template-columns: 1fr auto;
      align-items: center;
      gap: 10px 12px;
      padding-block: 10px;
    }
    .brand {
      grid-column: 1;
      grid-row: 1;
    }
    .logo {
      width: 38px;
      height: 38px;
      border-radius: 11px;
    }
    .wordmark {
      font-size: 1.35em;
    }
    /* Four segmented controls never fit beside the wordmark on a phone.
     * They sit in a panel under the navigation, opened by one toggle. */
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
      padding-top: 4px;
    }
    .tools.open { display: flex; }
    .tools :global(.text-size .label) { display: none; }
    nav {
      grid-column: 1 / -1;
      grid-row: 2;
      width: 100%;
    }
    nav ul {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 4px;
      width: 100%;
    }
    nav :global(.nav-link) {
      width: 100%;
      justify-content: center;
      padding: 0 4px;
      font-size: 0.88rem;
      min-height: 38px;
      text-align: center;
      white-space: nowrap;
    }
  }
</style>
