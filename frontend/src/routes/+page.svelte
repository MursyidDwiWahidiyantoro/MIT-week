<script lang="ts">
  import { resolve } from '$app/paths';
  import { ArrowRight, ShieldCheck, Link2, Image, MessageSquare } from '@lucide/svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import Panel from '$lib/components/ui/Panel.svelte';

  let link = $state('');

  function go(): void {
    const value = link.trim();
    window.location.href = value
      ? `${resolve('/deteksi')}?q=${encodeURIComponent(value)}`
      : resolve('/deteksi');
  }

  const kinds = [
    { icon: Link2, title: 'Tautan', detail: 'Tempel alamat situs yang mengirim pesan mencurigakan.' },
    { icon: Image, title: 'Tangkapan layar', detail: 'Unggah gambar percakapan atau unggahan.' },
    { icon: MessageSquare, title: 'Teks pesan', detail: 'Salin isi pesan berantai yang Anda terima.' }
  ];

  const steps = [
    { title: 'Masukkan konten', detail: 'Tempel tautan, unggah gambar, atau salin teks.' },
    { title: 'AMBA memeriksa', detail: 'Beberapa pemeriksaan berjalan bersamaan dan hasilnya muncul bertahap.' },
    { title: 'Baca hasilnya', detail: 'Lihat tanda bahaya, batas pemeriksaan, dan langkah berikutnya.' }
  ];
</script>

<svelte:head>
  <title>AMBA — Periksa sebelum percaya</title>
  <meta name="description" content="Periksa tautan, tangkapan layar, dan pesan yang mencurigakan. Hasilnya indikatif, bukan bukti pasti." />
</svelte:head>

<section class="hero">
  <p class="mono-label">ALAT BANTU LITERASI DIGITAL</p>
  <h1>Periksa dulu sebelum menyebarkan.</h1>
  <p class="lede">
    Tempel tautan, unggah tangkapan layar, atau salin pesan. AMBA menunjukkan
    tanda-tanda yang perlu Anda waspadai — bukan menghakimi.
  </p>

  <form class="instant" onsubmit={(e) => { e.preventDefault(); go(); }}>
    <label class="sr-only" for="hero-link">Tautan yang ingin diperiksa</label>
    <input id="hero-link" type="url" bind:value={link} placeholder="https://contoh.com/artikel" />
    <Button type="submit" size="lg">
      Mulai periksa
      <span class="icon" aria-hidden="true"><ArrowRight size={16} /></span>
    </Button>
  </form>

  <p class="note">
    <ShieldCheck size={16} aria-hidden="true" />
    Kami tidak menyimpan tautan, teks, atau gambar Anda.
  </p>
</section>

<section class="kinds" aria-labelledby="kinds-title">
  <h2 id="kinds-title">Tiga hal yang bisa diperiksa</h2>
  <div class="grid">
    {#each kinds as kind (kind.title)}
      <Panel>
        <span class="kind-icon" aria-hidden="true"><kind.icon size={20} /></span>
        <h3>{kind.title}</h3>
        <p>{kind.detail}</p>
      </Panel>
    {/each}
  </div>
</section>

<section class="steps" aria-labelledby="steps-title">
  <h2 id="steps-title">Cara kerjanya</h2>
  <ol>
    {#each steps as step, i (step.title)}
      <li>
        <span class="num" aria-hidden="true">{i + 1}</span>
        <h3>{step.title}</h3>
        <p>{step.detail}</p>
      </li>
    {/each}
  </ol>
</section>

<style>
  .hero { max-width: 46rem; padding-block: var(--space-8); }
  h1 {
    font-size: clamp(2.25rem, 6vw, var(--text-display));
    letter-spacing: -0.02em;
    margin: var(--space-3) 0 var(--space-5);
    max-width: 18ch;
  }
  .lede { font-size: var(--text-lg); color: var(--text-2); max-width: var(--measure); margin: 0 0 var(--space-6); }

  .instant { display: flex; flex-wrap: wrap; gap: var(--space-3); align-items: stretch; }
  .instant input {
    flex: 1 1 20rem; min-height: 52px; padding: 0 var(--space-4);
    background: var(--surface-2); color: var(--text);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-md); font: inherit;
  }
  .instant input:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .icon { display: grid; place-items: center; width: 24px; height: 24px; }

  .note { display: flex; align-items: center; gap: var(--space-2); margin-top: var(--space-6); color: var(--text-3); font-size: var(--text-sm); }

  .kinds, .steps { padding-block: var(--space-7); border-top: var(--border-width) solid var(--border); }
  h2 { font-size: var(--text-2xl); margin: 0 0 var(--space-5); }
  .grid { display: grid; gap: var(--space-4); }
  @media (min-width: 48rem) { .grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
  .kind-icon { color: var(--accent); display: inline-flex; margin-bottom: var(--space-3); }
  h3 { font-size: var(--text-lg); margin: 0 0 var(--space-2); }
  .grid p, .steps p { margin: 0; color: var(--text-2); font-size: var(--text-sm); }

  ol { list-style: none; margin: 0; padding: 0; display: grid; gap: var(--space-5); }
  @media (min-width: 48rem) { ol { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
  .num {
    display: grid; place-items: center; width: 32px; height: 32px; margin-bottom: var(--space-3);
    border: var(--border-width) solid var(--border-strong); border-radius: 50%;
    font-family: var(--font-mono); font-size: var(--text-sm);
  }
  .sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); }
</style>
