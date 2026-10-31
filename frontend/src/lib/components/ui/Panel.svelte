<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title?: string;
    eyebrow?: string;
    children: Snippet;
    footer?: Snippet;
  }

  let { title, eyebrow, children, footer }: Props = $props();
</script>

<section class="panel">
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
    background: var(--surface);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5);
  }
  .eyebrow {
    font-family: var(--font-mono);
    font-size: var(--text-xs);
    letter-spacing: 0.06em;
    color: var(--text-2);
    margin: 0 0 var(--space-2);
  }
  h2 {
    font-size: var(--text-xl);
    margin: 0 0 var(--space-4);
  }
  .body { color: var(--text); }
  .footer {
    margin-top: var(--space-4);
    padding-top: var(--space-4);
    border-top: var(--border-width) solid var(--border);
  }
</style>
