<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { t, type MessageKey } from '$lib/i18n/index.svelte';
  import type { LinkProblem } from '$lib/link-input';
  import { inview } from '$lib/motion';
  import { handOver } from '$lib/handoff';
  import { EXAMPLES, openExample } from '$lib/demos';
  import Icon from '$lib/components/ui/Icon.svelte';
  import PromptBar from '$lib/components/ui/PromptBar.svelte';

  let url = $state('');
  let problem = $state<LinkProblem | null>(null);

  // The check runs on /deteksi: a link travels as ?q=, an image is handed
  // over in memory and the page starts the check as soon as it opens.
  function onlink(link: string): void {
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() is applied to the route; the query string is appended
    void goto(`${resolve('/deteksi')}?q=${encodeURIComponent(link)}`);
  }
  function onfile(file: File): void {
    handOver(file);
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() is applied to the route; the query string is appended
    void goto(`${resolve('/deteksi')}?tab=file&from=prompt`);
  }

  // What one check covers, in the order a forwarded video is examined.
  const CHECKS: Array<{ icon: 'video' | 'search' | 'link' | 'image'; title: MessageKey; body: MessageKey }> = [
    { icon: 'video', title: 'home.checks.video.t', body: 'home.checks.video.b' },
    { icon: 'search', title: 'home.checks.fact.t', body: 'home.checks.fact.b' },
    { icon: 'link', title: 'home.checks.link.t', body: 'home.checks.link.b' },
    { icon: 'image', title: 'home.checks.image.t', body: 'home.checks.image.b' }
  ];
</script>

<svelte:head>
  <title>{t('home.title')}</title>
  <meta name="description" content={t('home.description')} />
</svelte:head>

<section class="home" use:inview>
  <div class="glow" aria-hidden="true"></div>
  <div class="center">
    <h1 class="greet">
      <span class="g1 gradient-text rise" style:--i="0">{t('home.greet1')}</span>
      <span class="g2 rise" style:--i="1">{t('home.greet2')}</span>
    </h1>

    <div class="dock rise" style:--i="2">
      <PromptBar id="hero-url" bind:value={url} bind:problem {onlink} {onfile} />
    </div>

    <div class="chips rise" style:--i="3" role="group" aria-label={t('home.chips')}>
      {#each EXAMPLES as ex (ex.id)}
        <button class="chip-btn" type="button" onclick={() => openExample(ex.id)}>
          <Icon name={ex.icon} size={18} />
          <span>{t(ex.label)}</span>
        </button>
      {/each}
    </div>

    <p class="disclaimer rise" style:--i="4">
      {t('home.disclaimer')}
      <a href={resolve('/tentang')}>{t('home.disclaimerLink')}</a>
    </p>
  </div>
</section>

<section class="checks" aria-labelledby="checks-title" use:inview>
  <div class="wrap wrap--wide">
    <h2 id="checks-title" class="checks-title rise">{t('home.checks.title')}</h2>
    <ul class="check-grid">
      {#each CHECKS as check, i (check.title)}
        <li class="check rise" style:--i={i + 1}>
          <span class="ic" aria-hidden="true"><Icon name={check.icon} size={24} /></span>
          <h3>{t(check.title)}</h3>
          <p>{t(check.body)}</p>
        </li>
      {/each}
    </ul>
    <p class="learn rise" style:--i="5">
      {t('home.learn.text')}
      <a href={resolve('/edukasi')}>{t('home.learn.cta')}</a>
    </p>
  </div>
</section>

<style>
  /* ---------- the greeting screen ---------- */
  .home {
    position: relative;
    isolation: isolate;
    display: flex;
    min-height: calc(100dvh - 60px);
    overflow: hidden;
  }
  /* The soft blue light Gemini puts behind its greeting, breathing slowly. */
  .glow {
    position: absolute;
    inset: -10% -10% 0;
    z-index: -1;
    background:
      radial-gradient(48% 42% at 50% 62%, var(--glow-a), transparent 72%),
      radial-gradient(34% 26% at 50% 82%, var(--glow-b), transparent 72%);
    animation: breathe 9s var(--ease-in-out) infinite alternate;
    pointer-events: none;
  }
  @keyframes breathe {
    from { opacity: 0.75; transform: scale(1); }
    to { opacity: 1; transform: scale(1.06); }
  }
  .center {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: var(--space-5);
    width: min(100%, 50rem);
    margin-inline: auto;
    padding: var(--space-6) var(--gutter) var(--space-5);
  }
  .greet {
    display: grid;
    gap: 6px;
    margin: 0;
    text-align: center;
    font-size: clamp(1.75rem, 1.2rem + 2.2vw, 2.6rem);
    font-weight: 400;
    line-height: 1.22;
    letter-spacing: -0.01em;
  }
  .g1 { font-weight: 500; }
  .g2 { color: var(--muted); font-weight: 350; font-size: 0.72em; }
  .chips {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-2);
  }
  .chips :global(svg) { color: var(--indigo-text); }
  .disclaimer {
    margin: 0;
    text-align: center;
    color: var(--muted);
    font-size: var(--text-xs);
  }

  /* ---------- what one check covers ---------- */
  .checks { padding-block: var(--section-pad); }
  .checks-title { font-size: var(--text-2xl); margin-bottom: var(--space-5); }
  .check-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 15rem), 1fr));
    gap: var(--space-3);
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .check {
    display: grid;
    align-content: start;
    gap: var(--space-2);
    padding: var(--space-5);
    background: var(--card);
    border-radius: var(--radius-lg);
  }
  .ic { color: var(--indigo-text); }
  .check h3 { font-size: var(--text-lg); font-weight: 500; }
  .check p { margin: 0; color: var(--body-2); font-size: var(--text-sm); line-height: 1.6; }
  .learn { margin: var(--space-5) 0 0; color: var(--body-2); }
  .learn a { font-weight: 500; }

  /* ---------- phones: the prompt sits at the bottom, as in a chat app ---------- */
  @media (max-width: 47.99rem) {
    .center { justify-content: flex-start; padding-top: clamp(2rem, 14vh, 7rem); }
    .greet { order: 1; }
    .chips { order: 2; }
    .dock {
      order: 3;
      position: sticky;
      bottom: 0;
      margin-top: auto;
      padding-bottom: calc(var(--space-2) + env(safe-area-inset-bottom));
    }
    .disclaimer { order: 4; }
    .glow {
      background:
        radial-gradient(70% 34% at 50% 92%, var(--glow-a), transparent 72%),
        radial-gradient(60% 24% at 50% 100%, var(--glow-b), transparent 72%);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .glow { animation: none; }
  }
</style>
