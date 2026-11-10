<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import { pillarSummary } from '$lib/pillar-summary';
  import type { PillarKey, StageStatus } from '$lib/audit.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  interface Props {
    stages: Record<PillarKey, StageStatus>;
    /** Pillar payloads already received, rendered under their own row. */
    partials?: Partial<Record<PillarKey, unknown>>;
    running?: boolean;
  }

  let { stages, partials = {}, running = false }: Props = $props();

  const order: PillarKey[] = ['ela', 'ai', 'fact', 'link'];
  const doneCount = $derived(order.filter((k) => stages[k] === 'done').length);
  const statusText = (s: StageStatus): string => t(`stage.${s}`);
</script>

<!-- The checks as a step list, the way Gemini shows its reasoning: a spinner
     on the step in progress, a tick when it is done. -->
<section class="station" aria-labelledby="steps-title">
  <h2 id="steps-title" class="sr-only">{t('radar.title')}</h2>
  <p class="summary tabular">{t('radar.summary', { done: doneCount, total: order.length })}</p>
  <ol>
    {#each order as key (key)}
      <li class={stages[key]}>
        <span class="mark" aria-hidden="true">
          {#if stages[key] === 'done'}<Icon name="check" size={14} />
          {:else if stages[key] === 'failed' || stages[key] === 'skipped'}<span class="dash"></span>
          {/if}
        </span>
        <span class="name">{t(`pillar.${key}`)}</span>
        <span class="state">{statusText(stages[key])}</span>
        {#if pillarSummary(key, partials[key])}
          <span class="detail">{pillarSummary(key, partials[key])}</span>
        {/if}
      </li>
    {/each}
  </ol>
  {#if running}<p class="wait">{t('det.waitNote')}</p>{/if}
</section>

<style>
  .station { display: grid; gap: var(--space-3); }
  .summary { margin: 0; font-size: var(--text-sm); color: var(--body-2); }
  ol { list-style: none; margin: 0; padding: 0; display: grid; gap: 2px; }
  li {
    display: grid;
    grid-template-columns: 28px minmax(0, 1fr) auto;
    align-items: center;
    gap: 4px var(--space-3);
    min-height: 44px;
    padding: 6px var(--space-3);
    border-radius: var(--radius-sm);
  }
  li.running { background: var(--surface-2); }
  .mark {
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    border: 2px solid var(--border);
    color: var(--on-brand);
  }
  li.running .mark {
    border-color: var(--indigo);
    border-right-color: transparent;
    animation: spin 0.9s linear infinite;
  }
  li.done .mark { background: var(--indigo); border-color: var(--indigo); color: var(--on-indigo-text); }
  li.failed .mark,
  li.skipped .mark { border-color: var(--border); }
  .dash { width: 8px; height: 2px; background: var(--muted); }
  .name { font-weight: 500; }
  .state { font-size: var(--text-sm); color: var(--muted); }
  li.running .state { color: var(--indigo-text); }
  li.done .state { color: var(--body-2); }
  /* The detail sits on its own line so a late arrival never shifts the row. */
  .detail { grid-column: 2 / -1; font-size: var(--text-sm); color: var(--body-2); }
  .wait { margin: 0; font-size: var(--text-sm); color: var(--muted); }

  @keyframes spin { to { transform: rotate(360deg); } }
  @media (prefers-reduced-motion: reduce) {
    li.running .mark { animation: none; border-right-color: var(--indigo); }
  }
</style>
