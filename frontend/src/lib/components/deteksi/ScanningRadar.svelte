<script lang="ts">
  import { mode } from '$lib/mode.svelte';
  import type { PillarKey, StageStatus } from '$lib/audit.svelte';

  interface Props {
    stages: Record<PillarKey, StageStatus>;
  }

  let { stages }: Props = $props();

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
  li { display: flex; align-items: center; gap: var(--space-3); color: var(--text-2); }
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

  @keyframes pulse { from { opacity: 0.4; } to { opacity: 1; } }

  /* Calm mode: no looping animation at all, just static states. */
  .radar[data-calm='true'] li.running .dot { animation: none; opacity: 1; }
</style>
