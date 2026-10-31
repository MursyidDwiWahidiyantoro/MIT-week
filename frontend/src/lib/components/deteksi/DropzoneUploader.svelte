<script lang="ts">
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
      return 'Jenis berkas ini belum didukung. Gunakan JPG, PNG, atau WebP.';
    }
    if (file.size > MAX_BYTES) {
      return 'Berkas terlalu besar. Maksimal 10 MB.';
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
  aria-label="Pilih gambar atau lepas berkas di sini"
  onclick={() => document.getElementById('amba-file')?.click()}
  onkeydown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      document.getElementById('amba-file')?.click();
    }
  }}
  ondragover={(e) => { e.preventDefault(); dragging = true; }}
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
  <p class="title">{fileName ?? 'Pilih gambar atau lepas di sini'}</p>
  <p class="hint">JPG, PNG, atau WebP. Maksimal 10 MB.</p>
</div>

{#if error}
  <p class="error" role="alert">{error}</p>
{/if}

<style>
  .dropzone {
    display: grid;
    place-items: center;
    gap: var(--space-1);
    min-height: 140px;
    padding: var(--space-5);
    text-align: center;
    border: var(--border-width-strong) dashed var(--border-strong);
    border-radius: var(--radius-lg);
    background: var(--surface);
    cursor: pointer;
  }
  .dropzone:hover, .dropzone.dragging { background: var(--surface-2); }
  .dropzone:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .title { margin: 0; font-weight: 600; }
  .hint { margin: 0; color: var(--text-2); font-size: var(--text-sm); }
  .error { margin: var(--space-2) 0 0; color: var(--err); font-size: var(--text-sm); }
  .sr-only {
    position: absolute; width: 1px; height: 1px;
    overflow: hidden; clip-path: inset(50%);
  }
</style>
