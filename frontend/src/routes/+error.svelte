<script lang="ts">
  import { page } from '$app/state';
  import { resolve } from '$app/paths';
</script>

<svelte:head><title>Halaman tidak ditemukan — AMBA</title></svelte:head>

<section class="state">
  <p class="eyebrow">{page.status === 404 ? '404' : 'Kendala'}</p>
  <h1 class="h1">{page.status === 404 ? 'Halaman tidak ditemukan' : 'Ada kendala di halaman ini'}</h1>
  <p class="lede">
    {#if page.status === 404}
      Alamat ini tidak ada. Mungkin tautannya salah ketik atau sudah dipindahkan.
    {:else}
      Halaman ini gagal dimuat. Coba muat ulang; kalau masih gagal, kembali ke
      beranda dan mulai lagi.
    {/if}
  </p>

  <div class="actions">
    <a class="btn btn--primary" href={resolve('/')}>Kembali ke beranda</a>
    <a class="btn btn--quiet" href={resolve('/deteksi')}>Periksa konten</a>
  </div>
</section>

<style>
  .state { max-width: 44rem; padding-block: var(--space-7); }
  h1 { margin-bottom: var(--space-4); }
  .lede { margin-bottom: var(--space-6); }
  .actions { display: flex; gap: var(--space-3); flex-wrap: wrap; }
</style>
