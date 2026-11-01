<script lang="ts">
  import Button from '$lib/components/ui/Button.svelte';
  import { statusLabel } from '$lib/format';
  import type { DossierParsed } from '$lib/api/schemas';
  import { Download } from '@lucide/svelte';

  interface Props {
    dossier: DossierParsed;
    shareUrl: string;
  }

  let { dossier, shareUrl }: Props = $props();

  let busy = $state(false);
  let error = $state<string | null>(null);
  let card: HTMLDivElement | undefined = $state();

  async function exportPng(): Promise<void> {
    if (!card) return;
    busy = true;
    error = null;
    try {
      // Both libraries are imported on first use so they never enter the
      // initial bundle.
      const [{ toPng }, QRCode] = await Promise.all([import('html-to-image'), import('qrcode')]);
      const qrDataUrl = await QRCode.toDataURL(shareUrl, { margin: 1, width: 160 });

      const qrImg = card.querySelector<HTMLImageElement>('[data-qr]');
      if (qrImg) qrImg.src = qrDataUrl;

      // Wait for the QR image to decode before rasterising.
      if (qrImg) await qrImg.decode().catch(() => undefined);

      const dataUrl = await toPng(card, { width: 1200, height: 630, pixelRatio: 1 });
      const link = document.createElement('a');
      link.download = `amba-${dossier.id}.png`;
      link.href = dataUrl;
      link.click();
    } catch {
      error = 'Kartu bukti tidak bisa dibuat di peramban ini. Coba lagi atau gunakan tangkapan layar.';
    } finally {
      busy = false;
    }
  }
</script>

<div class="exporter">
  <Button variant="quiet" onclick={exportPng} loading={busy}>
    <Download size={16} aria-hidden="true" /> Simpan kartu bukti
  </Button>
  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <!-- The exported surface. Fixed 1200x630, off-screen but rendered. -->
  <div class="card" bind:this={card} aria-hidden="true">
    <div class="card-top">
      <p class="brand">AMBA</p>
      <p class="stamp">{dossier.is_demo ? 'CONTOH / SIMULASI' : 'LAPORAN PEMERIKSAAN'}</p>
    </div>

    <p class="status">{statusLabel(dossier.verdict.status)}</p>
    <p class="score">
      {dossier.verdict.score === null ? 'Belum ada skor' : `Skor risiko ${dossier.verdict.score} dari 100`}
    </p>
    <p class="sentence">{dossier.verdict.explanation_simple}</p>

    <div class="card-bottom">
      <p class="disclaimer">
        INDIKATIF — bukan alat bukti hukum. Skor menunjukkan risiko, bukan tingkat keyakinan.
      </p>
      <img data-qr alt="" width="160" height="160" />
    </div>
  </div>
</div>

<style>
  .exporter { display: grid; gap: var(--space-2); justify-items: start; }
  .error { margin: 0; color: var(--err); font-size: var(--text-sm); max-width: var(--measure); }

  /* Kept in the layout flow but visually hidden: html-to-image needs a laid-out
     element with real dimensions. */
  .card {
    position: fixed;
    left: -10000px;
    top: 0;
    width: 1200px;
    height: 630px;
    padding: 56px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    background: var(--bg);
    color: var(--text);
    font-family: var(--font-sans);
  }
  .card-top { display: flex; justify-content: space-between; align-items: baseline; }
  .brand { margin: 0; font-family: var(--font-mono); letter-spacing: 0.14em; font-size: 28px; }
  .stamp { margin: 0; font-family: var(--font-mono); font-size: 18px; color: var(--text-2); }
  .status { margin: 0; font-size: 52px; font-weight: 600; line-height: 1.1; }
  .score { margin: 8px 0 0; font-size: 26px; color: var(--text-2); }
  .sentence { margin: 24px 0 0; font-size: 30px; line-height: 1.35; max-width: 780px; }
  .card-bottom { display: flex; justify-content: space-between; align-items: flex-end; gap: 40px; }
  .disclaimer { margin: 0; font-size: 18px; color: var(--text-2); max-width: 820px; line-height: 1.4; }
  .card img { background: #fff; border-radius: 8px; }
</style>
