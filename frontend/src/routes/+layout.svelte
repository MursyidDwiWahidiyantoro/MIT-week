<script lang="ts">
  import { t, type MessageKey } from '$lib/i18n/index.svelte';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { onNavigate } from '$app/navigation';
  import { fade } from 'svelte/transition';
  import '../lib/styles/tokens.css';
  import '../app.css';
  import '../lib/styles/fonts.css';
  import '../lib/styles/refinement.css';
  import ThemeSwitch from '$lib/components/ui/ThemeSwitch.svelte';
  import PreviewBadge from '$lib/components/ui/PreviewBadge.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { reducedMotion } from '$lib/motion';

  let { children } = $props();

  // Two shells: the marketing pages (home, Belajar, Tentang, errors) sit
  // under a top navigation; the detector itself (/deteksi, /hasil) is a
  // dashboard in a tinted L-shaped frame around a work sheet.
  const path = $derived(page.url.pathname);
  const isApp = $derived(path === '/deteksi' || path === '/hasil');
  let menuOpen = $state(false);
  let closeButton: HTMLButtonElement | undefined = $state();

  const SITE: Array<{ href: string; label: MessageKey }> = [
    { href: '/#cara-kerja', label: 'nav.how' },
    { href: '/#modul', label: 'nav.modules' },
    { href: '/edukasi', label: 'nav.learn' },
    { href: '/tentang', label: 'nav.about' }
  ];
  const APP: Array<{ href: '/deteksi' | '/edukasi' | '/tentang' | '/'; label: MessageKey; icon: 'panel' | 'book' | 'info' | 'shield-check' }> = [
    { href: '/deteksi', label: 'nav.dashboard', icon: 'panel' },
    { href: '/edukasi', label: 'nav.learn', icon: 'book' },
    { href: '/tentang', label: 'nav.about', icon: 'info' },
    { href: '/', label: 'nav.home', icon: 'shield-check' }
  ];
  const appCurrent = (href: string): boolean =>
    href === '/deteksi' ? path === '/deteksi' || path === '/hasil' : path === href;

  // Page changes run through the View Transitions API where it exists.
  onNavigate((navigation) => {
    if (!document.startViewTransition || reducedMotion()) return;
    return new Promise((done) => {
      document.startViewTransition(async () => {
        done();
        await navigation.complete;
      });
    });
  });

  // Move focus to the main heading on NAVIGATION only.
  let previousPath: string | null = null;
  $effect(() => {
    const now = page.url.pathname;
    if (previousPath === null) {
      previousPath = now;
      return;
    }
    if (now === previousPath) return;
    previousPath = now;
    menuOpen = false;
    const heading = document.getElementById('main')?.querySelector('h1');
    if (heading) {
      heading.setAttribute('tabindex', '-1');
      heading.focus({ preventScroll: true });
    }
  });

  $effect(() => {
    if (menuOpen) closeButton?.focus();
  });
</script>

<svelte:window onkeydown={(e) => { if (e.key === 'Escape' && menuOpen) menuOpen = false; }} />

<a class="skip-link" href="#main">{t('layout.skip')}</a>

<PreviewBadge />

