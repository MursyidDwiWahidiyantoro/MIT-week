<script lang="ts">
  import Button from '$lib/components/ui/Button.svelte';
  import {
    evaluateQuickCheck,
    generateFamilyCode,
    quickCheckQuestions,
    type QuickCheckResult
  } from '$lib/quickcheck';
  import Icon from '$lib/components/ui/Icon.svelte';

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

<section class="head">
  <p class="eyebrow">Tanpa tautan</p>
  <h1 class="h1">Cek cepat</h1>
  <p class="lede">
    Untuk pesan, telepon, atau situasi yang tidak punya tautan atau gambar. Enam
    pertanyaan, tanpa mengirim data ke mana pun.
  </p>
</section>

<section class="questions no-print" aria-labelledby="questions-title">
  <h2 class="h2" id="questions-title">Jawab sejujur mungkin</h2>
  <p class="hint">Pilih “Ya” atau “Tidak”. Hasil muncul setelah pertanyaan terakhir dijawab.</p>

  {#each quickCheckQuestions as question, i (question.id)}
    <div class="q lift" role="group" aria-labelledby={`qc-label-${question.id}`}>
      <p class="q-legend" id={`qc-label-${question.id}`}>
        <span class="q-num" aria-hidden="true">{i + 1}</span>
        {question.text}
      </p>
      <div class="choices">
        <button
          type="button"
          class="choice"
          class:selected={answers[question.id] === true}
          aria-pressed={answers[question.id] === true}
          onclick={() => choose(question.id, true)}
        >
          Ya
        </button>
        <button
          type="button"
          class="choice"
          class:selected={answers[question.id] === false}
          aria-pressed={answers[question.id] === false}
          onclick={() => choose(question.id, false)}
        >
          Tidak
        </button>
      </div>
    </div>
  {/each}

  {#if answered}
    <p class="answered" role="status">Semua pertanyaan sudah dijawab.</p>
  {/if}
</section>

{#if result}
  <section class="result-panel card lift" aria-labelledby="result-title">
    <h2 id="result-title" class="h2">Hasil cek cepat</h2>
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
        <Icon name="refresh" size={18} /> Ulangi
      </Button>
    </div>
  </section>
{/if}

<section class="family no-print" aria-labelledby="family-title">
  <div class="family-panel card--soft lift">
    <h2 class="h2" id="family-title">Sandi Keluarga</h2>
    <p class="family-lede">
      Sepakati dua kata rahasia dengan keluarga. Saat ada yang mengaku sebagai
      kerabat lewat nomor baru, minta sandi ini untuk memastikan.
    </p>

    <div class="family-actions">
      <Button onclick={makeCode}>
        <Icon name="lock" size={18} /> Buat sandi
      </Button>
      {#if code}
        <Button variant="quiet" onclick={copyCode}>
          <Icon name={copied ? 'check' : 'copy'} size={18} /> {copied ? 'Tersalin' : 'Salin'}
        </Button>
        <Button variant="quiet" onclick={printCard}>
          <Icon name="print" size={18} /> Cetak kartu
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
  </div>
</section>

<style>
  .head { max-width: 54rem; margin-bottom: var(--space-6); }
  h1 { margin-bottom: var(--space-3); }
  .hint, .note { color: var(--body-2); font-size: var(--text-sm); max-width: var(--measure); }
  .hint { margin: var(--space-3) 0 var(--space-5); }

  .q {
    margin: 0 0 var(--space-3);
    padding: var(--space-4) var(--space-5);
    background: var(--white);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-hard);
  }
  .q-legend {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: 0;
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.1em;
    line-height: 1.35;
    max-width: 46rem;
  }
  .q-num {
    display: grid;
    place-items: center;
    width: 30px;
    height: 30px;
    flex: none;
    background: var(--ink);
    color: var(--lime);
    border-radius: 10px;
    font-size: 0.85em;
  }
  .choices { display: flex; gap: var(--space-2); margin-top: var(--space-4); }
  .choice {
    min-width: 96px;
    min-height: 48px;
    padding: 0 var(--space-5);
    background: var(--white);
    color: var(--ink);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-hard-xs);
    font: inherit;
    font-weight: 700;
    cursor: pointer;
    transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
  }
  .choice:hover { transform: translate(-2px, -2px); box-shadow: 5px 5px 0 var(--ink); }
  .choice.selected { background: var(--indigo); color: var(--white); }
  .choice.selected:hover { background: var(--indigo); box-shadow: 5px 5px 0 var(--ink); }
  .choice:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  @media (max-width: 36rem) {
    .choices { display: flex; width: 100%; }
    .choice { flex: 1 1 0; min-width: 0; text-align: center; }
  }
  .answered { color: var(--body-2); font-weight: 700; font-size: var(--text-sm); }

  .result-panel { padding: clamp(1.25rem, 3vw, 2rem); margin-top: var(--space-5); }
  .result { display: grid; gap: var(--space-3); justify-items: start; }
  .level {
    margin: 0;
    padding: 6px 18px;
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-pill);
    box-shadow: var(--shadow-hard-xs);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.15em;
    color: var(--ink);
  }
  .level[data-level='high'] { background: var(--coral); }
  .level[data-level='wary'] { background: var(--yellow); }
  .level[data-level='low'] { background: var(--lime); }
  .summary {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: clamp(1.2rem, 2.6vw, 1.55rem);
    line-height: 1.25;
    max-width: 32ch;
  }
  .override { margin: 0; color: var(--err); font-weight: 700; font-size: var(--text-sm); max-width: var(--measure); }
  .result h3 { font-size: 1.15em; margin: var(--space-2) 0 0; }
  .steps { margin: 0; padding-left: var(--space-5); display: grid; gap: var(--space-2); max-width: var(--measure); }

  .family { margin-top: var(--space-8); }
  .family-panel {
    padding: clamp(1.5rem, 4vw, 2.25rem);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-hard-lg);
  }
  .family-lede { margin: var(--space-3) 0 0; max-width: var(--measure); }
  .family-actions { display: flex; flex-wrap: wrap; gap: var(--space-3); margin: var(--space-5) 0 var(--space-4); }

  .print-card {
    margin-top: var(--space-4);
    padding: var(--space-6);
    max-width: 30rem;
    background: var(--white);
    border: var(--border-width-strong) solid var(--ink);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-hard);
  }
  .card-label {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: var(--text-xs);
    letter-spacing: 0.1em;
    color: var(--indigo-text);
  }
  .card-code {
    margin: var(--space-3) 0;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: clamp(1.5rem, 5vw, 2.25rem);
    letter-spacing: 0.02em;
    overflow-wrap: anywhere;
  }
  .card-copy { margin: 0 0 var(--space-3); max-width: var(--measure); }
  .card-foot { margin: 0; color: var(--muted); font-size: var(--text-sm); }

  @media print {
    :global(.site-header), :global(.site-footer), :global(.skip-link) { display: none !important; }
    :global(body) { background: #fff; color: #000; }
    :global(main) { padding: 0; max-width: none; }
    h1, .lede, .eyebrow, .no-print { display: none !important; }
    .family { border: 0; margin: 0; }
    .family-panel { border: 0; box-shadow: none; padding: 0; }
    .print-card { border: 3px solid #000; background: #fff; color: #000; max-width: none; box-shadow: none; }
    .card-label, .card-foot { color: #333; }
  }
</style>
