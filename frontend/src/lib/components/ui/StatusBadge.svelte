<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import Icon, { type IconName } from './Icon.svelte';
  import type { Status } from '$lib/api/types';
  import { statusLabel } from '$lib/format';

  interface Props {
    status: Status;
    score?: number | null;
  }

  let { status, score = null }: Props = $props();

  /* Status is never conveyed by colour alone: icon + text label always ship.
   * Tones are the saturated sticker fills, with ink text for AA contrast. */
  const TONES: Record<Status, { icon: IconName; tone: string }> = {
    SAFE: { icon: 'shield-check', tone: 'tone--safe' },
    SUSPICIOUS: { icon: 'alert', tone: 'tone--warn' },
    TAMPERED: { icon: 'alert', tone: 'tone--danger' },
    INSUFFICIENT: { icon: 'chat', tone: 'tone--neutral' }
  };
  const config = $derived(TONES[status]);
</script>

<p class="badge {config.tone}">
  <span class="icon" aria-hidden="true"><Icon name={config.icon} size={20} /></span>
  <span class="label">{statusLabel(status)}</span>
  {#if score !== null}
    <span class="score">{t('score.of100', { score })}</span>
  {/if}
</p>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
    margin: 0;
    padding: 6px 14px;
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-pill);
    box-shadow: var(--shadow-hard-xs);
    color: var(--ink);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: var(--text-base);
    line-height: 1.25;
  }
  .icon { display: inline-flex; }
  .tone--safe { background: var(--safe); }
  .tone--warn { background: var(--yellow); }
  .tone--danger { background: var(--coral); }
  .tone--neutral { background: var(--indigo-soft); }
  .score {
    font-family: var(--font-sans);
    font-weight: 700;
    font-size: var(--text-sm);
  }
</style>
