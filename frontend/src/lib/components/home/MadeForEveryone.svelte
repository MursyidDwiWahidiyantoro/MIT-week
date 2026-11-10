<script lang="ts">
  import Icon, { type IconName } from '$lib/components/ui/Icon.svelte';
  import { t, type MessageKey } from '$lib/i18n/index.svelte';

  const TILES: Array<{
    icon: IconName;
    title: MessageKey;
    body: MessageKey;
    tone: string;
  }> = [
    {
      icon: 'type',
      title: 'access.1.title',
      body: 'access.1.body',
      tone: 'tile--soft'
    },
    {
      icon: 'shield-check',
      title: 'access.2.title',
      body: 'access.2.body',
      tone: 'tile--yellow'
    },
    {
      icon: 'check',
      title: 'access.3.title',
      body: 'access.3.body',
      tone: 'tile--ink'
    },
    {
      icon: 'chat',
      title: 'access.4.title',
      body: 'access.4.body',
      tone: 'tile--safe'
    }
  ];
</script>

<section class="section access-section" id="access" aria-labelledby="access-title">
  <div class="wrap">
    <div class="access-grid">
      <div class="access-copy">
        <p class="eyebrow">{t('access.eyebrow')}</p>
        <h2 class="h2" id="access-title">{t('access.title')}</h2>
        <p class="lede">{t('access.lede')}</p>
      </div>

      <ul class="tiles">
        {#each TILES as tile (tile.title)}
          <li class="tile card lift {tile.tone}">
            <span class="icon" aria-hidden="true"><Icon name={tile.icon} size={30} /></span>
            <h3>{t(tile.title)}</h3>
            <p>{t(tile.body)}</p>
          </li>
        {/each}
      </ul>
    </div>
  </div>
</section>

<style>
  .access-section {
    width: 100%;
    background: var(--paper);
    border-bottom: var(--border-width) solid var(--border);
    padding-block: clamp(3.5rem, 7vw, 6rem);
  }
  .access-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: clamp(2rem, 4vw, 3.5rem);
    align-items: center;
  }
  @media (min-width: 60rem) {
    .access-grid {
      grid-template-columns: 1fr 1.3fr;
      gap: clamp(2.5rem, 5vw, 4.5rem);
    }
  }

  .access-copy { min-width: 0; }
  .access-copy .lede { margin-top: var(--space-4); max-width: 36ch; }

  .tiles {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-4);
    margin: 0;
    padding: 0;
    list-style: none;
    min-width: 0;
  }
  @media (min-width: 34rem) {
    .tiles {
      grid-template-columns: repeat(2, 1fr);
    }
  }

  .tile {
    min-width: 0;
    padding: clamp(1.2rem, 2.5vw, 1.6rem);
    border-radius: 24px;
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .tile--soft { background: var(--indigo-soft); }
  .tile--yellow { background: var(--yellow); }
  .tile--ink { background: var(--panel-dark); color: var(--on-panel-dark); box-shadow: var(--shadow-indigo); }
  .tile--safe { background: var(--safe); }
  .icon { display: inline-flex; margin-bottom: var(--space-3); }
  h3 { font-size: 1.3em; margin: 0 0 var(--space-2); }
  p { margin: 0; font-size: 0.98em; line-height: 1.5; }
  .tile--ink p { color: var(--on-panel-dark); }
</style>
