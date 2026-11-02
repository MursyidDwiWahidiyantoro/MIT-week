<script lang="ts">
  import Button from '$lib/components/ui/Button.svelte';
  import Panel from '$lib/components/ui/Panel.svelte';
  import {
    evaluateQuickCheck,
    generateFamilyCode,
    quickCheckQuestions,
    type QuickCheckResult
  } from '$lib/quickcheck';
  import { Printer, RefreshCw, Copy, Check, KeyRound } from '@lucide/svelte';

  let answers = $state<Record<string, boolean>>({});
  let result = $state<QuickCheckResult | null>(null);
  let code = $state<string | null>(null);
  let copied = $state(false);

  const answered = $derived(quickCheckQuestions.every((q) => q.id in answers));

  function choose(id: string, value: boolean): void {
    const next = { ...answers, [id]: value };
    answers = next;
    // Six questions, six taps: the result appears as soon as the last one is
    // answered, with no extra submit step.
    result = quickCheckQuestions.every((q) => q.id in next) ? evaluateQuickCheck(next) : null;
  }

  function reset(): void {
    answers = {};
    result = null;
  }

  function makeCode(): void {
    code = generateFamilyCode();
    copied = false;
  }

  async function copyCode(): Promise<void> {
    if (!code) return;
    try {
      await navigator.clipboard.writeText(code);
      copied = true;
      setTimeout(() => (copied = false), 2500);
    } catch {
      // The code is visible and selectable; no error is needed.
    }
  }

  function printCard(): void {
    window.print();
  }
</script>

<svelte:head>
  <title>Cek cepat — AMBA</title>
  <meta
    name="description"
    content="Enam pertanyaan singkat untuk menilai situasi mencurigakan tanpa tautan atau gambar."
  />
</svelte:head>

<h1>Cek cepat</h1>
<p class="lede">
  Untuk pesan, telepon, atau situasi yang tidak punya tautan atau gambar. Enam
  pertanyaan, tanpa mengirim data ke mana pun.
</p>

