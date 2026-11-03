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
      title: 'Mode huruf besar',
      body: 'Tiga ukuran, satu ketukan, dari atas setiap halaman.',
      tone: 'tile--soft'
    },
    {
      icon: 'speaker',
      title: 'Bacakan hasil',
      body: 'Setiap kesimpulan dan pelajaran bisa didengar dengan kata yang tenang.',
      tone: 'tile--yellow'
    },
    {
      icon: 'moon',
      title: 'Mode tenang',
      body: 'Gerak minimal untuk mata yang lelah dan kepala yang penuh.',
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

<section class="section access" id="access" aria-labelledby="access-title">
  <div class="access-grid">
    <div class="access-copy">
      <p class="eyebrow">Untuk semua orang</p>
      <h2 class="h2" id="access-title">
        Cukup keren untuk feed Anda. Cukup jelas untuk orang tua Anda.
      </h2>
      <p class="lede">
        Tombol besar, kontras kuat, dan kata sehari-hari sejak awal. Tambahan bisa
        dinyalakan kapan saja.
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
</section>

<style>
  .access-grid {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: var(--space-6);
  }
  .access-copy { flex: 1 1 20rem; min-width: 0; }
  .access-copy .lede { margin-top: var(--space-4); }
  .tiles {
    flex: 1.4 1 28rem;
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-4);
    margin: 0;
    padding: 0;
    list-style: none;
    min-width: 0;
  }
  .tile {
    flex: 1 1 12rem;
    min-width: 0;
    padding: clamp(1.1rem, 2.5vw, 1.5rem);
    border-radius: 24px;
  }
  .tile--soft { background: var(--indigo-soft); }
  .tile--yellow { background: var(--yellow); }
  .tile--ink { background: var(--ink); color: var(--white); box-shadow: var(--shadow-indigo); }
  .tile--lime { background: var(--lime); }
  .icon { display: inline-flex; margin-bottom: var(--space-3); }
  h3 { font-size: 1.3em; margin: 0 0 var(--space-2); }
  p { margin: 0; font-size: 0.95em; line-height: 1.45; }
  .tile--ink p { color: var(--on-indigo-text); }
</style>
