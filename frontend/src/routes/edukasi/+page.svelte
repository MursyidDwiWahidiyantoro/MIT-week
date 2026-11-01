<script lang="ts">
  import Panel from '$lib/components/ui/Panel.svelte';
  import Button from '$lib/components/ui/Button.svelte';

  const modus = [
    {
      id: 'bansos',
      title: 'Bantuan sosial palsu di media sosial',
      body: 'Akun meniru program bantuan resmi, meminta Anda mengisi data dan menyebarkan tautan ke grup lain.'
    },
    {
      id: 'judul',
      title: 'Judul berita yang diedit',
      body: 'Gambar tangkapan layar judul berita diubah agar terbaca lebih mengejutkan daripada artikel aslinya.'
    },
    {
      id: 'shortlink',
      title: 'Tautan pendek menuju halaman palsu',
      body: 'Tautan dipendekkan supaya alamat aslinya tidak terlihat, lalu diarahkan ke halaman yang meminta kata sandi.'
    }
  ];

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

<svelte:head><title>Edukasi — AMBA</title></svelte:head>

<h1>Belajar mengenali tanda bahaya</h1>
<p class="lede">
  Tiga modus yang sering muncul, lalu tiga pertanyaan singkat. Tidak ada nilai yang
  disimpan.
</p>

<section aria-labelledby="modus-title">
  <h2 id="modus-title">Modus yang sering terjadi</h2>
  <div class="grid">
    {#each modus as m (m.id)}
      <Panel title={m.title}>
        <p>{m.body}</p>
      </Panel>
    {/each}
  </div>
</section>

<section aria-labelledby="kuis-title" class="quiz">
  <h2 id="kuis-title">Kuis singkat</h2>

  {#each questions as q, i (q.id)}
    <fieldset>
      <legend>{i + 1}. {q.question}</legend>
      {#each q.options as option (option.id)}
        <label class="option" class:correct={checked && option.correct} class:wrong={checked && answers[q.id] === option.id && !option.correct}>
          <input type="radio" name={q.id} value={option.id} bind:group={answers[q.id]} disabled={checked} />
          <span>{option.text}</span>
        </label>
      {/each}
      {#if checked}
        <p class="explain">{q.explain}</p>
      {/if}
    </fieldset>
  {/each}

  <div class="actions">
    {#if !checked}
      <Button onclick={() => (checked = true)} disabled={Object.keys(answers).length < questions.length}>
        Periksa jawaban
      </Button>
    {:else}
      <p class="result" role="status">Benar {score} dari {questions.length}.</p>
      <Button variant="quiet" onclick={reset}>Ulangi</Button>
    {/if}
  </div>
</section>

<style>
  h1 { font-size: var(--text-3xl); margin: 0 0 var(--space-3); }
  h2 { font-size: var(--text-2xl); margin: var(--space-7) 0 var(--space-4); }
  .lede { color: var(--text-2); max-width: var(--measure); margin: 0; }
  .grid { display: grid; gap: var(--space-4); }
  @media (min-width: 48rem) { .grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
  fieldset { border: var(--border-width) solid var(--border); border-radius: var(--radius-md); padding: var(--space-4); margin: 0 0 var(--space-4); }
  legend { padding: 0 var(--space-2); font-weight: 600; }
  .option { display: flex; gap: var(--space-3); align-items: flex-start; min-height: 44px; padding: var(--space-2) 0; cursor: pointer; }
  .option input { margin-top: 4px; }
  .option.correct { color: var(--ok); }
  .option.wrong { color: var(--err); }
  .explain { margin: var(--space-2) 0 0; color: var(--text-2); font-size: var(--text-sm); max-width: var(--measure); }
  .actions { display: flex; align-items: center; gap: var(--space-4); flex-wrap: wrap; }
  .result { margin: 0; font-weight: 600; }
</style>
