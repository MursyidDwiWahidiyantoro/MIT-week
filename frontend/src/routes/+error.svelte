<script lang="ts">
  import { page } from '$app/state';
  import { resolve } from '$app/paths';
</script>

<svelte:head><title>Halaman tidak ditemukan — AMBA</title></svelte:head>

<h1>{page.status === 404 ? 'Halaman tidak ditemukan' : 'Ada kendala di halaman ini'}</h1>
<p class="lede">
  {#if page.status === 404}
    Alamat ini tidak ada. Mungkin tautannya salah ketik atau sudah dipindahkan.
  {:else}
    Halaman ini gagal dimuat. Coba muat ulang; kalau masih gagal, kembali ke
    beranda dan mulai lagi.
  {/if}
</p>

<div class="actions">
  <a class="cta" href={resolve('/')}>Kembali ke beranda</a>
  <a class="quiet" href={resolve('/deteksi')}>Periksa konten</a>
</div>

<style>
  h1 { font-size: var(--text-3xl); margin: 0 0 var(--space-3); }
  .lede { color: var(--text-2); max-width: var(--measure); margin: 0 0 var(--space-5); }
  .actions { display: flex; gap: var(--space-3); flex-wrap: wrap; }
  .cta {
    display: inline-flex; align-items: center; min-height: 44px;
    padding: 0 var(--space-5);
    background: var(--accent); color: #10161d;
    border-radius: var(--radius-md); text-decoration: none; font-weight: 600;
  }
  .quiet {
    display: inline-flex; align-items: center; min-height: 44px;
    padding: 0 var(--space-5);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-md); text-decoration: none; color: var(--text);
  }
</style>
