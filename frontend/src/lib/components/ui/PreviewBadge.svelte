<script lang="ts">
  import { isMockMode } from '$lib/api/config';
  import { FlaskConical } from '@lucide/svelte';
</script>

{#if isMockMode}
  <p class="preview" role="note">
    <span class="icon" aria-hidden="true"><FlaskConical size={14} /></span>
    <span><strong>Mode pratinjau.</strong> Hasil dibuat di perangkat dari data contoh, bukan audit sungguhan.</span>
  </p>
{/if}

<style>
  .preview {
    display: flex; align-items: flex-start; gap: var(--space-2);
    margin: 0; padding: var(--space-2) var(--space-4);
    border-bottom: var(--border-width) solid var(--border);
    background: var(--surface);
    color: var(--text-2); font-size: var(--text-sm);
  }
  .icon { display: inline-flex; margin-top: 2px; color: var(--warn); flex: none; }
  strong { color: var(--text); }
</style>
