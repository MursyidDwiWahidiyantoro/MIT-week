<script lang="ts">
  import { resolve } from '$app/paths';
  import { result } from '$lib/result.svelte';
  import { statusLabel, formatScore } from '$lib/format';
  import StatusBadge from '$lib/components/ui/StatusBadge.svelte';
  import Panel from '$lib/components/ui/Panel.svelte';

  const dossier = result.value;

  const BASIS_LABELS: Record<string, string> = {
    ela: 'Analisis gambar (ELA)',
    fact: 'Pengecekan fakta',
    ai: 'Deteksi jejak AI',
    link: 'Pemeriksaan tautan',
    llm_untrusted: 'Catatan: teks dari AI tidak dipakai sebagai bukti'
  };
</script>

<svelte:head><title>Hasil pemeriksaan — AMBA</title></svelte:head>

{#if !dossier}
  <h1>Hasil tidak tersimpan</h1>
  <p class="lede">Hasil pemeriksaan hanya tersimpan selama halaman ini terbuka. Silakan periksa lagi.</p>
  <a class="cta" href={resolve('/deteksi')}>Periksa lagi</a>
{:else}
  <h1>Hasil pemeriksaan</h1>

  {#if dossier.is_demo}
    <p class="demo" role="note">CONTOH / SIMULASI — ini bukan hasil pemeriksaan sungguhan.</p>
  {/if}

  {#if dossier.degraded}
    <p class="degraded" role="note">
      Sebagian pemeriksaan tidak berjalan. Hasil ini tetap berguna, tetapi belum lengkap.
    </p>
  {/if}

  <div data-info-key="verdict.score">
    <StatusBadge status={dossier.verdict.status} score={dossier.verdict.score} />
  </div>

  <p class="sentence" data-info-key="verdict.sentence">{dossier.verdict.explanation_simple}</p>

  <Panel title="Apa yang ditemukan">
    <p data-info-key="verdict.explanation">{dossier.verdict.explanation}</p>
  </Panel>

  <Panel title="Batas pemeriksaan ini">
    <dl data-info-key="verdict.limits">
      <dt>Berdasarkan</dt>
      <dd>{dossier.verdict.basis.map((b) => BASIS_LABELS[b] ?? b).join(', ') || '—'}</dd>
      <dt>Tidak dapat diperiksa</dt>
      <dd>
        {dossier.verdict.not_checked.map((b) => BASIS_LABELS[b] ?? b).join(', ') || 'Semua pemeriksaan berjalan'}
      </dd>
    </dl>
    <p class="note">
      Skor ini menunjukkan tingkat risiko, bukan tingkat keyakinan. Hasil bersifat indikatif
      dan bukan alat bukti hukum.
    </p>
  </Panel>

  <p class="meta">Skor risiko: {formatScore(dossier.verdict.score)}</p>
  <p class="meta">Status: {statusLabel(dossier.verdict.status)}</p>
{/if}

<style>
  h1 { font-size: var(--text-3xl); margin: 0 0 var(--space-4); }
  .lede { color: var(--text-2); max-width: var(--measure); }
  .cta {
    display: inline-flex; align-items: center; min-height: 44px;
    padding: 0 var(--space-5); margin-top: var(--space-4);
    background: var(--accent); color: #10161d;
    border-radius: var(--radius-md); text-decoration: none; font-weight: 600;
  }
  .demo, .degraded {
    border: var(--border-width) solid var(--warn); border-radius: var(--radius-md);
    padding: var(--space-3); color: var(--warn); font-size: var(--text-sm);
  }
  .sentence { font-size: var(--text-xl); margin: var(--space-4) 0; max-width: var(--measure); }
  dl { display: grid; grid-template-columns: max-content 1fr; gap: var(--space-2) var(--space-4); margin: 0; }
  dt { color: var(--text-2); font-size: var(--text-sm); }
  dd { margin: 0; }
  .note { margin: var(--space-4) 0 0; color: var(--text-2); font-size: var(--text-sm); max-width: var(--measure); }
  .meta { color: var(--text-3); font-family: var(--font-mono); font-size: var(--text-xs); margin: var(--space-2) 0 0; }
</style>
