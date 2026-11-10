<script lang="ts">
  import Icon from '$lib/components/ui/Icon.svelte';
  import { t } from '$lib/i18n/index.svelte';

  const LIMITS = ['about.l1', 'about.l2', 'about.l3', 'about.l4', 'about.l5'] as const;
</script>

<svelte:head>
  <title>{t('about.title')}</title>
  <meta name="description" content={t('about.description')} />
</svelte:head>

<section class="head">
  <h1 class="h1">AMBA</h1>
  <p class="lede">{t('about.lede')}</p>
</section>

<section class="team" aria-labelledby="team-title">
  <h2 class="h2" id="team-title">{t('about.teamTitle')}</h2>
  <ul class="team-grid">
    <li><strong>Mursyid D. W.</strong><span>PM &amp; Lead Architect</span></li>
    <li><strong>M. Ridho A.</strong><span>UI/UX Lead</span></li>
    <li><strong>Bagus Joshua P.</strong><span>Golang Backend Lead</span></li>
    <li><strong>Jericho M. T.</strong><span>Frontend Lead</span></li>
  </ul>
</section>

<section class="data card lift" aria-labelledby="data-title">
  <h2 class="h2" id="data-title">{t('about.dataTitle')}</h2>
  <dl>
    <div class="row">
      <dt><span class="tag tag--send">{t('about.sent')}</span></dt>
      <dd>{t('about.sentBody')}</dd>
    </div>
    <div class="row">
      <dt><span class="tag tag--no">{t('about.notSent')}</span></dt>
      <dd>{t('about.notSentBody')}</dd>
    </div>
    <div class="row">
      <dt><span class="tag tag--none">{t('about.stored')}</span></dt>
      <dd>{t('about.storedBody')}</dd>
    </div>
  </dl>
</section>

<section class="card limits lift" id="batasan" aria-labelledby="limits-title">
  <h2 class="h2" id="limits-title">{t('about.limitsTitle')}</h2>
  <ul>
    {#each LIMITS as item (item)}
      <li>
        <span class="bullet" aria-hidden="true"><Icon name="alert" size={16} /></span>
        {t(item)}
      </li>
    {/each}
  </ul>
</section>

<details class="legal">
  <summary>{t('about.legal')}</summary>
  <p>{t('disclaimer.full')}</p>
</details>

<style>
  .head { max-width: 54rem; margin-bottom: var(--space-7); }
  h1 { margin-bottom: var(--space-3); }

  .team { margin-bottom: var(--space-7); }
  .team-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-3);
    margin: var(--space-5) 0 0;
    padding: 0;
    list-style: none;
  }
  @media (min-width: 48rem) { .team-grid { grid-template-columns: repeat(4, 1fr); } }
  .team-grid li {
    display: grid;
    gap: var(--space-1);
    padding: var(--space-4);
    background: var(--white);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-hard-xs);
  }
  .team-grid strong { font-family: var(--font-display); font-size: 1.05em; }
  .team-grid span { color: var(--body-2); font-size: var(--text-sm); }

  .data { padding: clamp(1.25rem, 3vw, 2rem); }
  dl { display: grid; gap: var(--space-5); margin: var(--space-5) 0 0; }
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: var(--space-3) var(--space-5);
    padding-top: var(--space-4);
    border-top: var(--border-width) dashed var(--border);
  }
  .row:first-child { padding-top: 0; border-top: 0; }
  dt { flex: 0 0 auto; }
  dd { flex: 1 1 22rem; margin: 0; min-width: 0; max-width: var(--measure); }
  .tag {
    display: inline-flex;
    padding: 4px 14px;
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-pill);
    font-weight: 700;
    font-size: var(--text-sm);
    white-space: nowrap;
  }
  .tag--send { background: var(--yellow); }
  .tag--no { background: var(--coral); }
  .tag--none { background: var(--safe); }

  .limits { margin-top: var(--space-5); padding: clamp(1.25rem, 3vw, 1.9rem); background: var(--yellow); }
  ul { margin: var(--space-4) 0 0; padding: 0; list-style: none; display: grid; gap: var(--space-3); }
  li { display: flex; align-items: flex-start; gap: var(--space-3); line-height: 1.45; }
  .bullet {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    flex: none;
    background: var(--white);
    border: 2px solid var(--border);
    border-radius: 50%;
  }

  .legal { margin-top: var(--space-6); max-width: 52rem; }
  .legal summary { display: inline-flex; align-items: center; min-height: 44px; color: var(--body-2); font-size: var(--text-sm); font-weight: 500; cursor: pointer; }
  .legal p { margin: var(--space-2) 0 0; color: var(--muted); font-size: var(--text-xs); line-height: 1.6; }
</style>
