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

  // Desktop: a side rail that can shrink to icons. Phones and tablets: the
  // same rail slides in as a drawer from the top bar's menu button.
  let collapsed = $state(false);
  let drawerOpen = $state(false);
  let closeButton: HTMLButtonElement | undefined = $state();

  const LINKS: Array<{ href: '/edukasi' | '/tentang'; label: MessageKey; icon: 'book' | 'info' }> = [
    { href: '/edukasi', label: 'nav.learn', icon: 'book' },
    { href: '/tentang', label: 'nav.about', icon: 'info' }
  ];
  const path = $derived(page.url.pathname);
  const onCheck = $derived(path === '/' || path === '/deteksi' || path === '/hasil');
  const wide = $derived(path.startsWith('/edukasi'));

  // Page changes run through the View Transitions API where it exists: the
  // page cross-fades and the prompt bar (view-transition-name: prompt)
  // glides between pages. Reduced motion skips it.
  onNavigate((navigation) => {
    if (!document.startViewTransition || reducedMotion()) return;
    return new Promise((done) => {
      document.startViewTransition(async () => {
        done();
        await navigation.complete;
      });
    });
  });

  // Move focus to the main heading on NAVIGATION only. Focusing on the initial
  // mount would paint a focus ring on a heading the user never touched.
  let previousPath: string | null = null;
  $effect(() => {
    const now = page.url.pathname;
    if (previousPath === null) {
      previousPath = now;
      return;
    }
    if (now === previousPath) return;
    previousPath = now;
    drawerOpen = false;
    const heading = document.getElementById('main')?.querySelector('h1');
    if (heading) {
      heading.setAttribute('tabindex', '-1');
      heading.focus({ preventScroll: true });
    }
  });

  $effect(() => {
    if (drawerOpen) closeButton?.focus();
  });
</script>

<svelte:window onkeydown={(e) => { if (e.key === 'Escape' && drawerOpen) drawerOpen = false; }} />

<a class="skip-link" href="#main">{t('layout.skip')}</a>

<PreviewBadge />

