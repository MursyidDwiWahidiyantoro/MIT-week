<script lang="ts">
  import type { Snippet } from 'svelte';

  /**
   * Renders below-the-fold content in its own task once the browser is idle,
   * instead of in the same long task as the first screen. Each section gets
   * a later slot, so the main thread stays responsive while the page fills
   * in (PRD §5 performance budget). Nothing above the fold should use this.
   */
  interface Props {
    children: Snippet;
    /** Later slots render later; keeps each mount in a separate task. */
    order?: number;
  }

  let { children, order = 0 }: Props = $props();
  let ready = $state(false);

  $effect(() => {
    let cancelled = false;
    const run = () => {
      if (!cancelled) ready = true;
    };
    const w = window as Window & {
      requestIdleCallback?: (cb: () => void, opts?: { timeout: number }) => number;
    };
    if (w.requestIdleCallback) w.requestIdleCallback(run, { timeout: 500 + order * 120 });
    else setTimeout(run, 40 + order * 40);
    return () => {
      cancelled = true;
    };
  });
</script>

{#if ready}{@render children()}{/if}
