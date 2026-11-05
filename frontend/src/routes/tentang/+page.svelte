<script lang="ts">
  import Icon from '$lib/components/ui/Icon.svelte';

  const PRINCIPLES = [
    'Tidak mengklaim kepastian. Kami memakai kata “terindikasi”, bukan “pasti”.',
    'Tidak mengarang sumber. Sumber yang tidak terverifikasi tidak ditampilkan sebagai bukti.',
    'Tidak menyimpan data Anda. Tidak ada akun dan tidak ada basis data.',
    'Tidak menghakimi orang. Hasil membahas kontennya, bukan pengirimnya.'
  ];

  const LIMITS = [
    'Skor 0–100 menunjukkan tingkat risiko, bukan tingkat keyakinan.',
    'Analisis gambar (ELA) melemah pada PNG dan tangkapan layar.',
    'Pemeriksaan fakta hanya memakai sumber yang sudah kami tinjau.',
    'Hasil tidak dapat dipakai sebagai alat bukti hukum.',
    'Menu bagikan ke AMBA hanya tersedia di peramban yang mendukung Web Share Target, misalnya Chrome di Android. Di iOS, gunakan tombol tempel.'
  ];
</script>

<svelte:head>
  <title>Tentang — AMBA</title>
  <meta
    name="description"
    content="AMBA adalah alat bantu literasi digital. Hasil analisis bersifat indikatif, bukan bukti pasti."
  />
</svelte:head>

<section class="head">
  <p class="eyebrow">Tentang</p>
  <h1 class="h1">AMBA</h1>
  <p class="lede">
    AMBA (Aplikasi Mengenali &amp; Bertindak Aman) membantu Anda memeriksa tautan,
    tangkapan layar, dan pesan yang mencurigakan sebelum menyebarkannya.
  </p>
</section>

<section class="team" aria-labelledby="team-title">
  <h2 class="h2" id="team-title">Tim di balik AMBA</h2>
  <p class="lede team-lede">
    Dibuat untuk MIT-Week 2026, Universitas Mulawarman.
  </p>
  <ul class="team-grid">
    <li><strong>Mursyid D. W.</strong><span>PM &amp; Lead Architect</span></li>
    <li><strong>M. Ridho A.</strong><span>UI/UX Lead</span></li>
    <li><strong>Bagus Joshua P.</strong><span>Golang Backend Lead</span></li>
    <li><strong>Jericho M. T.</strong><span>Frontend Lead</span></li>
  </ul>
</section>

<section class="data card lift" aria-labelledby="data-title">
  <h2 class="h2" id="data-title">Apa yang kami kirim ke layanan AI</h2>
  <dl>
    <div class="row">
      <dt><span class="tag tag--send">Dikirim</span></dt>
      <dd>
        Isi yang Anda periksa, yaitu tautan, teks, atau gambar, dikirim ke model
        bahasa untuk dinilai. Pemeriksaan fakta memakai pencarian ke sumber terpercaya.
      </dd>
    </div>
    <div class="row">
      <dt><span class="tag tag--no">Tidak dikirim</span></dt>
      <dd>
        Nama, akun, alamat IP, atau riwayat pemeriksaan Anda. AMBA tidak punya akun
        dan tidak menyimpan input.
      </dd>
    </div>
    <div class="row">
      <dt><span class="tag tag--none">Disimpan</span></dt>
      <dd>
        Tidak ada. Hasil hanya tersimpan selama halaman hasil terbuka. Satu-satunya
        pengecualian: saat Anda membagikan konten lewat menu bagikan ponsel, item
        ditahan beberapa detik di perangkat Anda lalu dihapus.
      </dd>
    </div>
  </dl>
</section>

<div class="two-col">
  <section class="card limits lift" aria-labelledby="limits-title">
    <h2 class="h2" id="limits-title">Batas yang perlu Anda tahu</h2>
    <ul>
      {#each LIMITS as item (item)}
        <li>
          <span class="bullet" aria-hidden="true"><Icon name="alert" size={16} /></span>
          {item}
        </li>
      {/each}
    </ul>
  </section>

  <section class="card principles lift" aria-labelledby="principles-title">
    <h2 class="h2" id="principles-title">Prinsip kami</h2>
    <ul>
      {#each PRINCIPLES as item (item)}
        <li>
          <span class="bullet bullet--ok" aria-hidden="true"><Icon name="check" size={16} /></span>
          {item}
        </li>
      {/each}
    </ul>
  </section>
</div>

<p class="disclaimer">
  Hasil analisis bersifat indikatif dan bukan alat bukti hukum. Untuk keputusan
  penting, konfirmasikan ke pihak berwenang atau sumber resmi.
</p>

<style>
  .head { max-width: 54rem; margin-bottom: var(--space-7); }
  h1 { margin-bottom: var(--space-3); }

  .team { margin-bottom: var(--space-7); }
  .team-lede { margin-top: var(--space-3); }
  .team-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-3);
    margin: var(--space-5) 0 0;
    padding: 0;
    list-style: none;
  }
  @media (min-width: 48rem) { .team-grid { grid-template-columns: repeat(4, 1fr); } }
  .team-grid li {
    display: grid;
    gap: var(--space-1);
    padding: var(--space-4);
    background: var(--white);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-hard-xs);
  }
  .team-grid strong { font-family: var(--font-display); font-size: 1.05em; }
  .team-grid span { color: var(--body-2); font-size: var(--text-sm); }

  .data { padding: clamp(1.25rem, 3vw, 2rem); }
  dl { display: grid; gap: var(--space-5); margin: var(--space-5) 0 0; }
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: var(--space-3) var(--space-5);
    padding-top: var(--space-4);
    border-top: var(--border-width) dashed var(--border);
  }
  .row:first-child { padding-top: 0; border-top: 0; }
  dt { flex: 0 0 auto; }
  dd { flex: 1 1 22rem; margin: 0; min-width: 0; max-width: var(--measure); }
  .tag {
    display: inline-flex;
    padding: 4px 14px;
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-pill);
    font-weight: 700;
    font-size: var(--text-sm);
    white-space: nowrap;
  }
  .tag--send { background: var(--yellow); }
  .tag--no { background: var(--coral); }
  .tag--none { background: var(--lime); }

  .two-col { display: grid; grid-template-columns: 1fr; gap: var(--space-5); margin-top: var(--space-5); }
  @media (min-width: 48rem) { .two-col { grid-template-columns: repeat(2, 1fr); } }
  .two-col > * { min-width: 0; padding: clamp(1.25rem, 3vw, 1.9rem); display: flex; flex-direction: column; }
  .limits { background: var(--yellow); }
  .principles { background: var(--indigo-soft); }
  ul { margin: var(--space-4) 0 0; padding: 0; list-style: none; display: grid; gap: var(--space-3); }
  li { display: flex; align-items: flex-start; gap: var(--space-3); line-height: 1.45; }
  .bullet {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    flex: none;
    background: var(--white);
    border: 2px solid var(--border);
    border-radius: 50%;
  }
  .bullet--ok { background: var(--lime); }

  .disclaimer {
    margin-top: var(--space-6);
    padding: var(--space-4) var(--space-5);
    background: var(--white);
    border: var(--border-width) dashed var(--border);
    border-radius: var(--radius-md);
    color: var(--body-2);
    font-size: var(--text-sm);
    max-width: 52rem;
  }
</style>