<section class="questions no-print" aria-labelledby="questions-title">
  <h2 id="questions-title">Jawab sejujur mungkin</h2>
  <p class="hint">
    Pilih “Ya” atau “Tidak”. Hasil muncul setelah pertanyaan terakhir dijawab.
  </p>

  {#each quickCheckQuestions as question, i (question.id)}
    <fieldset>
      <legend><span class="num" aria-hidden="true">{i + 1}</span>{question.text}</legend>
      <div class="choices">
        <button
          type="button"
          class:selected={answers[question.id] === true}
          aria-pressed={answers[question.id] === true}
          onclick={() => choose(question.id, true)}
        >
          Ya
        </button>
        <button
          type="button"
          class:selected={answers[question.id] === false}
          aria-pressed={answers[question.id] === false}
          onclick={() => choose(question.id, false)}
        >
          Tidak
        </button>
      </div>
    </fieldset>
  {/each}

  {#if answered}
    <p class="answered" role="status">Semua pertanyaan sudah dijawab.</p>
  {/if}
</section>

{#if result}
  <Panel title="Hasil cek cepat">
    <div class="result" role="status" aria-live="polite">
      <p class="level" data-level={result.level}>{result.label}</p>
      <p class="summary">{result.summary}</p>
      {#if result.forcedByOtp}
        <p class="override">
          Karena ada permintaan OTP/PIN/kata sandi, hasil langsung Tinggi.
        </p>
      {/if}
      <h3>Langkah yang disarankan</h3>
      <ol class="steps">
        {#each result.steps as step (step)}<li>{step}</li>{/each}
      </ol>
      <p class="note">
        Ini bukan penilaian pasti. Jawaban hanya ada di halaman ini dan tidak
        dikirim ke server.
      </p>
      <Button variant="quiet" onclick={reset}>
        <RefreshCw size={16} aria-hidden="true" /> Ulangi
      </Button>
    </div>
  </Panel>
{/if}

<section class="family no-print" aria-labelledby="family-title">
  <h2 id="family-title">Sandi Keluarga</h2>
  <p class="lede">
    Sepakati dua kata rahasia dengan keluarga. Saat ada yang mengaku sebagai
    kerabat lewat nomor baru, minta sandi ini untuk memastikan.
  </p>

  <div class="family-actions">
    <Button onclick={makeCode}>
      <KeyRound size={16} aria-hidden="true" /> Buat sandi
    </Button>
    {#if code}
      <Button variant="quiet" onclick={copyCode}>
        {#if copied}<Check size={16} aria-hidden="true" /> Tersalin{:else}<Copy size={16} aria-hidden="true" /> Salin{/if}
      </Button>
      <Button variant="quiet" onclick={printCard}>
        <Printer size={16} aria-hidden="true" /> Cetak kartu
      </Button>
    {/if}
  </div>

  {#if code}
    <div class="print-card" data-testid="family-card">
      <p class="card-label">SANDI KELUARGA AMBA</p>
      <p class="card-code">{code}</p>
      <p class="card-copy">
        Jangan berikan sandi ini lewat pesan. Saat ada yang mengaku keluarga
        lewat nomor baru, telepon nomor lama dan minta sandi ini.
      </p>
      <p class="card-foot">Simpan di tempat yang hanya keluarga tahu.</p>
    </div>
    <p class="note no-print">
      Sandi dibuat di perangkat Anda dengan pengacak aman. Tidak ada yang dikirim
      atau disimpan di server.
    </p>
  {/if}
</section>

<style>
  h1 { font-size: var(--text-3xl); margin: 0 0 var(--space-3); }
  .lede { color: var(--text-2); max-width: var(--measure); margin: 0 0 var(--space-5); }
  h2 { font-size: var(--text-2xl); margin: var(--space-7) 0 var(--space-3); }
  h3 { font-size: var(--text-lg); margin: var(--space-5) 0 var(--space-2); }
  .hint, .note { color: var(--text-2); font-size: var(--text-sm); max-width: var(--measure); }
  .hint { margin: 0 0 var(--space-4); }

  fieldset {
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-4);
    margin: 0 0 var(--space-3);
  }
  legend { padding: 0 var(--space-2); font-weight: 600; max-width: var(--measure); }
  .num {
    display: inline-grid; place-items: center; width: 24px; height: 24px;
    margin-right: var(--space-2);
    border: var(--border-width) solid var(--border-strong); border-radius: 50%;
    font-family: var(--font-mono); font-size: var(--text-xs); font-weight: 400;
  }
  .choices { display: flex; gap: var(--space-2); margin-top: var(--space-3); }
  .choices button {
    min-width: 88px; min-height: 44px; padding: 0 var(--space-4);
    background: var(--surface-2); color: var(--text);
    border: var(--border-width) solid var(--border-strong); border-radius: var(--radius-md);
    font: inherit; cursor: pointer;
  }
  .choices button:hover { background: var(--hover-overlay); }
  .choices button.selected { background: var(--selected-bg); border-color: var(--accent); font-weight: 600; }
  .choices button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .answered { color: var(--text-2); font-size: var(--text-sm); }

  .result { display: grid; gap: var(--space-2); justify-items: start; }
  .level {
    margin: 0; padding: var(--space-1) var(--space-4);
    border: var(--border-width) solid currentColor; border-radius: var(--radius-md);
    font-family: var(--font-mono); font-size: var(--text-sm); font-weight: 600;
  }
  .level[data-level='high'] { color: var(--err); }
  .level[data-level='wary'] { color: var(--warn); }
  .level[data-level='low'] { color: var(--ok); }
  .summary { margin: 0; font-size: var(--text-lg); max-width: var(--measure); }
  .override { margin: 0; color: var(--warn); font-size: var(--text-sm); max-width: var(--measure); }
  .steps { margin: 0; padding-left: var(--space-5); display: grid; gap: var(--space-2); max-width: var(--measure); }

  .family { border-top: var(--border-width) solid var(--border); margin-top: var(--space-7); }
  .family-actions { display: flex; flex-wrap: wrap; gap: var(--space-3); margin-bottom: var(--space-4); }

  .print-card {
    border: var(--border-width-strong) solid var(--border-strong);
    border-radius: var(--radius-md);
    padding: var(--space-5);
    max-width: 28rem;
    background: var(--surface);
  }
  .card-label { margin: 0; font-family: var(--font-mono); font-size: var(--text-xs); letter-spacing: 0.08em; color: var(--text-2); }
  .card-code { margin: var(--space-3) 0; font-family: var(--font-mono); font-size: var(--text-2xl); font-weight: 700; overflow-wrap: anywhere; }
  .card-copy { margin: 0 0 var(--space-3); max-width: var(--measure); }
  .card-foot { margin: 0; color: var(--text-3); font-size: var(--text-sm); }

  @media print {
    :global(.site-header), :global(.site-footer), :global(.skip-link) { display: none !important; }
    :global(body) { background: #fff; color: #000; }
    :global(main) { padding: 0; max-width: none; }
    h1, .lede, .no-print { display: none !important; }
    .family { border: 0; margin: 0; }
    .print-card { border: 2px solid #000; background: #fff; color: #000; max-width: none; }
    .card-label, .card-foot { color: #333; }
  }
</style>
