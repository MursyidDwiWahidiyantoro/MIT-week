<script lang="ts">
  import { resolve } from '$app/paths';
  import Button from '$lib/components/ui/Button.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  /**
   * The family guide the home page links to. Everything here is client-only:
   * the safe word is generated in the browser and never sent anywhere, and the
   * card prints from the same page.
   */

  // Two concrete, easy-to-say Indonesian nouns. A safe word only works if the
  // whole family can remember and pronounce it.
  const WORDS = [
    'Bakso', 'Sepeda', 'Kucing', 'Payung', 'Kopi', 'Bambu', 'Cempedak', 'Layang',
    'Mangga', 'Kereta', 'Bebek', 'Jendela', 'Sapu', 'Kelapa', 'Sepatu', 'Gunting'
  ];

  let safeWord = $state('');
  let copied = $state(false);

  function generate(): void {
    // crypto.getRandomValues avoids Math.random's predictability and works in
    // every browser this app supports.
    const buf = new Uint32Array(2);
    crypto.getRandomValues(buf);
    const a = WORDS[buf[0] % WORDS.length];
    const b = WORDS[buf[1] % WORDS.length];
    safeWord = a === b ? `${a} Dua` : `${a} ${b}`;
    copied = false;
  }

  async function copy(): Promise<void> {
    try {
      await navigator.clipboard.writeText(safeWord);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      // Selecting the text by hand still works.
    }
  }

  function print(): void {
    window.print();
  }
</script>

<svelte:head>
  <title>Panduan keluarga — AMBA</title>
  <meta
    name="description"
    content="Panduan singkat untuk keluarga: kata sandi aman, tiga aturan sederhana, dan kartu yang bisa dicetak."
  />
</svelte:head>

<section class="head">
  <p class="eyebrow">Untuk keluarga</p>
  <h1 class="h1">Satu kata rahasia, satu keluarga aman.</h1>
  <p class="lede">
    Penipu bisa memalsukan suara, foto, dan nomor. Kata sandi keluarga adalah
    hal yang tidak bisa mereka tebak. Buat sekarang, lalu ajarkan ke orang di
    rumah.
  </p>
</section>

