<script lang="ts">
  import { resolve } from '$app/paths';
  // Token definitions must load before any rule that consumes them.
  import '../lib/styles/tokens.css';
  import '../app.css';
  import '@fontsource-variable/geist';
  import '@fontsource-variable/geist-mono';

  let { children } = $props();
</script>

<a class="skip-link" href="#main">Langsung ke isi utama</a>

<header class="site-header">
  <nav aria-label="Navigasi utama">
    <a class="wordmark" href={resolve('/')}>AMBA</a>
    <ul>
      <li><a href={resolve('/deteksi')}>Periksa</a></li>
      <li><a href={resolve('/edukasi')}>Edukasi</a></li>
      <li><a href={resolve('/tentang')}>Tentang</a></li>
    </ul>
  </nav>
</header>

<main id="main" tabindex="-1">
  {@render children()}
</main>

<footer class="site-footer">
  <p>
    AMBA membantu Anda menilai risiko informasi. Hasilnya <strong>indikatif</strong>,
    bukan bukti pasti dan bukan alat bukti hukum.
  </p>
</footer>

<style>
  .site-header {
    border-bottom: var(--border-width) solid var(--border);
    background: var(--bg);
  }
  nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-5);
    max-width: 1100px;
    margin: 0 auto;
    padding: var(--space-4);
  }
  .wordmark {
    font-family: var(--font-mono);
    font-size: var(--text-lg);
    letter-spacing: 0.12em;
    color: var(--text);
    text-decoration: none;
  }
  ul {
    display: flex;
    gap: var(--space-5);
    margin: 0;
    padding: 0;
    list-style: none;
  }
  nav a:not(.wordmark) {
    color: var(--text-2);
    text-decoration: none;
    font-size: var(--text-sm);
  }
  nav a:not(.wordmark):hover {
    color: var(--text);
  }
  main {
    max-width: 1100px;
    margin: 0 auto;
    padding: var(--space-7) var(--space-4) var(--space-9);
  }
  main:focus {
    outline: none;
  }
  .site-footer {
    border-top: var(--border-width) solid var(--border);
    padding: var(--space-5) var(--space-4);
    color: var(--text-2);
    font-size: var(--text-sm);
  }
  .site-footer p {
    max-width: var(--measure);
    margin: 0 auto;
  }
</style>
