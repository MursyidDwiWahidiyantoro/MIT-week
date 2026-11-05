<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title?: string;
    eyebrow?: string;
    /** Colour role of the panel surface. `white` is the default card. */
    tone?: 'white' | 'paper' | 'ink' | 'lime' | 'yellow' | 'coral' | 'indigo' | 'soft';
    children: Snippet;
    footer?: Snippet;
  }

  let { title, eyebrow, tone = 'white', children, footer }: Props = $props();

  const TONE_CLASS: Record<NonNullable<Props['tone']>, string> = {
    white: '',
    paper: 'panel--paper',
    ink: 'panel--ink',
    lime: 'panel--lime',
    yellow: 'panel--yellow',
    coral: 'panel--coral',
    indigo: 'panel--indigo',
    soft: 'panel--soft'
  };
</script>

<section class="panel lift {TONE_CLASS[tone]}">
  {#if eyebrow}
    <p class="eyebrow">{eyebrow}</p>
  {/if}
  {#if title}
    <h2>{title}</h2>
  {/if}
  <div class="body">{@render children()}</div>
  {#if footer}
    <div class="footer">{@render footer()}</div>
  {/if}
</section>

<style>
  .panel {
    background: var(--white);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-hard-lg);
    padding: clamp(1.25rem, 3vw, 1.9rem);
    min-width: 0;
    margin-bottom: var(--space-5);
  }
  .panel--paper { background: var(--paper); }
  .panel--ink { background: var(--ink); color: var(--white); box-shadow: var(--shadow-indigo); }
  .panel--lime { background: var(--lime); color: var(--ink); }
  .panel--yellow { background: var(--yellow); color: var(--ink); }
  .panel--coral { background: var(--coral); color: var(--ink); }
  .panel--indigo { background: var(--indigo); color: var(--white); box-shadow: var(--shadow-hard-xl); }
  .panel--soft { background: var(--indigo-soft); color: var(--ink); }

  h2 { font-size: 1.45em; margin: 0 0 var(--space-4); }
  .body { min-width: 0; }
  .footer {
    margin-top: var(--space-5);
    padding-top: var(--space-4);
    border-top: var(--border-width) solid currentColor;
  }
  .panel--ink :global(a),
  .panel--indigo :global(a) { color: var(--lime); }
</style>
