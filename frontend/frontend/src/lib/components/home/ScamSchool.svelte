<script lang="ts">
  import Icon from '$lib/components/ui/Icon.svelte';

  const LESSONS = [
    { title: 'Biaya paket palsu', minutes: 4, progress: 100 },
    { title: 'Trik nomor baru', minutes: 5, progress: 60 },
    { title: 'Mengenali tautan palsu', minutes: 3, progress: 25 },
    { title: 'Investasi yang terlalu bagus', minutes: 6, progress: 5 }
  ];

  const BADGES = [
    { label: 'Detektif tautan', tone: 'badge--indigo' },
    { label: 'Penelepon tenang', tone: 'badge--yellow' },
    { label: 'Ahli paket', tone: 'badge--lime' },
    { label: 'Penakluk hadiah palsu', tone: 'badge--coral' }
  ];

  let answer = $state<null | 'real' | 'scam'>(null);
  const correct = $derived(answer === 'scam');
</script>

<section class="section learn-section" id="learn" aria-labelledby="learn-title">
  <div class="wrap">
    <p class="eyebrow">Sekolah scam</p>
    <h2 class="h2" id="learn-title">Pelajari triknya, bukan cuma hasilnya.</h2>
    <p class="lede learn-lede">
      Pelajaran singkat yang disusun dari pesan penipuan sungguhan. Lima menit sudah cukup.
    </p>

    <div class="bento">
      <!-- Red flag of the week (7 cols) -->
      <article class="bento-card card card-flag lift">
        <div class="card-head">
          <span class="pill pill--coral">Tanda bahaya minggu ini</span>
          <span class="head-note">Trik “nomor baru”</span>
        </div>

        <blockquote class="bubble">
          Halo Bu, HP saya rusak. Ini <mark class="hl hl--yellow">nomor baru</mark> saya. Bisa
          kirim uang <mark class="hl hl--coral">sekarang juga</mark>?
          <mark class="hl hl--lime">Jangan telepon</mark>, saya sedang rapat.
        </blockquote>

        <ol class="explain">
          <li>
            <span class="dot dot--yellow" aria-hidden="true">1</span>
            <span><strong>Nomor asing.</strong> Keluarga asli tetap bisa dihubungi di nomor lama.</span>
          </li>
          <li>
            <span class="dot dot--coral" aria-hidden="true">2</span>
            <span><strong>Desakan.</strong> Penipu tidak suka Anda punya waktu berpikir.</span>
          </li>
          <li>
            <span class="dot dot--lime" aria-hidden="true">3</span>
            <span><strong>Larang menelepon.</strong> Mereka menghindari satu hal yang membuktikan identitas.</span>
          </li>
        </ol>
      </article>

      <!-- Quick lessons (5 cols) -->
      <article class="bento-card card--ink card-lessons lift">
        <div class="lessons-head">
          <span class="book" aria-hidden="true"><Icon name="book" size={22} /></span>
          <h3>Pelajaran singkat</h3>
        </div>
        <ul>
          {#each LESSONS as lesson (lesson.title)}
            <li>
              <div class="lesson-row">
                <span class="lesson-title">{lesson.title}</span>
                <span class="lesson-time">{lesson.minutes} mnt</span>
              </div>
              <div
                class="bar"
                role="img"
                aria-label={`Progres ${lesson.progress} persen`}
              >
                <span class="bar-fill" style:width="{lesson.progress}%"></span>
              </div>
            </li>
          {/each}
        </ul>
      </article>

      <!-- Quiz (7 cols) -->
      <article class="bento-card card--lime card-quiz lift">
        <p class="eyebrow quiz-eyebrow">Kuis kilat</p>
        <h3>Asli atau scam?</h3>
        <p class="quiz-bubble">
          Selamat! Anda menang HP gratis. Klaim dalam 10 menit: bit.ly/win-now
        </p>

        <div class="quiz-live" aria-live="polite">
          {#if answer === null}
            <div class="quiz-actions">
              <button class="btn btn--quiet btn--sm" type="button" onclick={() => (answer = 'real')}>
                Itu asli
              </button>
              <button class="btn btn--ink btn--sm" type="button" onclick={() => (answer = 'scam')}>
                Itu scam
              </button>
            </div>
          {:else}
            <div class="feedback">
              <p class="feedback-title">{correct ? 'Tepat sekali!' : 'Hampir! Ini scam.'}</p>
              <p class="feedback-body">
                Batas waktu palsu dan tautan yang dipendekkan adalah dua tanda klasik.
                Hadiah asli tidak pernah mengejar Anda.
              </p>
              <button class="btn btn--ink btn--sm" type="button" onclick={() => (answer = null)}>
                <Icon name="refresh" size={16} /> Coba lagi
              </button>
            </div>
          {/if}
        </div>
      </article>

      <!-- Badges (5 cols) -->
      <article class="bento-card card card-badges lift">
        <h3>Kumpulkan lencananya</h3>
        <p>Tuntaskan pelajaran, dapatkan stiker, lalu tantang keluarga.</p>
        <ul class="badge-list">
          {#each BADGES as badge (badge.label)}
            <li class="badge {badge.tone} lift">{badge.label}</li>
          {/each}
        </ul>
      </article>
    </div>
  </div>
</section>

<style>
  .learn-section {
    width: 100%;
    background: var(--white);
    border-bottom: var(--border-width) solid var(--border);
    padding-block: clamp(3.5rem, 7vw, 6rem);
  }
  .learn-lede { margin-top: var(--space-3); }

  /* 12-column Bento Grid */
  .bento {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-5);
    margin-top: var(--space-6);
  }
  @media (min-width: 60rem) {
    .bento {
      grid-template-columns: repeat(12, 1fr);
      gap: clamp(1.25rem, 2vw, 1.75rem);
    }
    .card-flag {
      grid-column: span 7;
    }
    .card-lessons {
      grid-column: span 5;
    }
    .card-quiz {
      grid-column: span 7;
    }
    .card-badges {
      grid-column: span 5;
    }
  }

  .bento-card {
    padding: clamp(1.25rem, 3vw, 2rem);
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  /* a) red flag */
  .card-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
  }
  .pill--coral { background: var(--coral); }
  .head-note { font-weight: 700; color: var(--muted); font-size: var(--text-sm); }
  .bubble {
    margin: 0;
    padding: 20px 22px;
    background: var(--paper);
    border: var(--border-width) solid var(--border);
    border-radius: 22px 22px 22px 6px;
    font-size: 1.1em;
    line-height: 1.8;
  }
  .hl {
    padding: 2px 8px;
    border: 2px solid var(--border);
    border-radius: 8px;
    font-weight: 700;
  }
  .hl--yellow { background: var(--yellow); }
  .hl--coral { background: var(--coral); }
  .hl--lime { background: var(--lime); }
  .explain { display: grid; gap: var(--space-3); margin: var(--space-5) 0 0; padding: 0; list-style: none; }
  .explain li { display: flex; align-items: flex-start; gap: var(--space-3); font-size: 0.98em; line-height: 1.45; }
  .dot {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    flex: none;
    border: var(--border-width) solid var(--border);
    border-radius: 50%;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: var(--text-sm);
    color: var(--ink);
  }
  .dot--yellow { background: var(--yellow); }
  .dot--coral { background: var(--coral); }
  .dot--lime { background: var(--lime); }

  /* b) lessons */
  .card-lessons {
    border-radius: var(--radius-lg);
    border: var(--border-width) solid var(--border);
    box-shadow: var(--shadow-hard);
  }
  .lessons-head { display: flex; align-items: center; gap: var(--space-3); margin-bottom: var(--space-4); }
  .book {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    background: var(--lime);
    color: var(--ink);
    border: var(--border-width) solid var(--border);
    border-radius: 14px;
    flex: none;
  }
  .card-lessons h3 { font-size: 1.4em; margin: 0; }
  .card-lessons ul { display: grid; gap: var(--space-4); margin: 0; padding: 0; list-style: none; }
  .lesson-row { display: flex; align-items: baseline; justify-content: space-between; gap: var(--space-3); }
  .lesson-title { font-weight: 700; }
  .lesson-time { color: var(--lime); font-size: var(--text-sm); white-space: nowrap; }
  .bar {
    height: 12px;
    margin-top: var(--space-2);
    background: var(--ink-soft);
    border-radius: var(--radius-pill);
    overflow: hidden;
  }
  .bar-fill { display: block; height: 100%; background: var(--lime); border-radius: inherit; }

  /* c) quiz */
  .card-quiz {
    border-radius: var(--radius-lg);
    border: var(--border-width) solid var(--border);
    box-shadow: var(--shadow-hard);
  }
  .quiz-eyebrow { color: var(--ink); margin-bottom: var(--space-1); }
  .card-quiz h3 { font-size: 1.5em; margin: 0 0 var(--space-3); }
  .quiz-bubble {
    margin: 0 0 var(--space-4);
    padding: var(--space-4);
    background: var(--white);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    font-size: 1.02em;
  }
  .quiz-actions { display: flex; flex-wrap: wrap; gap: var(--space-3); }
  .quiz-actions > * { flex: 1 1 8rem; }
  .feedback { display: grid; gap: var(--space-2); justify-items: start; }
  .feedback-title { margin: 0; font-family: var(--font-display); font-weight: 800; font-size: 1.2em; }
  .feedback-body { margin: 0; font-size: 0.98em; line-height: 1.45; }

  /* d) badges */
  .card-badges h3 { font-size: 1.4em; margin: 0 0 var(--space-2); }
  .card-badges p { margin: 0 0 var(--space-4); color: var(--body-2); font-size: 0.98em; }
  .badge-list { display: flex; flex-wrap: wrap; gap: var(--space-2); margin: 0; padding: 0; list-style: none; }
  .badge {
    padding: 10px 16px;
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-pill);
    box-shadow: 4px 4px 0 var(--ink);
    font-weight: 700;
    font-size: var(--text-sm);
    color: var(--ink);
  }
  .badge--indigo { background: var(--indigo); color: var(--white); }
  .badge--yellow { background: var(--yellow); }
  .badge--lime { background: var(--lime); }
  .badge--coral { background: var(--coral); }
</style>
