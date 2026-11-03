<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import '../lib/styles/tokens.css';
  import '../app.css';
  import '@fontsource-variable/bricolage-grotesque';
  import '@fontsource/atkinson-hyperlegible/400.css';
  import '@fontsource/atkinson-hyperlegible/700.css';
  import { mode } from '$lib/mode.svelte';
  import { textSize } from '$lib/a11y.svelte';
  import ModeSwitch from '$lib/components/mode/ModeSwitch.svelte';
  import TextSize from '$lib/components/a11y/TextSize.svelte';
  import PreviewBadge from '$lib/components/ui/PreviewBadge.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  let { children } = $props();

  const LINKS = [
    { href: '/deteksi', label: 'Periksa', match: ['/deteksi', '/hasil'] },
    { href: '/cek-cepat', label: 'Cek cepat', match: ['/cek-cepat'] },
    { href: '/edukasi', label: 'Belajar', match: ['/edukasi'] },
    { href: '/tentang', label: 'Tentang', match: ['/tentang'] }
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

<a class="skip-link" href="#main">Langsung ke isi utama</a>

<PreviewBadge />

<header class="site-header">
  <div class="wrap header-inner">
    <a class="brand" href={resolve('/')} aria-label="AMBA, ke beranda">
      <span class="logo" aria-hidden="true"><Icon name="shield-check" size={24} /></span>
      <span class="wordmark">AMBA</span>
    </a>

    <nav aria-label="Navigasi utama">
      <ul>
        {#each LINKS as link (link.href)}
          <li>
            <a
              class="nav-link"
              href={resolve(link.href as '/' | '/deteksi' | '/cek-cepat' | '/edukasi' | '/tentang')}
              aria-current={link.match.includes(page.url.pathname) ? 'page' : undefined}
            >
              {link.label}
            </a>
          </li>
        {/each}
      </ul>
    </nav>

    <div class="tools">
      <TextSize />
      <ModeSwitch />
    </div>
  </div>
</header>

<main id="main" tabindex="-1">
  <div class="wrap">
    {@render children()}
  </div>
</main>

<footer class="site-footer">
  <div class="wrap footer-inner">
    <p class="footer-brand">AMBA</p>
    <p class="footer-note">
      AMBA membantu Anda menilai risiko informasi. Hasilnya <strong>indikatif</strong>,
      bukan bukti pasti dan bukan alat bukti hukum.
    </p>
    <ul class="footer-links">
      <li><a href={resolve('/deteksi')}>Periksa</a></li>
      <li><a href={resolve('/edukasi')}>Belajar</a></li>
      <li><a href={resolve('/cek-cepat')}>Cek cepat</a></li>
      <li><a href={resolve('/tentang')}>Tentang</a></li>
    </ul>
  </div>
</footer>

<style>
  .site-header {
    background: var(--paper);
    border-bottom: var(--border-width) solid var(--ink);
  }
  .header-inner {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3) var(--space-5);
    padding-block: 18px;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    text-decoration: none;
    color: var(--ink);
  }
  .logo {
    display: grid;
    place-items: center;
    width: 46px;
    height: 46px;
    background: var(--indigo);
    color: var(--white);
    border: var(--border-width) solid var(--ink);
    border-radius: 14px;
    box-shadow: var(--shadow-hard-xs);
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
    gap: var(--space-2) var(--space-5);
    margin: 0;
    padding: 0;
    list-style: none;
  }
  nav { font-size: var(--text-base); }
  .tools {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-3);
  }
  main { padding-block: clamp(1.5rem, 4vw, 2.5rem) 0; }
  main:focus { outline: none; }

  .site-footer {
    margin-top: var(--space-9);
    background: var(--white);
    border-top: var(--border-width) solid var(--ink);
    padding-block: var(--space-6);
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
  }
  .footer-links a:hover { color: var(--indigo-text); text-decoration: underline; }

  /* Phone: three deliberate rows (brand, nav, tools) instead of the four or
   * five an unmanaged wrap produces. Nav stays on ONE line so it never reads
   * as a broken menu. */
  @media (max-width: 44rem) {
    .header-inner { gap: 8px 12px; padding-block: 14px; }
    .brand { order: 1; }
    nav { order: 2; flex-basis: 100%; }
    nav ul { gap: 0; }
    /* 360px is the tightest phone in the brief. Four links + the 2.5px borders
     * must fit one line, so the padding is tuned to that width. */
    nav :global(.nav-link) { padding: 0 5px; font-size: 0.95rem; }
    .tools { order: 3; flex-basis: 100%; justify-content: flex-start; gap: var(--space-2); }
    .tools :global(.text-size .label) { display: none; }
  }
</style>
