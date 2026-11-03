<script lang="ts">
  const STEPS = [
    {
      n: '1',
      tone: 'indigo',
      title: 'Tempel isinya',
      body: 'Dapat pesan aneh, tautan, atau panggilan? Masukkan di sini. Tidak perlu akun untuk mencoba.'
    },
    {
      n: '2',
      tone: 'lime',
      title: 'AI memeriksa',
      body: 'Kata-kata, tautan, dan taktik desakan diperiksa, lalu keluar satu kesimpulan yang jelas.'
    },
    {
      n: '3',
      tone: 'yellow',
      title: 'Pahami alasannya',
      body: 'Setiap tanda bahaya dijelaskan, jadi lain kali Anda bisa mengenalinya sendiri.'
    }
  ] as const;
</script>

<section class="section" aria-labelledby="how-title">
  <p class="eyebrow">Cara kerja</p>
  <h2 class="h2" id="how-title">Tiga langkah, tanpa panik.</h2>

  <ol class="steps">
    {#each STEPS as step (step.n)}
      <li class="step card lift">
        <span class="num num--{step.tone}" aria-hidden="true">{step.n}</span>
        <h3>{step.title}</h3>
        <p>{step.body}</p>
      </li>
    {/each}
  </ol>
</section>

<style>
  .steps {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-5);
    margin: var(--space-6) 0 0;
    padding: 0;
    list-style: none;
  }
  .step {
    flex: 1 1 17rem;
    min-width: 0;
    padding: clamp(1.25rem, 3vw, 1.75rem);
  }
  .num {
    display: grid;
    place-items: center;
    width: 64px;
    height: 64px;
    margin-bottom: var(--space-4);
    border: var(--border-width) solid var(--ink);
    border-radius: 20px;
    box-shadow: var(--shadow-hard-xs);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.9em;
    line-height: 1;
  }
  .num--indigo { background: var(--indigo); color: var(--white); }
  .num--lime { background: var(--lime); color: var(--ink); }
  .num--yellow { background: var(--yellow); color: var(--ink); }
  h3 { font-size: 1.5em; margin: 0 0 var(--space-2); }
  p { margin: 0; color: var(--body-2); }
</style>
