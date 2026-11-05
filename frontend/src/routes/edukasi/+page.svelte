<script lang="ts">
  import Button from '$lib/components/ui/Button.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  const modus = [
    {
      id: 'bansos',
      tone: 'yellow',
      icon: 'alert',
      title: 'Bantuan sosial palsu',
      body: 'Akun meniru program bantuan resmi, meminta Anda mengisi data dan menyebarkan tautan ke grup lain.'
    },
    {
      id: 'judul',
      tone: 'soft',
      icon: 'image',
      title: 'Judul berita yang diedit',
      body: 'Gambar tangkapan layar judul berita diubah agar terbaca lebih mengejutkan daripada artikel aslinya.'
    },
    {
      id: 'shortlink',
      tone: 'lime',
      icon: 'link',
      title: 'Tautan pendek ke halaman palsu',
      body: 'Tautan dipendekkan supaya alamat aslinya tidak terlihat, lalu diarahkan ke halaman yang meminta kata sandi.'
    }
  ] as const;

  const questions = [
    {
      id: 'q1',
      question: 'Gambar terlihat punya area lebih terang setelah diperiksa. Apa artinya?',
      options: [
        { id: 'a', text: 'Gambar pasti palsu.', correct: false },
        { id: 'b', text: 'Ada indikasi area itu pernah diubah. Perlu dicek lagi.', correct: true },
        { id: 'c', text: 'Gambar pasti asli.', correct: false }
      ],
      explain:
        'Area terang hanya menunjukkan perbedaan hasil kompresi. Itu petunjuk, bukan bukti. Memotong atau menyimpan ulang gambar juga bisa menimbulkannya.'
    },
    {
      id: 'q2',
      question: 'Mengapa analisis gambar melemah pada tangkapan layar?',
      options: [
        { id: 'a', text: 'Karena tangkapan layar sudah disimpan ulang oleh ponsel, sehingga perbedaannya tersebar.', correct: true },
        { id: 'b', text: 'Karena tangkapan layar selalu lebih kecil.', correct: false },
        { id: 'c', text: 'Karena tangkapan layar tidak bisa dibaca komputer.', correct: false }
      ],
      explain:
        'Tangkapan layar melewati proses penyimpanan ulang, jadi seluruh gambar sudah berubah. Perubahan suntingan jadi lebih sulit dibedakan.'
    },
    {
      id: 'q3',
      question: 'Pesan meminta Anda menyebarkan ke 10 grup hari ini juga. Tindakan paling aman?',
      options: [
        { id: 'a', text: 'Sebarkan dulu, nanti dicek.', correct: false },
        { id: 'b', text: 'Cek sumber resminya, jangan sebarkan sebelum jelas.', correct: true },
        { id: 'c', text: 'Balas dengan tuduhan bahwa pengirimnya penipu.', correct: false }
      ],
      explain:
        'Tekanan waktu adalah taktik yang umum. Menahan diri dan memeriksa sumber resmi adalah langkah paling aman.'
    }
  ];

  let answers = $state<Record<string, string>>({});
  let checked = $state(false);

  const score = $derived(
    questions.filter((q) => q.options.find((o) => o.id === answers[q.id])?.correct).length
  );

  function reset(): void {
    answers = {};
    checked = false;
  }
</script>

<svelte:head>
  <title>Belajar — AMBA</title>
  <meta
    name="description"
    content="Tiga modus penipuan yang sering muncul dan kuis singkat untuk melatih kepekaan. Tanpa nilai yang disimpan."
  />
</svelte:head>

<section class="head">
  <p class="eyebrow">Sekolah scam</p>
  <h1 class="h1">Belajar mengenali tanda bahaya</h1>
  <p class="lede">
    Tiga modus yang sering muncul, lalu tiga pertanyaan singkat. Tidak ada nilai yang disimpan.
  </p>
</section>