<section class="card lift" aria-labelledby="sandi-title">
  <h2 class="h2" id="sandi-title">1. Buat kata sandi keluarga</h2>
  <p class="body">
    Dua kata yang mudah diingat, tidak berkaitan dengan nama, alamat, atau
    tanggal lahir. Simpan di kepala, bukan di chat.
  </p>

  <div class="generator">
    <output class="word" aria-live="polite">
      {safeWord || 'Belum dibuat'}
    </output>
    <div class="gen-actions">
      <Button onclick={generate}>
        <Icon name="refresh" size={18} />
        {safeWord ? 'Buat ulang' : 'Buat kata sandi'}
      </Button>
      {#if safeWord}
        <Button variant="quiet" onclick={copy}>
          <Icon name={copied ? 'check' : 'copy'} size={18} />
          {copied ? 'Tersalin' : 'Salin'}
        </Button>
      {/if}
    </div>
  </div>
  <p class="note">
    Dibuat di perangkat ini saja. AMBA tidak menyimpan atau mengirim kata sandi.
  </p>
</section>

<section class="rules" aria-labelledby="aturan-title">
  <h2 class="h2" id="aturan-title">2. Tiga aturan untuk semua orang</h2>
  <ol class="rule-list">
    <li class="card lift">
      <span class="num" aria-hidden="true">1</span>
      <h3>Minta kata sandi</h3>
      <p>
        Kalau ada yang mengaku keluarga lewat nomor baru, minta kata sandi
        keluarga. Salah jawab berarti bukan mereka.
      </p>
    </li>
    <li class="card lift">
      <span class="num" aria-hidden="true">2</span>
      <h3>Telepon nomor lama</h3>
      <p>
        Jangan pakai nomor di pesan itu. Hubungi nomor yang sudah lama tersimpan
        di kontak Anda.
      </p>
    </li>
    <li class="card lift">
      <span class="num" aria-hidden="true">3</span>
      <h3>Tidak ada rahasia</h3>
      <p>
        Permintaan uang atau OTP yang harus disembunyikan dari keluarga adalah
        tanda bahaya, bukan alasan untuk diam.
      </p>
    </li>
  </ol>
</section>

<section class="card print-card" aria-labelledby="kartu-title">
  <h2 class="h2" id="kartu-title">3. Cetak kartunya</h2>
  <p class="body">
    Tempel di pintu kulkas atau simpan di dompet. Kartu ini berisi kata sandi
    dan tiga aturan di atas.
  </p>

  <div class="printable">
    <p class="p-brand">AMBA — Panduan keluarga</p>
    <p class="p-label">Kata sandi keluarga kami</p>
    <p class="p-word">{safeWord || '____________________'}</p>
    <ol class="p-rules">
      <li>Minta kata sandi sebelum percaya.</li>
      <li>Telepon nomor lama, bukan nomor baru.</li>
      <li>Tidak ada permintaan uang yang perlu dirahasiakan.</li>
    </ol>
    <p class="p-foot">
      Hasil pemeriksaan AMBA bersifat indikatif, bukan bukti pasti.
    </p>
  </div>

  <Button variant="quiet" onclick={print}>
    <Icon name="print" size={18} /> Cetak kartu
  </Button>
</section>

<section class="next">
  <h2 class="h2">Langkah berikutnya</h2>
  <div class="next-actions">
    <Button href={resolve('/cek-cepat')} variant="primary">Cek situasi sekarang</Button>
    <Button href={resolve('/edukasi')} variant="quiet">Baca modus yang sering terjadi</Button>
  </div>
</section>

<style>
  .head { max-width: 54rem; margin-bottom: var(--space-7); }
  h1 { margin-bottom: var(--space-3); }
  .body { margin: var(--space-3) 0 0; color: var(--body-2); max-width: var(--measure); }
  .note { margin: var(--space-3) 0 0; color: var(--muted); font-size: var(--text-sm); }

  section.card { padding: clamp(1.25rem, 3vw, 2rem); margin-bottom: var(--space-5); }

  .generator {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-4);
    margin-top: var(--space-5);
  }
  .word {
    min-height: 64px;
    display: inline-flex;
    align-items: center;
    padding: 0 var(--space-5);
    background: var(--safe);
    color: var(--on-fill);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-md);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: clamp(1.4rem, 3vw, 2rem);
    letter-spacing: -0.02em;
  }
  .gen-actions { display: flex; flex-wrap: wrap; gap: var(--space-3); }

  .rules { margin-bottom: var(--space-5); }
  .rule-list {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-4);
    margin: var(--space-5) 0 0;
    padding: 0;
    list-style: none;
  }
  @media (min-width: 48rem) { .rule-list { grid-template-columns: repeat(3, 1fr); } }
  .rule-list li { padding: var(--space-5); }
  .num {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    margin-bottom: var(--space-3);
    background: var(--brand);
    color: var(--on-brand);
    border-radius: 14px;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.2em;
  }
  h3 { font-size: 1.25em; margin: 0 0 var(--space-2); }
  .rule-list p { margin: 0; color: var(--body-2); font-size: 0.98em; line-height: 1.5; }

  .printable {
    margin: var(--space-5) 0;
    padding: clamp(1.25rem, 3vw, 2rem);
    border: 2px dashed var(--border-strong);
    border-radius: var(--radius-lg);
    max-width: 34rem;
  }
  .p-brand { margin: 0 0 var(--space-4); font-family: var(--font-display); font-weight: 800; letter-spacing: -0.02em; }
  .p-label { margin: 0; font-size: var(--text-sm); font-weight: 700; color: var(--body-2); }
  .p-word {
    margin: var(--space-2) 0 var(--space-4);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: clamp(1.6rem, 4vw, 2.2rem);
    letter-spacing: -0.02em;
  }
  .p-rules { margin: 0 0 var(--space-4); padding-left: var(--space-5); display: grid; gap: var(--space-2); }
  .p-foot { margin: 0; color: var(--muted); font-size: var(--text-xs); }

  .next-actions { display: flex; flex-wrap: wrap; gap: var(--space-3); margin-top: var(--space-4); }

  /* Print: only the card, black on white, no navigation or controls. */
  @media print {
    :global(header),
    :global(footer),
    .head,
    .rules,
    .next,
    .generator,
    .gen-actions,
    :global(.btn),
    .p-brand { display: none !important; }
    section.card { box-shadow: none; border: 0; padding: 0; margin: 0; }
    .printable { border: 2px solid #000; max-width: none; }
    :global(body) { background: #fff; color: #000; }
  }
</style>