{#snippet brand()}
  <a class="brand" href={resolve('/')} aria-label={t('layout.home')}>
    <span class="logo" aria-hidden="true"><Icon name="shield-check" size={18} /></span>
    <span class="wordmark">AMBA</span>
  </a>
{/snippet}

{#if isApp}
  <!-- ================= the dashboard: L frame + work sheet ================= -->
  <div class="frame">
    <aside class="side" class:open={menuOpen} id="app-menu" aria-label={t('layout.nav')}>
      <div class="side-head">
        {@render brand()}
        <button class="icon-btn only-mobile" type="button" bind:this={closeButton} aria-label={t('shell.menuClose')} onclick={() => (menuOpen = false)}>
          <Icon name="menu" size={22} />
        </button>
      </div>
      <p class="tagline">{t('app.tagline')}</p>
      <a class="btn btn--primary btn--sm side-cta" href={resolve('/deteksi')}>
        <Icon name="plus" size={18} />{t('shell.newCheck')}
      </a>
      <nav aria-label={t('layout.nav')}>
        <ul>
          {#each APP as link (link.href)}
            <li>
              <a class="nav-link" href={resolve(link.href)} aria-current={appCurrent(link.href) ? 'page' : undefined}>
                <Icon name={link.icon} size={20} /><span>{t(link.label)}</span>
              </a>
            </li>
          {/each}
        </ul>
      </nav>
      <div class="side-foot"><ThemeSwitch withLabel /></div>
    </aside>

    {#if menuOpen}
      <button class="scrim" type="button" tabindex="-1" aria-label={t('shell.menuClose')} onclick={() => (menuOpen = false)} transition:fade={{ duration: reducedMotion() ? 0 : 200 }}></button>
    {/if}

    <div class="frame-main" inert={menuOpen}>
      <header class="app-top">
        <button class="icon-btn only-mobile" type="button" aria-label={t('shell.menuOpen')} aria-expanded={menuOpen} aria-controls="app-menu" onclick={() => (menuOpen = true)}>
          <Icon name="menu" size={22} />
        </button>
        <span class="only-mobile">{@render brand()}</span>
        <span class="crumb">
          <a href={resolve('/deteksi')}>{t('res.crumb')}</a>
          {#if path === '/hasil'}<span aria-hidden="true">/</span><span>{t('res.h1')}</span>{/if}
        </span>
        <span class="spacer"></span>
        <span class="only-mobile"><ThemeSwitch /></span>
      </header>
      <main id="main" tabindex="-1" class="sheet">
        {@render children()}
        <p class="sheet-legal">{t('disclaimer.full')}</p>
      </main>
    </div>
  </div>
{:else}
  <!-- ================= the site: top navigation + footer ================= -->
  <header class="site-top">
    <div class="site-top-inner">
      {@render brand()}
      <nav class="site-nav" aria-label={t('layout.nav')}>
        <ul>
          {#each SITE as link (link.href)}
            <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- in-page anchors and routes; resolve() is applied below -->
            <li><a class="site-link" href={link.href.startsWith('/#') ? `${resolve('/')}${link.href.slice(1)}` : resolve(link.href as '/edukasi' | '/tentang')} aria-current={path === link.href ? 'page' : undefined}>{t(link.label)}</a></li>
          {/each}
        </ul>
      </nav>
      <div class="site-actions">
        <ThemeSwitch />
        <a class="btn btn--primary btn--sm only-desktop" href={resolve('/deteksi')}>{t('land.cta')}</a>
        <button class="icon-btn only-mobile" type="button" aria-label={t('shell.menuOpen')} aria-expanded={menuOpen} aria-controls="site-menu" onclick={() => (menuOpen = true)}>
          <Icon name="menu" size={22} />
        </button>
      </div>
    </div>
  </header>

  {#if menuOpen}
    <div class="site-menu" id="site-menu" role="dialog" aria-modal="true" aria-label={t('layout.nav')} transition:fade={{ duration: reducedMotion() ? 0 : 160 }}>
      <div class="site-menu-head">
        {@render brand()}
        <button class="icon-btn" type="button" bind:this={closeButton} aria-label={t('shell.menuClose')} onclick={() => (menuOpen = false)}><Icon name="menu" size={22} /></button>
      </div>
      <ul>
        {#each SITE as link (link.href)}
          <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- in-page anchors and routes; resolve() is applied below -->
          <li><a class="nav-link" href={link.href.startsWith('/#') ? `${resolve('/')}${link.href.slice(1)}` : resolve(link.href as '/edukasi' | '/tentang')} onclick={() => (menuOpen = false)}>{t(link.label)}</a></li>
        {/each}
      </ul>
      <a class="btn btn--primary" href={resolve('/deteksi')} onclick={() => (menuOpen = false)}>{t('land.cta')}</a>
    </div>
  {/if}

  <main id="main" tabindex="-1" class:is-home={path === '/'} inert={menuOpen}>
    {#if path === '/'}
      {@render children()}
    {:else}
      <div class="wrap page-wrap" class:wrap--wide={path.startsWith('/edukasi')}>
        {@render children()}
      </div>
    {/if}
  </main>

  <footer class="site-foot" inert={menuOpen}>
    <div class="foot-grid">
      <div class="foot-col foot-about">
        {@render brand()}
        <p>{t('land.foot.about')}</p>
      </div>
      <div class="foot-col">
        <p class="foot-head">{t('land.foot.pages')}</p>
        <ul>
          <li><a href={resolve('/deteksi')}>{t('nav.dashboard')}</a></li>
          <li><a href={resolve('/edukasi')}>{t('nav.learn')}</a></li>
          <li><a href={resolve('/tentang')}>{t('nav.about')}</a></li>
        </ul>
      </div>
      <div class="foot-col">
        <p class="foot-head">{t('land.foot.source')}</p>
        <ul>
          <li><a href="https://turnbackhoax.id/" target="_blank" rel="noopener noreferrer">TurnBackHoax.id (Mafindo)</a></li>
        </ul>
      </div>
    </div>
    <p class="foot-legal">{t('disclaimer.full')}</p>
    <p class="foot-mark" aria-hidden="true">AMBA</p>
  </footer>
{/if}

<style>
  /* ---------- shared ---------- */
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    min-height: 44px;
    color: var(--ink);
    text-decoration: none;
  }
  .brand:hover { color: var(--ink); }
  .logo {
    display: grid;
    place-items: center;
    width: 30px;
    height: 30px;
    flex: none;
    border-radius: 9px;
    background: var(--grad);
    color: #ffffff;
  }
  .wordmark { font-size: 1.15rem; font-weight: 500; letter-spacing: 0.03em; }
  .spacer { flex: 1; }
  .scrim { position: fixed; inset: 0; z-index: 55; border: 0; background: rgba(0, 0, 0, 0.55); cursor: pointer; }

  /* ---------- the dashboard frame ---------- */
  .frame { min-height: 100dvh; background: var(--frame); }
  .side {
    position: fixed;
    inset: 0 auto 0 0;
    z-index: 60;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    width: min(86vw, 280px);
    padding: 12px 14px;
    background: var(--frame);
    overflow-y: auto;
    transform: translateX(-105%);
    visibility: hidden;
    transition: transform var(--dur-slow) var(--ease-out);
  }
  .side.open { transform: none; visibility: visible; box-shadow: var(--shadow-3); }
  .side-head { display: flex; align-items: center; justify-content: space-between; min-height: 48px; padding-left: 4px; }
  .tagline { margin: 0 0 var(--space-2); padding-left: 4px; color: var(--muted); font-size: var(--text-xs); }
  .side-cta { justify-content: flex-start; margin-bottom: var(--space-3); }
  .side nav ul { display: grid; gap: 2px; margin: 0; padding: 0; list-style: none; }
  .side :global(.nav-link[aria-current='page']) { background: var(--surface-2); }
  .side-foot { margin-top: auto; }
  .frame-main { display: flex; flex-direction: column; min-height: 100dvh; min-width: 0; }
  .app-top {
    display: flex;
    align-items: center;
    gap: 6px;
    height: 48px;
    padding: 0 8px;
    background: var(--frame);
  }
  .crumb { display: none; align-items: center; gap: 8px; color: var(--muted); font-size: var(--text-sm); }
  .crumb a { display: inline-flex; align-items: center; min-height: 44px; padding: 0 4px; color: var(--body-2); text-decoration: none; }
  .crumb a:hover { color: var(--ink); }
  .sheet {
    flex: 1;
    min-width: 0;
    padding: clamp(16px, 2.4vw, 24px);
    background: var(--sheet);
  }
  .sheet:focus { outline: none; }
  .sheet-legal { margin: var(--space-7) 0 0; color: var(--muted); font-size: var(--text-xs); line-height: 1.6; max-width: 110ch; }

  @media (min-width: 64rem) {
    .frame { display: grid; grid-template-columns: 232px minmax(0, 1fr); }
    .side {
      position: sticky;
      top: 0;
      width: auto;
      height: 100dvh;
      transform: none;
      visibility: visible;
      box-shadow: none;
      padding: 12px 12px 16px;
    }
    .crumb { display: inline-flex; padding-left: 16px; }
    .frame-main { padding: 0 8px 8px 0; }
    .sheet { border: 1px solid var(--sheet-edge); border-radius: 16px; }
    .only-mobile { display: none !important; }
    .scrim { display: none; }
  }

  /* ---------- the site navigation ---------- */
  .site-top {
    position: sticky;
    top: 0;
    z-index: 40;
    background: color-mix(in srgb, var(--paper) 82%, transparent);
    -webkit-backdrop-filter: blur(14px);
    backdrop-filter: blur(14px);
    border-bottom: 1px solid var(--border-soft);
  }
  .site-top-inner {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    width: 100%;
    max-width: var(--maxw-wide);
    min-height: 64px;
    margin-inline: auto;
    padding-inline: var(--gutter);
  }
  .site-nav { flex: 1; display: none; justify-content: center; }
  .site-nav ul { display: flex; gap: 4px; margin: 0; padding: 0; list-style: none; }
  .site-link {
    display: inline-flex;
    align-items: center;
    min-height: 40px;
    padding: 0 14px;
    border-radius: 999px;
    color: var(--body-2);
    font-size: var(--text-sm);
    font-weight: 500;
    text-decoration: none;
  }
  .site-link:hover { background: var(--hover-overlay); color: var(--ink); }
  .site-link[aria-current='page'] { color: var(--ink); }
  .site-actions { display: flex; align-items: center; gap: var(--space-2); margin-left: auto; }
  .site-menu {
    position: fixed;
    inset: 0;
    z-index: 70;
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: 12px var(--gutter) var(--space-6);
    background: var(--paper);
  }
  .site-menu-head { display: flex; align-items: center; justify-content: space-between; min-height: 52px; }
  .site-menu ul { display: grid; gap: 4px; margin: 0; padding: 0; list-style: none; }
  .site-menu :global(.nav-link) { font-size: var(--text-lg); min-height: 52px; }

  main { min-width: 0; }
  main:focus { outline: none; }
  .page-wrap { padding-block: clamp(1.5rem, 4vw, 3rem) clamp(2.5rem, 6vw, 4rem); }

  /* ---------- the site footer: columns, legal, then the name built large ---------- */
  .site-foot {
    position: relative;
    overflow: hidden;
    margin-top: var(--space-7);
    padding: var(--space-7) var(--gutter) 0;
    border-top: 1px solid var(--border-soft);
  }
  .foot-grid {
    display: grid;
    grid-template-columns: minmax(0, 2fr) repeat(2, minmax(0, 1fr));
    gap: var(--space-6);
    max-width: var(--maxw-wide);
    margin-inline: auto;
  }
  .foot-col p { margin: 0; color: var(--body-2); font-size: var(--text-sm); line-height: 1.6; }
  .foot-about { display: grid; gap: var(--space-3); max-width: 26rem; }
  .foot-head { font-weight: 500; color: var(--ink) !important; margin-bottom: var(--space-2) !important; }
  .foot-col ul { list-style: none; margin: 0; padding: 0; display: grid; gap: 2px; }
  .foot-col a { display: inline-flex; min-height: 44px; align-items: center; color: var(--body-2); font-size: var(--text-sm); text-decoration: none; }
  .foot-col a:hover { color: var(--ink); text-decoration: underline; }
  .foot-legal { max-width: var(--maxw-wide); margin: var(--space-6) auto 0; color: var(--muted); font-size: var(--text-xs); line-height: 1.6; }
  /* The name, set across the foot of the page and fading into it. */
  .foot-mark {
    margin: var(--space-6) 0 -0.18em;
    text-align: center;
    font-size: clamp(5rem, 26vw, 22rem);
    font-weight: 600;
    line-height: 0.8;
    letter-spacing: -0.04em;
    /* A solid, quiet ink that fades out downward through a mask. */
    color: color-mix(in srgb, var(--ink) 16%, transparent);
    -webkit-mask-image: linear-gradient(180deg, #000 10%, transparent 90%);
    mask-image: linear-gradient(180deg, #000 10%, transparent 90%);
    user-select: none;
  }

  @media (min-width: 60rem) {
    .site-nav { display: flex; }
    .site-top .only-mobile { display: none !important; }
  }
  @media (max-width: 59.99rem) {
    .only-desktop { display: none !important; }
    .foot-grid { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
    .foot-about { grid-column: 1 / -1; }
  }

  @media (prefers-reduced-motion: reduce) {
    .side { transition: none; }
  }
</style>
