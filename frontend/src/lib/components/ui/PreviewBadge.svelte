<script lang="ts">
  import { isMockMode } from '$lib/api/config';
  import Icon from './Icon.svelte';
</script>

{#if isMockMode}
  <p class="preview" role="note">
    <span class="icon" aria-hidden="true"><Icon name="sparkle" size={14} /></span>
    <span><strong>Mode pratinjau.</strong> Hasil dibuat di perangkat dari data contoh, bukan audit sungguhan.</span>
  </p>
{/if}

<style>
  .preview {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
    margin: 0;
    padding: 10px var(--gutter);
    background: var(--yellow);
    color: var(--on-fill);
    border-bottom: var(--border-width) solid var(--border);
    font-size: var(--text-sm);
    line-height: 1.4;
  }
  .icon { display: inline-flex; margin-top: 2px; flex: none; }
  strong { font-weight: 700; }
</style>
