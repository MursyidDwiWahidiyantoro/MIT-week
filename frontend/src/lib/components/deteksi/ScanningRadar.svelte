<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import { pillarSummary } from '$lib/pillar-summary';
  import type { PillarKey, StageStatus } from '$lib/audit.svelte';

  interface Props {
    stages: Record<PillarKey, StageStatus>;
    /** Pillar payloads already received, rendered under their own row. */
    partials?: Partial<Record<PillarKey, unknown>>;
  }

  let { stages, partials = {} }: Props = $props();


  const order: PillarKey[] = ['ela', 'ai', 'fact', 'link'];

  const doneCount = $derived(order.filter((k) => stages[k] === 'done').length);

  const statusText = (s: StageStatus): string => t(`stage.${s}`);
</script>

<div class="radar card">
  <p class="summary">
    {t('radar.summary', { done: doneCount, total: order.length })}
  </p>
  <ol>
    {#each order as key (key)}
      <li class={stages[key]}>
        <span class="dot" aria-hidden="true"></span>
        <span class="name">{t(`pillar.${key}`)}</span>
        <span class="state">{statusText(stages[key])}</span>
        {#if pillarSummary(key, partials[key])}
          <span class="detail">{pillarSummary(key, partials[key])}</span>
        {/if}
      </li>
    {/each}
  </ol>
</div>

<style>
  .radar { padding: clamp(1.25rem, 3vw, 1.75rem); }
  .summary {
    margin: 0 0 var(--space-4);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: var(--text-sm);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--indigo-text);
  }
  ol { list-style: none; margin: 0; padding: 0; display: grid; gap: var(--space-3); }
  li { display: flex; align-items: center; gap: var(--space-3); flex-wrap: wrap; }
  .dot {
    width: 18px;
    height: 18px;
    flex: none;
    border: var(--border-width) solid var(--border);
    border-radius: 50%;
    background: var(--white);
  }
  li.running .dot {
    background: var(--yellow);
    animation: pulse var(--dur-slow) var(--ease-in-out) infinite alternate;
  }
  li.done .dot { background: var(--safe); }
  li.failed .dot { background: var(--coral); }
  li.skipped .dot { background: var(--indigo-soft); }
  .name { font-weight: 700; }
  .state { margin-left: auto; font-size: var(--text-sm); color: var(--muted); }
  /* The detail sits on its own line so a late arrival never shifts the row. */
  .detail {
    flex-basis: 100%;
    padding-left: calc(18px + var(--space-3));
    font-size: var(--text-sm);
    color: var(--body-2);
  }

  @keyframes pulse {
    from { opacity: 0.45; }
    to { opacity: 1; }
  }

  @media (prefers-reduced-motion: reduce) {
    li.running .dot { animation: none; opacity: 1; }
  }
</style>
