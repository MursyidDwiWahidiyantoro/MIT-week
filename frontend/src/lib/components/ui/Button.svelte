<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    variant?: 'primary' | 'quiet' | 'danger' | 'ink' | 'lime';
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
  const classes = $derived(
    [
      'btn',
      variant === 'primary' ? 'btn--primary' : '',
      variant === 'quiet' ? 'btn--quiet' : '',
      variant === 'danger' ? 'btn--ink' : '',
      variant === 'ink' ? 'btn--ink' : '',
      variant === 'lime' ? 'btn--lime' : '',
      size === 'lg' ? 'btn--lg' : ''
    ]
      .filter(Boolean)
      .join(' ')
  );
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
  .btn--quiet {
    background: var(--white);
    color: var(--ink);
  }
  .btn--lg {
    min-height: 62px;
    padding: 0 var(--space-6);
    font-size: 1.15em;
  }
  .btn[aria-busy='true'] { cursor: progress; }
</style>
