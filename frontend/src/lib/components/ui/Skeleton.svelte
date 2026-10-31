<script lang="ts">
  interface Props {
    height?: string;
    lines?: number;
  }
  let { height = '1.5rem', lines = 1 }: Props = $props();
</script>

<div class="skeleton" aria-hidden="true">
  {#each Array.from({ length: lines }, (_, i) => i) as i (i)}
    <span style:height style:width={i === lines - 1 && lines > 1 ? '60%' : '100%'}></span>
  {/each}
</div>

<style>
  .skeleton {
    display: grid;
    gap: var(--space-2);
  }
  .skeleton span {
    display: block;
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    animation: pulse var(--dur-slow) var(--ease-in-out) infinite alternate;
  }
  @keyframes pulse {
    from { opacity: 0.55; }
    to { opacity: 1; }
  }
  @media (prefers-reduced-motion: reduce) {
    .skeleton span { animation: none; opacity: 0.7; }
  }
</style>