<div class="app" class:collapsed>
  <aside class="rail" class:open={drawerOpen} id="app-rail" aria-label={t('layout.nav')}>
    <div class="rail-head">
      <button
        class="icon-btn rail-toggle"
        type="button"
        aria-label={collapsed ? t('shell.expand') : t('shell.collapse')}
        aria-expanded={!collapsed}
        onclick={() => (collapsed = !collapsed)}
      >
        <Icon name="menu" size={22} />
      </button>
      <button
        class="icon-btn drawer-close"
        type="button"
        bind:this={closeButton}
        aria-label={t('shell.menuClose')}
        onclick={() => (drawerOpen = false)}
      >
        <Icon name="menu" size={22} />
      </button>
      <a class="brand" href={resolve('/')} aria-label={t('layout.home')}>
        <span class="logo" aria-hidden="true"><Icon name="shield-check" size={18} /></span>
        <span class="wordmark lbl">AMBA</span>
      </a>
    </div>

    <a class="new-check" href={resolve('/')} aria-current={onCheck ? 'page' : undefined} title={t('shell.newCheck')}>
      <Icon name="plus" size={20} />
      <span class="lbl">{t('shell.newCheck')}</span>
    </a>

    <nav aria-label={t('layout.nav')}>
      <ul>
        {#each LINKS as link (link.href)}
          <li>
            <a
              class="nav-link"
              href={resolve(link.href)}
              aria-current={path.startsWith(link.href) ? 'page' : undefined}
              title={t(link.label)}
            >
              <Icon name={link.icon} size={20} />
              <span class="lbl">{t(link.label)}</span>
            </a>
          </li>
        {/each}
      </ul>
    </nav>

    <div class="rail-foot">
      <ThemeSwitch withLabel />
    </div>
  </aside>

  {#if drawerOpen}
    <button
      class="scrim"
      type="button"
      tabindex="-1"
      aria-label={t('shell.menuClose')}
      onclick={() => (drawerOpen = false)}
      transition:fade={{ duration: reducedMotion() ? 0 : 200 }}
    ></button>
  {/if}

  <div class="main-col" inert={drawerOpen}>
    <header class="topbar">
      <button
        class="icon-btn"
        type="button"
        aria-label={t('shell.menuOpen')}
        aria-expanded={drawerOpen}
        aria-controls="app-rail"
        onclick={() => (drawerOpen = true)}
      >
        <Icon name="menu" size={22} />
      </button>
      <a class="brand" href={resolve('/')} aria-label={t('layout.home')}>
        <span class="logo" aria-hidden="true"><Icon name="shield-check" size={18} /></span>
        <span class="wordmark">AMBA</span>
      </a>
      <span class="spacer"></span>
      <ThemeSwitch />
    </header>

    <main id="main" tabindex="-1" class:is-home={path === '/'}>
      {#if path === '/'}
        {@render children()}
      {:else}
        <div class="wrap page-wrap" class:wrap--wide={wide}>
          {@render children()}
        </div>
      {/if}
    </main>

    <!-- PRD §10: the legal disclaimer stays verbatim on every page. -->
    <footer class="foot">
      <div class="wrap" class:wrap--wide={wide}>
        <p class="foot-line">{t('shell.footer')}</p>
        <p class="foot-legal">{t('disclaimer.full')}</p>
      </div>
    </footer>
  </div>
</div>

<style>
  .app { min-height: 100dvh; }

  /* ---------- the rail ---------- */
  .rail {
    position: fixed;
    inset: 0 auto 0 0;
    z-index: 60;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    width: min(86vw, 300px);
    padding: 12px;
    background: var(--rail-bg);
    overflow-y: auto;
    transform: translateX(-105%);
    transition: transform var(--dur-slow) var(--ease-out), width var(--dur-base) var(--ease-out);
    visibility: hidden;
  }
  .rail.open { transform: none; visibility: visible; box-shadow: var(--shadow-3); }
  .rail-head { display: flex; align-items: center; gap: 4px; min-height: 48px; }
  .rail-toggle { display: none; }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    min-height: 44px;
    padding: 0 8px;
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
    border-radius: 10px;
    background: var(--grad);
    color: #ffffff;
  }
  .wordmark { font-size: 1.2rem; font-weight: 500; letter-spacing: 0.02em; }

  .new-check {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-height: 48px;
    margin: var(--space-2) 0 var(--space-1);
    padding: 0 var(--space-4);
    border-radius: var(--radius-pill);
    background: var(--card);
    color: var(--ink);
    font-size: var(--text-sm);
    font-weight: 500;
    text-decoration: none;
    transition: background-color var(--dur-fast) var(--ease-out);
  }
  .new-check:hover { background: var(--surface-2); color: var(--ink); }
  .new-check[aria-current='page'] { background: var(--indigo-soft); }

  nav ul { display: grid; gap: 2px; margin: 0; padding: 0; list-style: none; }
  .rail-foot { margin-top: auto; padding-top: var(--space-3); }

  .scrim {
    position: fixed;
    inset: 0;
    z-index: 55;
    border: 0;
    background: rgba(0, 0, 0, 0.5);
    cursor: pointer;
  }

  /* ---------- the main column ---------- */
  .main-col { display: flex; flex-direction: column; min-height: 100dvh; min-width: 0; }
  .topbar {
    position: sticky;
    top: 0;
    z-index: 40;
    display: flex;
    align-items: center;
    gap: 4px;
    min-height: 60px;
    padding: 6px 8px;
    background: color-mix(in srgb, var(--paper) 88%, transparent);
    -webkit-backdrop-filter: blur(12px);
    backdrop-filter: blur(12px);
  }
  .spacer { flex: 1; }

  main { flex: 1; min-width: 0; }
  main:focus { outline: none; }
  .page-wrap { padding-block: clamp(1rem, 3vw, 2rem) clamp(2.5rem, 6vw, 4rem); }

  .foot { padding: var(--space-6) 0 calc(var(--space-6) + env(safe-area-inset-bottom)); }
  .foot-line { margin: 0 0 var(--space-2); color: var(--body-2); font-size: var(--text-sm); }
  .foot-legal { margin: 0; color: var(--muted); font-size: var(--text-xs); line-height: 1.6; }

  /* ---------- desktop: the rail sits beside the page ---------- */
  @media (min-width: 64rem) {
    .app {
      display: grid;
      grid-template-columns: var(--rail-w) minmax(0, 1fr);
      transition: grid-template-columns var(--dur-base) var(--ease-out);
    }
    .app.collapsed { grid-template-columns: var(--rail-w-collapsed) minmax(0, 1fr); }
    .rail {
      position: sticky;
      top: 0;
      width: auto;
      height: 100dvh;
      transform: none;
      visibility: visible;
      box-shadow: none;
    }
    .rail-toggle { display: inline-grid; }
    .drawer-close { display: none; }
    .collapsed .rail .lbl { display: none; }
    .collapsed .rail .brand { display: none; }
    .collapsed .rail .new-check,
    .collapsed .rail :global(.nav-link) {
      justify-content: center;
      width: 48px;
      padding: 0;
    }
    .topbar { display: none; }
    .scrim { display: none; }
  }

  @media (prefers-reduced-motion: reduce) {
    .rail,
    .app { transition: none; }
  }
</style>