<section aria-labelledby="modus-title">
  <h2 class="h2" id="modus-title">Modus yang sering terjadi</h2>
  <div class="grid">
    {#each modus as m (m.id)}
      <article class="mode card lift mode--{m.tone}">
        <span class="icon" aria-hidden="true"><Icon name={m.icon} size={28} /></span>
        <h3>{m.title}</h3>
        <p>{m.body}</p>
      </article>
    {/each}
  </div>
</section>

<section class="quiz" aria-labelledby="kuis-title">
  <h2 class="h2" id="kuis-title">Kuis singkat</h2>
  <p class="quiz-lede">Pilih satu jawaban per pertanyaan, lalu periksa di akhir.</p>

  {#each questions as q, i (q.id)}
    <div class="q card lift" role="radiogroup" aria-labelledby={`q-label-${q.id}`}>
      <p class="q-legend" id={`q-label-${q.id}`}>
        <span class="q-num" aria-hidden="true">{i + 1}</span>
        {q.question}
      </p>
      {#each q.options as option (option.id)}
        <label
          class="option"
          class:correct={checked && option.correct}
          class:wrong={checked && answers[q.id] === option.id && !option.correct}
        >
          <input type="radio" name={q.id} value={option.id} bind:group={answers[q.id]} disabled={checked} />
          <span>{option.text}</span>
          {#if checked && option.correct}
            <span class="verdict-tag">Jawaban tepat</span>
          {:else if checked && answers[q.id] === option.id}
            <span class="verdict-tag">Kurang tepat</span>
          {/if}
        </label>
      {/each}
      {#if checked}
        <p class="explain">{q.explain}</p>
      {/if}
    </div>
  {/each}

  <div class="actions">
    {#if !checked}
      <Button onclick={() => (checked = true)} disabled={Object.keys(answers).length < questions.length}>
        Periksa jawaban
      </Button>
    {:else}
      <p class="result" role="status">
        <span class="result-num">{score}<span class="result-total">/{questions.length}</span></span>
        benar.
      </p>
      <Button variant="quiet" onclick={reset}>
        <Icon name="refresh" size={18} /> Ulangi
      </Button>
    {/if}
  </div>
</section>

<style>
  .head { max-width: 54rem; margin-bottom: var(--space-7); }
  h1 { margin-bottom: var(--space-3); }
  .grid { display: grid; grid-template-columns: 1fr; gap: var(--space-5); margin-top: var(--space-5); }
  @media (min-width: 48rem) { .grid { grid-template-columns: repeat(3, 1fr); } }
  .mode { min-width: 0; padding: clamp(1.25rem, 3vw, 1.75rem); display: flex; flex-direction: column; height: 100%; }
  .mode--yellow { background: var(--yellow); }
  .mode--soft { background: var(--indigo-soft); }
  .mode--lime { background: var(--lime); }
  .icon {
    display: grid;
    place-items: center;
    width: 52px;
    height: 52px;
    margin-bottom: var(--space-4);
    background: var(--white);
    border: var(--border-width) solid var(--border);
    border-radius: 16px;
    box-shadow: var(--shadow-hard-xs);
  }
  h3 { font-size: 1.35em; margin: 0 0 var(--space-2); }
  .mode p { margin: 0; font-size: 0.98em; line-height: 1.45; }

  .quiz { margin-top: var(--space-9); }
  .quiz-lede { margin: var(--space-3) 0 var(--space-5); color: var(--body-2); max-width: var(--measure); }
  .q { margin: 0 0 var(--space-4); padding: clamp(1.1rem, 3vw, 1.6rem); }
  /* Desktop: question on the left, its three answers stacked on the right, so
   * the card carries its full width instead of trailing off into paper. */
  @media (min-width: 64rem) {
    .q {
      display: grid;
      grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr);
      column-gap: clamp(2rem, 4vw, 4rem);
      align-items: start;
    }
    .q-legend { grid-column: 1; margin-bottom: 0; }
    .q :global(.option) { grid-column: 2; }
    .explain { grid-column: 1 / -1; }
  }
  .q-legend {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: 0;
    margin: 0 0 var(--space-4);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.15em;
    line-height: 1.3;
  }
  .q-num {
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    flex: none;
    background: var(--ink);
    color: var(--on-ink);
    border-radius: 10px;
    font-size: 0.9em;
  }
  .option {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-height: 48px;
    padding: var(--space-3);
    border: 2px solid transparent;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition:
      background-color var(--dur-fast) var(--ease-out),
      border-color var(--dur-fast) var(--ease-out),
      transform var(--dur-fast) var(--ease-out);
  }
  .option:hover { background: var(--indigo-tint); border-color: var(--ink); transform: translateX(3px); }
  .option.correct:hover { background: var(--lime); }
  .option.wrong:hover { background: var(--coral-tint); }
  .option input { width: 22px; height: 22px; accent-color: var(--indigo); flex: none; }
  .option.correct { background: var(--lime); border-color: var(--on-fill); }
  .option.wrong { background: var(--coral-tint); border-color: var(--ink); }
  .verdict-tag {
    margin-left: auto;
    padding: 2px 10px;
    border: 2px solid var(--border);
    border-radius: var(--radius-pill);
    font-size: var(--text-xs);
    font-weight: 700;
    white-space: nowrap;
  }
  .explain {
    margin: var(--space-3) 0 0;
    padding: var(--space-3) var(--space-4);
    background: var(--paper);
    border-left: 6px solid var(--indigo);
    border-radius: var(--radius-sm);
    color: var(--body-2);
    font-size: var(--text-base);
    max-width: var(--measure);
  }
  .actions { display: flex; align-items: center; gap: var(--space-4); flex-wrap: wrap; margin-top: var(--space-5); }
  @media (max-width: 44rem) { .actions { flex-direction: column; align-items: stretch; } .actions :global(.btn) { width: 100%; justify-content: center; } }
  .result { margin: 0; font-weight: 700; }
  .result-num {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 2.4em;
    letter-spacing: -0.03em;
    margin-right: var(--space-2);
  }
  .result-total { font-size: 0.5em; color: var(--muted); }
</style>
