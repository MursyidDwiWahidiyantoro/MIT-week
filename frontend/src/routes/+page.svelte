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
  <!-- Atmospheric decorative background spanning full viewport width -->
  <div class="hero-backdrop" aria-hidden="true">
    <span class="blob blob--lime"></span>
    <span class="blob blob--soft"></span>
    <span class="blob blob--yellow"></span>
  </div>

  <div class="wrap hero-wrap">
    <div class="hero-grid">
      <div class="hero-copy">
        <div class="pill pill--lime eyebrow-pill">
          <Icon name="sparkle" size={14} />
          <span>Pemindai scam AI + sekolah scam gratis</span>
        </div>

        <h1 class="hero-title">
          Kenali scam <span class="mark">sebelum</span> kena.
        </h1>

        <p class="hero-lede">
          Tempel pesan, tautan, atau nomor telepon. Pemeriksaan berjalan dalam hitungan
          detik, dan alasannya dijelaskan dengan kata-kata sederhana. Semakin sering
          dipakai, semakin peka Anda.
        </p>

        <div class="cta-row">
          <a class="btn btn--primary btn--hero group" href={resolve('/deteksi')}>
            <span>Periksa sekarang</span>
            <span class="btn-icon-circle" aria-hidden="true">
              <Icon name="arrow-right" size={18} />
            </span>
          </a>
          <a class="btn btn--quiet" href="#learn">Ambil satu pelajaran</a>
        </div>

        <ul class="trust">
          {#each TRUST as item (item)}
            <li>
              <span class="tick" aria-hidden="true"><Icon name="check" size={16} /></span>
              <span>{item}</span>
            </li>
          {/each}
        </ul>
      </div>

      <div class="hero-scan">
        <ScannerCard />
      </div>
    </div>
  </div>
</section>

{#if mode.needsPicker}
  <div class="wrap">
    <ModePicker />
  </div>
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

<section class="family-section section" id="family">
  <div class="wrap">
    <div class="family-panel card--indigo lift">
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
  </div>
</section>

<style>
  /* Full screen, expansive desktop hero */
  .hero {
    position: relative;
    width: 100%;
    min-height: calc(100dvh - 5.5rem);
    display: flex;
    align-items: center;
    padding-block: clamp(2.5rem, 5vw, 4.5rem);
    overflow: hidden;
    background: var(--paper);
  }

  .hero-backdrop {
    position: absolute;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
    z-index: 0;
  }
  .blob {
    position: absolute;
    border-radius: 50%;
    filter: blur(0px);
  }
  .blob--lime {
    width: clamp(340px, 35vw, 540px);
    height: clamp(340px, 35vw, 540px);
    background: var(--lime);
    right: -8vw;
    top: -12vw;
  }
  .blob--soft {
    width: clamp(260px, 26vw, 400px);
    height: clamp(260px, 26vw, 400px);
    background: var(--indigo-soft);
    left: -6vw;
    bottom: -8vw;
  }
  .blob--yellow {
    width: clamp(160px, 16vw, 240px);
    height: clamp(160px, 16vw, 240px);
    background: var(--yellow);
    right: 30%;
    bottom: -60px;
    opacity: 0.7;
  }

  /* Inherits .wrap: one container for the whole page, so the hero spreads as
   * wide as every section below it instead of sitting in its own box. */
  .hero-wrap {
    position: relative;
    z-index: 1;
  }

  .hero-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: clamp(2.5rem, 4vw, 4rem);
    align-items: center;
  }
  @media (min-width: 64rem) {
    .hero-grid {
      grid-template-columns: minmax(0, 1.2fr) minmax(0, 0.8fr);
      gap: clamp(3rem, 4.5vw, 6.5rem);
    }
  }
  @media (min-width: 100rem) {
    .hero-grid {
      grid-template-columns: minmax(0, 1.3fr) minmax(0, 0.7fr);
    }
  }

  .hero-copy {
    min-width: 0;
  }
  .hero-scan {
    min-width: 0;
    display: flex;
    justify-content: center;
  }

  .eyebrow-pill {
    background: var(--lime);
    margin-bottom: var(--space-3);
  }
  .hero-title {
    margin: var(--space-3) 0 var(--space-4);
    font-size: var(--text-display);
    line-height: var(--leading-display);
    max-width: 16ch;
  }
  .mark {
    display: inline-block;
    padding: 0 12px;
    background: var(--lime);
    border: var(--border-width) solid var(--ink);
    border-radius: 14px;
    transform: rotate(-1.5deg);
  }
  .hero-lede {
    font-size: clamp(var(--text-base), 1.15vw, 1.35rem);
    line-height: var(--leading-normal);
    color: var(--body-2);
    max-width: 38ch;
    margin: 0;
  }

  .cta-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-3) var(--space-4);
    margin-top: var(--space-6);
  }

  /* Button-in-button trailing icon circle */
  .btn--hero {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    padding-right: 12px;
  }
  .btn-icon-circle {
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    background: var(--lime);
    color: var(--ink);
    border-radius: 50%;
    border: 2px solid var(--ink);
    transition: transform var(--dur-fast) var(--ease-out);
  }
  .btn--hero:hover .btn-icon-circle {
    transform: translateX(3px) scale(1.08);
  }

  .trust {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
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
    width: 26px;
    height: 26px;
    background: var(--ink);
    color: var(--lime);
    border-radius: 50%;
    flex: none;
  }

  /* Family callout section */
  .family-section {
    width: 100%;
    background: var(--paper);
    padding-block: clamp(3.5rem, 7vw, 6rem);
  }
  .family-panel {
    position: relative;
    overflow: hidden;
    padding: clamp(2rem, 5vw, 3.5rem);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-hard-xl);
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

  @media (max-width: 48rem) {
    .hero {
      min-height: auto;
      padding-block: var(--space-5) var(--space-8);
    }
    .hero-title {
      font-size: clamp(2.3rem, 9vw, 3.4rem);
    }
    .cta-row {
      flex-direction: column;
      align-items: stretch;
      gap: var(--space-3);
    }
    .cta-row :global(.btn) {
      width: 100%;
      justify-content: center;
    }
    .trust {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--space-2);
    }
  }
</style>
