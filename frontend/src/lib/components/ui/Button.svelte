<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    variant?: 'primary' | 'quiet' | 'danger';
    size?: 'md' | 'lg';
    type?: 'button' | 'submit';
    disabled?: boolean;
    loading?: boolean;
    href?: string;
    onclick?: (event: MouseEvent) => void;
    children: Snippet;
  }

  let {
    variant = 'primary',
    size = 'md',
    type = 'button',
    disabled = false,
    loading = false,
    href,
    onclick,
    children
  }: Props = $props();

  const inert = $derived(disabled || loading);
  const classes = $derived(`btn btn-${variant} btn-${size}${loading ? ' is-loading' : ''}`);
</script>

{#if href && !inert}
  <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- callers pass a resolved path -->
  <a class={classes} {href} aria-busy={loading}>{@render children()}</a>
{:else}
  <button class={classes} {type} disabled={inert} aria-busy={loading} {onclick}>
    {@render children()}
  </button>
{/if}

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    min-height: 44px;
    padding: 0 var(--space-5);
    border-radius: var(--radius-md);
    border: var(--border-width) solid transparent;
    font: inherit;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
    transition: background-color var(--dur-fast) var(--ease-out),
      border-color var(--dur-fast) var(--ease-out),
      transform var(--dur-fast) var(--ease-out);
  }
  .btn-lg { min-height: 52px; padding: 0 var(--space-6); font-size: var(--text-lg); }

  .btn-primary { background: var(--accent); color: #10161d; }
  .btn-primary:hover:not(:disabled) { background: color-mix(in srgb, var(--accent) 88%, white); }
  .btn-primary:active:not(:disabled) { transform: translateY(1px); }

  .btn-quiet {
    background: transparent;
    color: var(--text);
    border-color: var(--border-strong);
  }
  .btn-quiet:hover:not(:disabled) { background: var(--hover-overlay); }

  .btn-danger { background: var(--err); color: #14090a; }
  .btn-danger:hover:not(:disabled) { background: color-mix(in srgb, var(--err) 88%, white); }

  .btn:disabled {
    opacity: var(--disabled-opacity);
    cursor: not-allowed;
  }
  .btn.is-loading { cursor: progress; }
  .btn:focus-visible {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }
</style>
