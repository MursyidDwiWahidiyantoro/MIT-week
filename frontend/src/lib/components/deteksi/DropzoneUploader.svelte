<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  const MAX_BYTES = 10 * 1024 * 1024;
  const ACCEPTED = ['image/jpeg', 'image/png', 'image/webp'];

  interface Props {
    onselect: (file: File | null) => void;
  }

  let { onselect }: Props = $props();

  let error = $state<string | null>(null);
  let fileName = $state<string | null>(null);
  let dragging = $state(false);

  function validate(file: File): string | null {
    if (!ACCEPTED.includes(file.type)) {
      return t('drop.unsupported');
    }
    if (file.size > MAX_BYTES) {
      return t('drop.tooLarge');
    }
    return null;
  }

  function accept(file: File | undefined): void {
    if (!file) return;
    const problem = validate(file);
    if (problem) {
      error = problem;
      fileName = null;
      onselect(null);
      return;
    }
    error = null;
    fileName = file.name;
    onselect(file);
  }
</script>

<div
  class="dropzone"
  class:dragging
  role="button"
  tabindex="0"
  aria-label={t('drop.aria')}
  onclick={() => document.getElementById('amba-file')?.click()}
  onkeydown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      document.getElementById('amba-file')?.click();
    }
  }}
  ondragover={(e) => {
    e.preventDefault();
    dragging = true;
  }}
  ondragleave={() => (dragging = false)}
  ondrop={(e) => {
    e.preventDefault();
    dragging = false;
    accept(e.dataTransfer?.files?.[0]);
  }}
>
  <input
    id="amba-file"
    class="sr-only"
    type="file"
    accept="image/jpeg,image/png,image/webp"
    onchange={(e) => accept(e.currentTarget.files?.[0])}
  />
  <span class="icon" aria-hidden="true"><Icon name="image" size={34} /></span>
  <p class="title">{fileName ?? t('drop.title')}</p>
  <p class="hint">{t('drop.hint')}</p>
</div>

{#if error}
  <p class="error-text" role="alert">{error}</p>
{/if}

<style>
  .dropzone {
    display: grid;
    place-items: center;
    gap: var(--space-2);
    min-height: 180px;
    padding: var(--space-6) var(--space-5);
    text-align: center;
    border: var(--border-width-strong) dashed var(--border);
    border-radius: var(--radius-lg);
    background: var(--white);
    box-shadow: var(--shadow-hard);
    cursor: pointer;
    transition: background-color var(--dur-fast) var(--ease-out);
  }
  .dropzone:hover,
  .dropzone.dragging { background: var(--indigo-tint); }
  .dropzone:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .icon {
    display: grid;
    place-items: center;
    width: 64px;
    height: 64px;
    background: var(--indigo-soft);
    border: var(--border-width) solid var(--border);
    border-radius: 20px;
  }
  .title { margin: 0; font-family: var(--font-display); font-weight: 800; font-size: 1.15em; }
  .hint { margin: 0; color: var(--muted); font-size: var(--text-sm); }
</style>
