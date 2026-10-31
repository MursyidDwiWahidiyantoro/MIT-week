<script lang="ts">
  import { resolve } from '$app/paths';
  import { ArrowRight, ShieldCheck } from '@lucide/svelte';
</script>

<svelte:head>
  <title>AMBA — Periksa sebelum percaya</title>
  <meta
    name="description"
    content="Periksa tautan, tangkapan layar, dan pesan yang mencurigakan. Hasilnya indikatif, bukan bukti pasti."
  />
</svelte:head>

<section class="hero">
  <p class="mono-label">ALAT BANTU LITERASI DIGITAL</p>
  <h1>Periksa dulu sebelum menyebarkan.</h1>
  <p class="lede">
    Tempel tautan, unggah tangkapan layar, atau salin pesan. AMBA menunjukkan
    tanda-tanda yang perlu Anda waspadai — bukan menghakimi.
  </p>
  <div class="actions">
    <a class="btn-primary" href={resolve('/deteksi')}>
      Mulai periksa
      <span class="btn-icon" aria-hidden="true"><ArrowRight size={16} /></span>
    </a>
    <a class="btn-quiet" href={resolve('/edukasi')}>Pelajari cara mengenalinya</a>
  </div>
  <p class="note">
    <ShieldCheck size={16} aria-hidden="true" />
    Kami tidak menyimpan tautan, teks, atau gambar Anda.
  </p>
</section>

<style>
  .hero {
    max-width: 46rem;
    padding-block: var(--space-8);
  }
  h1 {
    font-size: clamp(2.25rem, 6vw, var(--text-display));
    letter-spacing: -0.02em;
    margin: var(--space-3) 0 var(--space-5);
    max-width: 18ch;
  }
  .lede {
    font-size: var(--text-lg);
    color: var(--text-2);
    max-width: var(--measure);
    margin: 0 0 var(--space-6);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
    align-items: center;
  }
  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    min-height: 44px;
    padding: var(--space-2) var(--space-3) var(--space-2) var(--space-5);
    background: var(--accent);
    color: #10161d;
    font-weight: 600;
    border-radius: var(--radius-md);
    text-decoration: none;
    transition: transform var(--dur-fast) var(--ease-out);
  }
  .btn-primary:hover { transform: translateY(-1px); }
  .btn-primary:active { transform: translateY(0); }
  .btn-icon {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border-radius: var(--radius-sm);
    background: rgb(0 0 0 / 0.12);
  }
  .btn-quiet {
    min-height: 44px;
    display: inline-flex;
    align-items: center;
    padding: 0 var(--space-4);
    color: var(--text-2);
    text-decoration: none;
    border-radius: var(--radius-md);
  }
  .btn-quiet:hover { color: var(--text); background: var(--hover-overlay); }
  .note {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-6);
    color: var(--text-3);
    font-size: var(--text-sm);
  }
</style>
