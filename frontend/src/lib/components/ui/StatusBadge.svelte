<script lang="ts">
  import { ShieldCheck, AlertTriangle, AlertOctagon, HelpCircle } from '@lucide/svelte';
  import type { Status } from '$lib/api/types';
  import { statusLabel } from '$lib/format';

  interface Props {
    status: Status;
    score?: number | null;
  }

  let { status, score = null }: Props = $props();

  // Status is never conveyed by colour alone: icon + text label always ship.
  const TONES: Record<Status, { icon: typeof ShieldCheck; tone: string }> = {
    SAFE: { icon: ShieldCheck, tone: 'ok' },
    SUSPICIOUS: { icon: AlertTriangle, tone: 'warn' },
    TAMPERED: { icon: AlertOctagon, tone: 'err' },
    INSUFFICIENT: { icon: HelpCircle, tone: 'neutral' }
  };
  const config = $derived(TONES[status]);
</script>

<p class="badge {config.tone}">
  <span class="icon" aria-hidden="true"><config.icon size={18} /></span>
  <span>{statusLabel(status)}</span>
  {#if score !== null}
    <span class="score">Skor risiko {score} dari 100</span>
  {/if}
</p>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
    margin: 0;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    border: var(--border-width) solid currentColor;
    font-weight: 600;
    font-size: var(--text-sm);
  }
  .icon { display: inline-flex; }
  .ok { color: var(--ok); }
  .warn { color: var(--warn); }
  .err { color: var(--err); }
  .neutral { color: var(--text-2); }
  .score {
    color: var(--text-2);
    font-weight: 400;
    font-family: var(--font-mono);
    font-size: var(--text-xs);
  }
</style>
