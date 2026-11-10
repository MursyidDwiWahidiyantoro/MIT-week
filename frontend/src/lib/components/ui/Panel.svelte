<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title?: string;
    /** Colour role of the panel surface. `white` is the default translucent pane. */
    tone?: 'white' | 'paper' | 'ink' | 'safe' | 'yellow' | 'coral' | 'indigo' | 'soft';
    children: Snippet;
    footer?: Snippet;
  }

  let { title, tone = 'white', children, footer }: Props = $props();

  const TONE_CLASS: Record<NonNullable<Props['tone']>, string> = {
    white: '',
    paper: 'panel--paper',
    ink: 'panel--ink',
    safe: 'panel--safe',
    yellow: 'panel--yellow',
    coral: 'panel--coral',
    indigo: 'panel--indigo',
    soft: 'panel--soft'
  };
</script>

<section class="panel {TONE_CLASS[tone]}">
  {#if title}
    <h2>{title}</h2>
  {/if}
  <div class="body">{@render children()}</div>
  {#if footer}
    <div class="footer">{@render footer()}</div>
  {/if}
</section>

<style>
  /* A secondary pane: translucent over the page light, no blur. The frosted
   * glass is kept for surfaces with content under them. */
  .panel {
    background: var(--glass-bg);
    border: 1px solid var(--glass-edge);
    border-radius: var(--radius-lg);
    box-shadow: var(--glass-highlight), var(--shadow-1);
    padding: clamp(1.25rem, 3vw, 1.85rem);
    min-width: 0;
    margin-bottom: var(--space-5);
  }
  .panel--paper { background: color-mix(in srgb, var(--paper) 70%, transparent); }
  .panel--ink { background: var(--panel-dark); color: var(--on-panel-dark); border-color: transparent; }
  .panel--safe { background: var(--safe); color: var(--on-fill); }
  .panel--yellow { background: var(--yellow); color: var(--on-fill); }
  .panel--coral { background: var(--coral); color: var(--on-fill); }
  .panel--indigo { background: var(--indigo-soft); color: var(--ink); }
  .panel--soft { background: color-mix(in srgb, var(--indigo-soft) 70%, transparent); color: var(--ink); }

  h2 { font-size: 1.4em; margin: 0 0 var(--space-4); }
  .body { min-width: 0; }
  .footer {
    margin-top: var(--space-5);
    padding-top: var(--space-4);
    border-top: 1px solid var(--glass-edge);
  }
  .panel--ink :global(a) { color: var(--on-panel-dark); }
</style>
