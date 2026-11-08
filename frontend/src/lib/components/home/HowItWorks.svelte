<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  const STEPS = [
    {
      n: '1',
      tone: 'indigo',
      title: 'how.1.title',
      body: 'how.1.body'
    },
    {
      n: '2',
      tone: 'lime',
      title: 'how.2.title',
      body: 'how.2.body'
    },
    {
      n: '3',
      tone: 'yellow',
      title: 'how.3.title',
      body: 'how.3.body'
    }
  ] as const;
</script>

<section class="section how-section" aria-labelledby="how-title">
  <div class="wrap">
    <p class="eyebrow">{t('how.eyebrow')}</p>
    <h2 class="h2" id="how-title">{t('how.title')}</h2>

    <ol class="steps">
      {#each STEPS as step (step.n)}
        <li class="step card lift">
          <span class="num num--{step.tone}" aria-hidden="true">{step.n}</span>
          <h3>{t(step.title)}</h3>
          <p>{t(step.body)}</p>
        </li>
      {/each}
    </ol>
  </div>
</section>

<style>
  .how-section {
    width: 100%;
    background: var(--paper);
    border-bottom: var(--border-width) solid var(--border);
    padding-block: clamp(3.5rem, 7vw, 6rem);
  }
  .steps {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-5);
    margin: var(--space-6) 0 0;
    padding: 0;
    list-style: none;
  }
  @media (min-width: 50rem) {
    .steps {
      grid-template-columns: repeat(3, 1fr);
      gap: clamp(1.25rem, 2.5vw, 2rem);
    }
  }
  .step {
    padding: clamp(1.4rem, 3vw, 2rem);
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .num {
    display: grid;
    place-items: center;
    width: 60px;
    height: 60px;
    margin-bottom: var(--space-4);
    border: var(--border-width) solid var(--border);
    border-radius: 18px;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.6em;
    box-shadow: var(--shadow-hard-xs);
  }
  .num--indigo { background: var(--indigo); color: var(--on-brand); }
  .num--lime { background: var(--lime); color: var(--on-fill); }
  .num--yellow { background: var(--yellow); color: var(--on-fill); }

  h3 { font-size: 1.4em; margin: 0 0 var(--space-2); }
  p { margin: 0; color: var(--body-2); font-size: 1.02em; line-height: 1.5; }
</style>
