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

<section class="steps card" aria-labelledby="steps-title">
  <h2 id="steps-title">{t('radar.title')}</h2>
  <p class="summary">{t('radar.summary', { done: doneCount, total: order.length })}</p>
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
  .steps { max-width: 46rem; margin-top: var(--space-5); padding: clamp(1.25rem, 3vw, 1.75rem); }
  h2 { font-size: var(--text-lg); margin: 0 0 var(--space-1); }
  .summary { margin: 0 0 var(--space-4); font-size: var(--text-sm); color: var(--body-2); }
  ol { list-style: none; margin: 0; padding: 0; display: grid; gap: var(--space-3); }
  li { display: flex; align-items: center; gap: var(--space-3); flex-wrap: wrap; min-height: 32px; }
  .mark {
    width: 24px;
    height: 24px;
    flex: none;
    display: grid;
    place-items: center;
    border: 2px solid var(--border-strong);
    border-radius: 50%;
    background: transparent;
  }
  li.running .mark { border-color: var(--brand); border-top-color: transparent; animation: spin 0.9s linear infinite; }
  li.done .mark { background: var(--brand); border-color: var(--brand); color: var(--on-brand); }
  li.failed .mark, li.skipped .mark { border-color: var(--muted); }
  .dash { width: 10px; height: 2px; background: var(--muted); }
  .name { font-weight: 700; }
  .state { margin-left: auto; font-size: var(--text-sm); color: var(--body-2); }
  /* The detail sits on its own line so a late arrival never shifts the row. */
  .detail { flex-basis: 100%; padding-left: calc(24px + var(--space-3)); font-size: var(--text-sm); color: var(--body-2); }
  .wait { margin: var(--space-5) 0 0; padding-top: var(--space-4); border-top: 1px solid var(--border); font-size: var(--text-sm); color: var(--body-2); }

  @keyframes spin { to { transform: rotate(360deg); } }
  @media (prefers-reduced-motion: reduce) {
    li.running .mark { animation: none; border-top-color: var(--brand); }
  }
</style>
