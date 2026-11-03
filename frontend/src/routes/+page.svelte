<script lang="ts">
  import { resolve } from '$app/paths';
  import Icon from '$lib/components/ui/Icon.svelte';
  import ScannerCard from '$lib/components/home/ScannerCard.svelte';
  import HowItWorks from '$lib/components/home/HowItWorks.svelte';
  import ScamSchool from '$lib/components/home/ScamSchool.svelte';
  import MadeForEveryone from '$lib/components/home/MadeForEveryone.svelte';
  import ModePicker from '$lib/components/mode/ModePicker.svelte';
  import { mode } from '$lib/mode.svelte';

  const TRUST = ['Tanpa jargon', 'Huruf bisa diperbesar', 'Aman untuk seisi rumah'];
</script>

<svelte:head>
  <title>AMBA — Periksa sebelum percaya</title>
  <meta
    name="description"
    content="Tempel tautan, pesan, atau nomor telepon. AMBA memeriksanya dalam hitungan detik dan menjelaskan alasannya dengan kata-kata sederhana."
  />
</svelte:head>

<section class="hero" id="top">
  <span class="blob blob--lime" aria-hidden="true"></span>
  <span class="blob blob--soft" aria-hidden="true"></span>

  <div class="hero-grid">
    <div class="hero-copy">
      <p class="pill pill--lime">
        <Icon name="sparkle" size={14} />
        Pemindai scam AI + sekolah scam gratis
      </p>

      <h1 class="h1">
        Kenali scam <span class="mark">sebelum</span> kena.
      </h1>

      <p class="lede">
        Tempel pesan, tautan, atau nomor telepon. Pemeriksaan berjalan dalam hitungan
        detik, dan alasannya dijelaskan dengan kata-kata sederhana. Semakin sering
        dipakai, semakin peka Anda.
      </p>

      <div class="cta-row">
        <a class="btn btn--primary" href="#scan">
          Periksa sekarang
          <Icon name="arrow-right" size={20} />
        </a>
        <a class="btn btn--quiet" href="#learn">Ambil satu pelajaran</a>
      </div>

      <ul class="trust">
        {#each TRUST as item (item)}
          <li>
            <span class="tick" aria-hidden="true"><Icon name="check" size={16} /></span>
            {item}
          </li>
        {/each}
      </ul>
    </div>

    <div class="hero-scan">
      <ScannerCard />
    </div>
  </div>
</section>

{#if mode.needsPicker}
  <ModePicker />
{/if}

<section class="marquee" aria-label="Jenis scam yang bisa dikenali">
  <div class="marquee__track">
    {#each [0, 1] as copy (copy)}
      <div class="marquee__group" aria-hidden={copy === 1 ? 'true' : undefined}>
        {#each ['Hadiah palsu', 'Phishing bank', 'Paket tertahan', 'Romansa online', 'Lowongan palsu', 'Investasi bodong', 'Penyamaran', 'Tagihan palsu'] as item (item)}
          <span class="marquee__item">{item}<Icon name="sparkle" size={20} /></span>
        {/each}
      </div>
    {/each}
  </div>
</section>

<HowItWorks />
<ScamSchool />
<MadeForEveryone />

<section class="family section" id="family">
  <div class="family-panel card--indigo">
    <span class="blob blob--lime" aria-hidden="true"></span>
    <span class="blob blob--yellow" aria-hidden="true"></span>
    <div class="family-body">
      <h2 class="h2">Ada pesan aneh masuk sekarang?</h2>
      <p class="family-lede">
        Periksa dulu, baru teruskan. Setelah itu kirim panduan keluarga ke orang yang
        Anda sayangi, supaya satu grup tetap selangkah lebih aman.
      </p>
      <div class="cta-row">
        <a class="btn btn--lime" href={resolve('/deteksi')}>Periksa gratis</a>
        <a class="btn btn--quiet" href={resolve('/cek-cepat')}>Buka panduan keluarga</a>
      </div>
    </div>
  </div>
</section>

<style>
  .hero {
    position: relative;
    overflow: hidden;
    padding-block: clamp(1.5rem, 4vw, 2.5rem) clamp(3rem, 7vw, 5rem);
  }
  .blob { z-index: 0; }
  .blob--lime {
    width: 440px;
    height: 440px;
    background: var(--lime);
    right: -140px;
    top: -90px;
  }
  .blob--soft {
    width: 300px;
    height: 300px;
    background: var(--indigo-soft);
    left: -110px;
    bottom: -80px;
  }
  .hero-grid {
    position: relative;
    z-index: 1;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: clamp(2rem, 5vw, 3.5rem);
  }
  .hero-copy { flex: 1 1 460px; min-width: 0; }
  .hero-scan { flex: 1 1 420px; min-width: 0; }

  .pill--lime { background: var(--lime); }
  h1 { margin: var(--space-4) 0 var(--space-5); max-width: 15ch; }
  .mark {
    display: inline-block;
    padding: 0 10px;
    background: var(--lime);
    border: var(--border-width) solid var(--ink);
    border-radius: 12px;
    transform: rotate(-1.5deg);
  }
  .cta-row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-4);
    margin-top: var(--space-6);
  }
  .trust {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3) var(--space-5);
    margin: var(--space-6) 0 0;
    padding: 0;
    list-style: none;
    font-weight: 700;
    font-size: 0.95em;
  }
  .trust li { display: inline-flex; align-items: center; gap: var(--space-2); }
  .tick {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    background: var(--ink);
    color: var(--lime);
    border-radius: 50%;
    flex: none;
  }

  .family-panel {
    position: relative;
    overflow: hidden;
    padding: clamp(2rem, 5vw, 3.25rem);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-xl);
  }
  .family-panel .blob--lime { width: 240px; height: 240px; right: -70px; top: -70px; }
  .family-panel .blob--yellow { width: 120px; height: 120px; right: 120px; bottom: -60px; }
  .family-body { position: relative; z-index: 1; max-width: 44rem; }
  .family-lede {
    margin: var(--space-4) 0 0;
    color: var(--on-indigo-text);
    font-size: 1.2em;
    max-width: 34em;
  }
  .family-panel :global(.btn--quiet) { color: var(--ink); }

  @media (max-width: 40rem) {
    .blob--lime { width: 260px; height: 260px; right: -120px; top: -80px; }
    .blob--soft { width: 180px; height: 180px; }
  }
</style>
