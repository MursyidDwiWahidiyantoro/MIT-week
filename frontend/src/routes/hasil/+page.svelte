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
  import ReadAloud from '$lib/components/a11y/ReadAloud.svelte';
  import ElaViewer from '$lib/components/hasil/ElaViewer.svelte';
  import tacticLabels from '$lib/data/tactics.json';
  import type { DossierParsed } from '$lib/api/schemas';

  let live = $state<DossierParsed | null>(result.value);
  let imageFile = $state<File | null>(result.imageFile);
  let shared = $state<VerifyResult | null>(null);
  let checking = $state(false);
  let verdictHeading: HTMLElement | null = $state(null);

  const shareOrigin = $derived.by(() =>
    typeof window === 'undefined' ? '' : window.location.origin
  );

  // A token in the fragment means this is a shared report, not a live session.
  $effect(() => {
    const token = tokenFromHash(window.location.hash);
    if (!token || live) return;
    const controller = new AbortController();
    checking = true;
    void verifyToken(token, controller.signal)
      .then((r) => (shared = r))
      .finally(() => (checking = false));
    return () => controller.abort();
  });

  // Move focus to the verdict once it exists, so a screen reader lands on the
  // answer instead of staying on whatever was focused during the scan.
  $effect(() => {
    if (live && verdictHeading) verdictHeading.focus({ preventScroll: true });
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
  // A demo has no signed token, so its share and reply panels cannot work.
  const canShare = $derived(!!live && live.share_token.length > 0);
  const elaAnomalies = $derived(
    live && live.pillars.ela.available ? live.pillars.ela.anomalies : []
  );
</script>

<svelte:head><title>Hasil pemeriksaan — AMBA</title></svelte:head>

{#if checking}
  <h1>Memeriksa laporan…</h1>
  <p class="lede" aria-live="polite">Mencocokkan tautan ini dengan laporan yang ditandatangani AMBA.</p>
{:else if shared && !shared.valid && shared.reason === 'invalid'}
  <h1>Laporan tidak sah</h1>
  <p class="lede">
    Tautan ini tidak dapat diverifikasi. Isinya mungkin sudah diubah, atau tautannya tidak lengkap.
  </p>
  <a class="cta" href={resolve('/deteksi')}>Periksa sendiri kontennya</a>
{:else if shared && !shared.valid && shared.reason === 'unreachable'}
  <h1>Laporan belum bisa diperiksa</h1>
  <p class="lede">
    Server tidak dapat dihubungi, jadi laporan ini belum bisa dipastikan sah atau tidak.
    Periksa koneksi Anda lalu buka tautannya lagi.
  </p>
  <a class="cta" href={resolve('/deteksi')}>Periksa sendiri kontennya</a>
{:else if shared && !shared.valid}
  <h1>Laporan sedang tidak bisa diperiksa</h1>
  <p class="lede">
    Layanan verifikasi sedang sibuk atau bermasalah. Ini bukan tanda laporan Anda palsu.
    {#if shared.reason === 'unavailable' && shared.retryAfter}
      Coba lagi dalam {shared.retryAfter} detik.
    {:else}
      Coba lagi beberapa saat lagi.
    {/if}
  </p>
  <button class="cta" type="button" onclick={() => window.location.reload()}>Coba lagi</button>
{:else if shared && shared.valid && shared.payload.preview}
  <h1>Laporan contoh</h1>
  <p class="verified preview" role="status">Mode pratinjau</p>
  <p class="sentence">{shared.payload.title ?? statusLabel(shared.payload.status as never)}</p>
  <p class="meta">Skor risiko: {formatScore(shared.payload.score)}</p>
  <p class="note">
    Laporan ini dibuat di perangkat Anda dari data contoh. Ini bukan laporan
    yang ditandatangani server dan bukan hasil pemeriksaan sungguhan.
  </p>
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

  <div class="verdict">
    <h2 class="verdict-heading" tabindex="-1" bind:this={verdictHeading} data-info-key="verdict.score">
      <StatusBadge status={dossier.verdict.status} score={dossier.verdict.score} />
    </h2>
    <p class="sentence" data-info-key="verdict.sentence">{dossier.verdict.explanation_simple}</p>
    <p class="explanation" data-info-key="verdict.explanation">{dossier.verdict.explanation}</p>
    <div class="read-aloud" data-info-key="readaloud">
      <ReadAloud text={`${dossier.verdict.explanation_simple} ${dossier.verdict.explanation}`} />
    </div>
  </div>

  {#if imageFile && dossier.input_type === 'FILE'}
    <Panel title="Tampilan gambar">
      <ElaViewer file={imageFile} anomalies={elaAnomalies} />
    </Panel>
  {/if}

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
      {#if canShare}
        <ShareButton token={dossier.share_token} title={dossier.verdict.short_title} />
      {:else}
        <p class="note" role="note">
          Data contoh ini tidak punya tautan bertanda tangan, jadi belum bisa dibagikan.
          Balas sopan dan kartu bukti tetap tersedia.
        </p>
      {/if}
    </div>
  </Panel>

  <Panel title="Balas sopan">
    <div data-info-key="reply">
      <ReplyCard token={dossier.share_token} status={dossier.verdict.status} demo={dossier.is_demo} />
    </div>
  </Panel>

  <Panel title="Simpan kartu bukti">
    <div data-info-key="proofcard">
      <ProofCardExporter {dossier} {shareUrl} />
    </div>
  </Panel>
{/if}

<style>
  h1 { font-size: var(--text-3xl); margin: 0 0 var(--space-4); }
  .lede { color: var(--text-2); max-width: var(--measure); }
  .cta {
    display: inline-flex; align-items: center; min-height: 44px;
    padding: 0 var(--space-5); margin-top: var(--space-4);
    background: var(--accent); color: #10161d;
    border: 0; border-radius: var(--radius-md);
    text-decoration: none; font-weight: 600; font: inherit; cursor: pointer;
  }
  .cta:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
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
  .verified.preview { border-color: var(--warn); color: var(--warn); }
  .verdict { margin-bottom: var(--space-5); }
  .verdict-heading { margin: 0; font-size: inherit; line-height: inherit; }
  .verdict-heading:focus { outline: none; }
  .verdict-heading:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .sentence { font-size: var(--text-xl); margin: var(--space-4) 0 var(--space-2); max-width: var(--measure); }
  .explanation { color: var(--text-2); max-width: var(--measure); margin: 0; }
  .read-aloud { margin-top: var(--space-3); }
  dl { display: grid; grid-template-columns: max-content 1fr; gap: var(--space-2) var(--space-4); margin: 0; }
  dt { color: var(--text-2); font-size: var(--text-sm); }
  dd { margin: 0; }
  .note { margin: var(--space-4) 0 0; color: var(--text-2); font-size: var(--text-sm); max-width: var(--measure); }
  .meta { color: var(--text-3); font-family: var(--font-mono); font-size: var(--text-xs); margin: var(--space-2) 0 0; }
</style>
