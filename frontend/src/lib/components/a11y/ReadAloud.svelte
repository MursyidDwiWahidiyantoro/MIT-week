<script lang="ts">
  import { speechAvailable, speak, stopSpeaking } from '$lib/a11y.svelte';
  import { Volume2, Square } from '@lucide/svelte';

  interface Props {
    text: string;
  }

  let { text }: Props = $props();

  const supported = $derived(speechAvailable());
  let speaking = $state(false);

  function toggle(): void {
    if (speaking) {
      stopSpeaking();
      speaking = false;
      return;
    }
    speak(text);
    speaking = true;
  }

  // Stop speech when the component goes away, so navigating never leaves a
  // voice reading a page the user has left.
  $effect(() => () => stopSpeaking());
</script>

{#if supported}
  <button type="button" class="read-aloud" onclick={toggle} aria-pressed={speaking}>
    {#if speaking}
      <Square size={16} aria-hidden="true" /> Hentikan bacaan
    {:else}
      <Volume2 size={16} aria-hidden="true" /> Bacakan hasil
    {/if}
  </button>
{/if}

<style>
  .read-aloud {
    display: inline-flex; align-items: center; gap: var(--space-2);
    min-height: 44px; padding: 0 var(--space-4);
    background: transparent; color: var(--text-2);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-md); font: inherit; font-size: var(--text-sm);
    cursor: pointer;
  }
  .read-aloud:hover { color: var(--text); background: var(--hover-overlay); }
  .read-aloud:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
