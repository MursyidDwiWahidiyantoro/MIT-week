<script lang="ts">
  import Panel from '$lib/components/ui/Panel.svelte';
</script>

<svelte:head>
  <title>Tentang — AMBA</title>
  <meta
    name="description"
    content="AMBA adalah alat bantu literasi digital. Hasil analisis bersifat indikatif, bukan bukti pasti."
  />
</svelte:head>

<h1>Tentang AMBA</h1>
<p class="lede">
  AMBA (Aplikasi Mengenali &amp; Bertindak Aman) membantu Anda memeriksa tautan,
  tangkapan layar, dan pesan yang mencurigakan sebelum menyebarkannya.
</p>

<Panel title="Apa yang kami kirim ke layanan AI">
  <dl>
    <dt>Dikirim</dt>
    <dd>
      Isi yang Anda periksa — tautan, teks, atau gambar — dikirim ke model bahasa
      untuk dinilai. Pemeriksaan fakta memakai pencarian ke sumber terpercaya.
    </dd>
    <dt>Tidak dikirim</dt>
    <dd>
      Nama, akun, alamat IP, atau riwayat pemeriksaan Anda. AMBA tidak punya akun
      dan tidak menyimpan input.
    </dd>
    <dt>Disimpan</dt>
    <dd>
      Tidak ada. Hasil hanya tersimpan selama halaman hasil terbuka. Satu-satunya
      pengecualian: saat Anda membagikan konten lewat menu bagikan ponsel, item
      ditahan beberapa detik di perangkat Anda lalu dihapus.
    </dd>
  </dl>
</Panel>

<Panel title="Batas yang perlu Anda tahu">
  <ul>
    <li>Skor 0–100 menunjukkan tingkat risiko, bukan tingkat keyakinan.</li>
    <li>Analisis gambar (ELA) melemah pada PNG dan tangkapan layar.</li>
    <li>Pemeriksaan fakta hanya memakai sumber yang sudah kami tinjau.</li>
    <li>Hasil tidak dapat dipakai sebagai alat bukti hukum.</li>
    <li>
      Menu bagikan ke AMBA hanya tersedia di peramban yang mendukung Web Share
      Target (misalnya Chrome di Android). Di iOS, gunakan tombol tempel.
    </li>
  </ul>
</Panel>

<Panel title="Prinsip kami">
  <ul>
    <li>Tidak mengklaim kepastian. Kami memakai kata “terindikasi”, bukan “pasti”.</li>
    <li>Tidak mengarang sumber. Sumber yang tidak terverifikasi tidak ditampilkan sebagai bukti.</li>
    <li>Tidak menyimpan data Anda. Tidak ada akun dan tidak ada basis data.</li>
    <li>Tidak menghakimi orang. Hasil membahas kontennya, bukan pengirimnya.</li>
  </ul>
</Panel>

<p class="disclaimer">
  Hasil analisis bersifat indikatif dan bukan alat bukti hukum. Untuk keputusan
  penting, konfirmasikan ke pihak berwenang atau sumber resmi.
</p>

<style>
  h1 { font-size: var(--text-3xl); margin: 0 0 var(--space-4); }
  .lede { color: var(--text-2); max-width: var(--measure); margin: 0 0 var(--space-6); }
  dl { display: grid; grid-template-columns: max-content 1fr; gap: var(--space-3) var(--space-5); margin: 0; }
  dt { color: var(--text-2); font-size: var(--text-sm); font-weight: 600; }
  dd { margin: 0; max-width: var(--measure); }
  ul { margin: 0; padding-left: var(--space-5); display: grid; gap: var(--space-2); max-width: var(--measure); }
  .disclaimer {
    margin-top: var(--space-6); padding: var(--space-4);
    border: var(--border-width) solid var(--border); border-radius: var(--radius-md);
    color: var(--text-2); font-size: var(--text-sm); max-width: var(--measure);
  }
  @media (max-width: 40rem) {
    dl { grid-template-columns: 1fr; }
  }
</style>
