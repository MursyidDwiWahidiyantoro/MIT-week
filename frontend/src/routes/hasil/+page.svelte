<script lang="ts">
  import { resolve } from '$app/paths';
  import { result } from '$lib/result.svelte';
  import { statusLabel, formatScore } from '$lib/format';
  import { tokenFromHash, verifyToken, buildShareUrl, type VerifyResult } from '$lib/verify';
  import StatusBadge from '$lib/components/ui/StatusBadge.svelte';
  import Panel from '$lib/components/ui/Panel.svelte';
  import EvidenceAccordion from '$lib/components/hasil/EvidenceAccordion.svelte';
  import HighlightedText from '$lib/components/hasil/HighlightedText.svelte';
  import ReplyCard from '$lib/components/hasil/ReplyCard.svelte';
  import ProofCardExporter from '$lib/components/hasil/ProofCardExporter.svelte';
  import ShareButton from '$lib/components/share/ShareButton.svelte';
  import tacticLabels from '$lib/data/tactics.json';
  import type { DossierParsed } from '$lib/api/schemas';

  let live = $state<DossierParsed | null>(result.value);
  let shared = $state<VerifyResult | null>(null);
  const shareOrigin = $derived.by(() =>
    typeof window === 'undefined' ? '' : window.location.origin
  );

  // A token in the fragment means this is a shared report, not a live session.
  $effect(() => {
    const token = tokenFromHash(window.location.hash);
    if (!token || live) return;
    const controller = new AbortController();
    void verifyToken(token, controller.signal).then((r) => (shared = r));
    return () => controller.abort();
  });

  const BASIS_LABELS: Record<string, string> = {
    ela: 'Analisis gambar (ELA)',
    fact: 'Pengecekan fakta',
    ai: 'Deteksi jejak AI',
    link: 'Pemeriksaan tautan',
    llm_untrusted: 'Catatan: teks dari AI tidak dipakai sebagai bukti'
  };

  const dossier = $derived(live);
  const shareUrl = $derived(
    live && shareOrigin ? buildShareUrl(shareOrigin, live.share_token) : ''
  );
</script>

<svelte:head><title>Hasil pemeriksaan — AMBA</title></svelte:head>

{#if shared && !shared.valid && shared.reason === 'unreachable'}
  <h1>Laporan belum bisa diperiksa</h1>
  <p class="lede">
    Server tidak dapat dihubungi, jadi laporan ini belum bisa dipastikan sah atau tidak.
    Periksa koneksi Anda lalu buka tautannya lagi.
  </p>
  <a class="cta" href={resolve('/deteksi')}>Periksa sendiri kontennya</a>
{:else if shared && !shared.valid}
  <h1>Laporan tidak sah</h1>
  <p class="lede">
    Tautan ini tidak dapat diverifikasi. Isinya mungkin sudah diubah, atau tautannya tidak lengkap.
  </p>
  <a class="cta" href={resolve('/deteksi')}>Periksa sendiri kontennya</a>
{:else if shared && shared.valid}
  <h1>Laporan terverifikasi</h1>
  <p class="verified" role="status">Terverifikasi AMBA</p>
  <p class="sentence">{shared.payload.title ?? statusLabel(shared.payload.status as never)}</p>
  <p class="meta">Skor risiko: {formatScore(shared.payload.score)}</p>
  <p class="note">
    Laporan ini hanya memuat ringkasan. Gambar dan teks aslinya tidak disimpan.
  </p>
{:else if !dossier}
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

  <div class="verdict" data-info-key="verdict.score">
    <StatusBadge status={dossier.verdict.status} score={dossier.verdict.score} />
    <p class="sentence" data-info-key="verdict.sentence">{dossier.verdict.explanation_simple}</p>
    <p class="explanation" data-info-key="verdict.explanation">{dossier.verdict.explanation}</p>
  </div>

  {#if dossier.tactics.length > 0}
    <Panel title="Pola yang perlu diwaspadai">
      <div data-info-key="tactics">
        <HighlightedText
          text={dossier.source_reference}
          tactics={dossier.tactics}
          labels={tacticLabels}
        />
      </div>
    </Panel>
  {/if}

  <Panel title="Rincian pemeriksaan">
    <EvidenceAccordion pillars={dossier.pillars} />
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
      Skor ini menunjukkan tingkat risiko, bukan tingkat keyakinan. Hasil bersifat indikatif dan bukan alat bukti hukum.
    </p>
  </Panel>

  <Panel title="Bagikan laporan">
    <div data-info-key="share">
      <ShareButton token={dossier.share_token} title={dossier.verdict.short_title} />
    </div>
  </Panel>

  <Panel title="Balas sopan">
    <div data-info-key="reply">
      <ReplyCard token={dossier.share_token} status={dossier.verdict.status} />
    </div>
  </Panel>

  <Panel title="Simpan kartu bukti">
    <div data-info-key="proofcard">
      <ProofCardExporter {dossier} {shareUrl} />
    </div>
  </Panel>

  <p class="meta">Skor risiko: {formatScore(dossier.verdict.score)}</p>
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
  .verified {
    display: inline-block; margin: 0 0 var(--space-3);
    padding: var(--space-1) var(--space-3);
    border: var(--border-width) solid var(--ok); border-radius: var(--radius-md);
    color: var(--ok); font-size: var(--text-sm); font-weight: 600;
  }
  .verdict { margin-bottom: var(--space-5); }
  .sentence { font-size: var(--text-xl); margin: var(--space-4) 0 var(--space-2); max-width: var(--measure); }
  .explanation { color: var(--text-2); max-width: var(--measure); margin: 0; }
  dl { display: grid; grid-template-columns: max-content 1fr; gap: var(--space-2) var(--space-4); margin: 0; }
  dt { color: var(--text-2); font-size: var(--text-sm); }
  dd { margin: 0; }
  .note { margin: var(--space-4) 0 0; color: var(--text-2); font-size: var(--text-sm); max-width: var(--measure); }
  .meta { color: var(--text-3); font-family: var(--font-mono); font-size: var(--text-xs); margin: var(--space-2) 0 0; }
</style>
