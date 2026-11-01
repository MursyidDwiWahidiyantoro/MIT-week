<script lang="ts">
  import { mode } from '$lib/mode.svelte';
  import { pillarSummary } from '$lib/pillar-summary';
  import type { PillarKey, StageStatus } from '$lib/audit.svelte';

  interface Props {
    stages: Record<PillarKey, StageStatus>;
    /** Pillar payloads already received, rendered under their own row. */
    partials?: Partial<Record<PillarKey, unknown>>;
  }

  let { stages, partials = {} }: Props = $props();

  const labels: Record<PillarKey, string> = {
    ela: 'Analisis gambar',
    ai: 'Deteksi jejak AI',
    fact: 'Pengecekan fakta',
    link: 'Pemeriksaan tautan'
  };
  const order: PillarKey[] = ['ela', 'ai', 'fact', 'link'];

  const doneCount = $derived(order.filter((k) => stages[k] === 'done').length);
  const calm = $derived(mode.resolved === 'essential');

  const STATUS_TEXT: Record<StageStatus, string> = {
    idle: 'menunggu',
    running: 'sedang diperiksa',
    done: 'selesai',
    skipped: 'dilewati',
    failed: 'tidak dapat diperiksa'
  };
  const statusText = (s: StageStatus): string => STATUS_TEXT[s];

</script>

<div class="radar" data-calm={calm}>
  <p class="summary">
    {doneCount} dari {order.length} pemeriksaan selesai
  </p>
  <ol>
    {#each order as key (key)}
      <li class={stages[key]}>
        <span class="dot" aria-hidden="true"></span>
        <span>{labels[key]}</span>
        <span class="state">{statusText(stages[key])}</span>
        {#if pillarSummary(key, partials[key])}
          <span class="detail">{pillarSummary(key, partials[key])}</span>
        {/if}
      </li>
    {/each}
  </ol>
</div>

<style>
  .radar {
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
    padding: var(--space-5);
  }
  .summary {
    margin: 0 0 var(--space-4);
    font-family: var(--font-mono);
    font-size: var(--text-xs);
    color: var(--text-2);
    letter-spacing: 0.06em;
  }
  ol { list-style: none; margin: 0; padding: 0; display: grid; gap: var(--space-3); }
  li { display: flex; align-items: center; gap: var(--space-3); color: var(--text-2); flex-wrap: wrap; }
  .dot {
    width: 10px; height: 10px; border-radius: 50%;
    border: var(--border-width) solid var(--border-strong);
    flex: none;
  }
  li.running .dot {
    border-color: var(--accent);
    background: var(--accent);
    animation: pulse var(--dur-slow) var(--ease-in-out) infinite alternate;
  }
  li.done .dot { border-color: var(--ok); background: var(--ok); }
  li.failed .dot, li.skipped .dot { border-color: var(--text-3); }
  li.done { color: var(--text); }
  .state { margin-left: auto; font-size: var(--text-xs); color: var(--text-3); }
  /* The detail sits on its own line so a late arrival never shifts the row. */
  .detail {
    flex-basis: 100%;
    padding-left: calc(10px + var(--space-3));
    font-size: var(--text-sm);
    color: var(--text-3);
  }

  @keyframes pulse { from { opacity: 0.4; } to { opacity: 1; } }

  /* Calm mode: no looping animation at all, just static states. */
  .radar[data-calm='true'] li.running .dot { animation: none; opacity: 1; }
</style>
