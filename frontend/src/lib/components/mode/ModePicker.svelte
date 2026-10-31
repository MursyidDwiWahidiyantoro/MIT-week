<script lang="ts">
  import { mode, type Mode } from '$lib/mode.svelte';

  const options: Array<{ value: Mode; title: string; detail: string }> = [
    { value: 'explore', title: 'Jelajah', detail: 'Lebih banyak detail visual dan animasi.' },
    { value: 'essential', title: 'Esensial', detail: 'Lebih tenang, teks lebih besar, gerak minimal.' }
  ];
</script>

<div class="picker" role="dialog" aria-labelledby="mode-picker-title" aria-modal="false">
  <h2 id="mode-picker-title">Pilih tampilan yang paling nyaman</h2>
  <p>Keduanya menampilkan informasi dan fungsi yang sama.</p>
  <div class="options">
    {#each options as option (option.value)}
      <button type="button" onclick={() => mode.set(option.value)}>
        <strong>{option.title}</strong>
        <span>{option.detail}</span>
      </button>
    {/each}
  </div>
  <button class="later" type="button" onclick={() => mode.dismiss()}>
    Pakai pengaturan saat ini
  </button>
</div>

<style>
  .picker {
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
    padding: var(--space-5);
    max-width: 34rem;
  }
  h2 { margin: 0 0 var(--space-2); font-size: var(--text-xl); }
  p { margin: 0 0 var(--space-4); color: var(--text-2); }
  .options { display: grid; gap: var(--space-3); }
  @media (min-width: 40rem) {
    .options { grid-template-columns: 1fr 1fr; }
  }
  button {
    display: grid;
    gap: var(--space-1);
    text-align: left;
    min-height: 44px;
    padding: var(--space-4);
    background: var(--surface-2);
    color: var(--text);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-md);
    font: inherit;
    cursor: pointer;
  }
  button:hover { background: var(--hover-overlay); }
  button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  span { color: var(--text-2); font-size: var(--text-sm); }
  .later {
    margin-top: var(--space-4);
    min-height: 44px;
    padding: 0 var(--space-4);
    background: transparent;
    color: var(--text-2);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    font: inherit;
    cursor: pointer;
  }
  .later:hover { color: var(--text); background: var(--hover-overlay); }
  .later:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
</style>
