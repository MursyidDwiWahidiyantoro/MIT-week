<script lang="ts">
  import Icon, { type IconName } from '$lib/components/ui/Icon.svelte';

  const TILES: Array<{
    icon: IconName;
    title: string;
    body: string;
    tone: string;
  }> = [
    {
      icon: 'type',
      title: 'Teks mudah dibaca',
      body: 'Huruf jelas dan kontras kuat, agar informasi lebih mudah dibaca.',
      tone: 'tile--soft'
    },
    {
      icon: 'shield-check',
      title: 'Alasan yang jelas',
      body: 'Kenali tanda yang mencurigakan melalui penjelasan singkat, satu per satu.',
      tone: 'tile--yellow'
    },
    {
      icon: 'check',
      title: 'Langkah sederhana',
      body: 'Tempel pesan atau tautan, periksa, lalu pahami hasilnya.',
      tone: 'tile--ink'
    },
    {
      icon: 'chat',
      title: 'Kata sederhana',
      body: 'Tanpa istilah teknis. Kalau satu kalimat butuh kamus, kami tulis ulang.',
      tone: 'tile--lime'
    }
  ];
</script>

<section class="section access-section" id="access" aria-labelledby="access-title">
  <div class="wrap">
    <div class="access-grid">
      <div class="access-copy">
        <p class="eyebrow">Untuk semua orang</p>
        <h2 class="h2" id="access-title">
          Cukup keren untuk feed Anda. Cukup jelas untuk orang tua Anda.
        </h2>
        <p class="lede">
          Tombol besar, kontras kuat, dan kata sehari-hari. Dibuat jelas sejak awal,
          tanpa pengaturan tambahan.
        </p>
      </div>

      <ul class="tiles">
        {#each TILES as tile (tile.title)}
          <li class="tile card lift {tile.tone}">
            <span class="icon" aria-hidden="true"><Icon name={tile.icon} size={30} /></span>
            <h3>{tile.title}</h3>
            <p>{tile.body}</p>
          </li>
        {/each}
      </ul>
    </div>
  </div>
</section>

<style>
  .access-section {
    width: 100%;
    background: var(--paper);
    border-bottom: var(--border-width) solid var(--border);
    padding-block: clamp(3.5rem, 7vw, 6rem);
  }
  .access-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: clamp(2rem, 4vw, 3.5rem);
    align-items: center;
  }
  @media (min-width: 60rem) {
    .access-grid {
      grid-template-columns: 1fr 1.3fr;
      gap: clamp(2.5rem, 5vw, 4.5rem);
    }
  }

  .access-copy { min-width: 0; }
  .access-copy .lede { margin-top: var(--space-4); max-width: 36ch; }

  .tiles {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-4);
    margin: 0;
    padding: 0;
    list-style: none;
    min-width: 0;
  }
  @media (min-width: 34rem) {
    .tiles {
      grid-template-columns: repeat(2, 1fr);
    }
  }

  .tile {
    min-width: 0;
    padding: clamp(1.2rem, 2.5vw, 1.6rem);
    border-radius: 24px;
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .tile--soft { background: var(--indigo-soft); }
  .tile--yellow { background: var(--yellow); }
  .tile--ink { background: var(--ink); color: var(--white); box-shadow: var(--shadow-indigo); }
  .tile--lime { background: var(--lime); }
  .icon { display: inline-flex; margin-bottom: var(--space-3); }
  h3 { font-size: 1.3em; margin: 0 0 var(--space-2); }
  p { margin: 0; font-size: 0.98em; line-height: 1.5; }
  .tile--ink p { color: var(--on-indigo-text); }
</style>
