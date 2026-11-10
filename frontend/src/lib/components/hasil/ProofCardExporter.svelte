<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { statusLabel, formatScore } from '$lib/format';
  import { simpleSentence } from '$lib/i18n/verdict';
  import type { DossierParsed } from '$lib/api/schemas';

  interface Props {
    dossier: DossierParsed;
    /** Empty for bundled demos, which have no signed share token. */
    shareUrl?: string;
  }

  let { dossier, shareUrl = '' }: Props = $props();

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

      if (shareUrl) {
        const qrDataUrl = await QRCode.toDataURL(shareUrl, { margin: 1, width: 160 });
        const qrImg = card.querySelector<HTMLImageElement>('[data-qr]');
        if (qrImg) qrImg.src = qrDataUrl;
        // Wait for the QR image to decode before rasterising.
        if (qrImg) await qrImg.decode().catch(() => undefined);
      }

      const dataUrl = await toPng(card, { width: 1200, height: 630, pixelRatio: 1 });
      const link = document.createElement('a');
      link.download = `amba-${dossier.id}.png`;
      link.href = dataUrl;
      link.click();
    } catch {
      error = t('proof.error');
    } finally {
      busy = false;
    }
  }
</script>

<div class="exporter">
  <Button variant="quiet" onclick={exportPng} loading={busy}>
    <Icon name="download" size={18} /> {t('proof.save')}
  </Button>
  {#if error}<p class="error-text" role="alert">{error}</p>{/if}

  <!-- The exported surface. Fixed 1200x630, off-screen but rendered. -->
  <div class="card" bind:this={card} aria-hidden="true">
    <div class="card-top">
      <p class="brand">AMBA</p>
      <p class="stamp">{dossier.is_demo ? t('proof.stampDemo') : t('proof.stamp')}</p>
    </div>

    <p class="status">{statusLabel(dossier.verdict.status)}</p>
    <p class="score">
      {formatScore(dossier.verdict.score)}
    </p>
    <p class="sentence">{simpleSentence(dossier)}</p>

    <div class="card-bottom">
      <p class="disclaimer">
        {t('proof.disclaimer')}
      </p>
      {#if shareUrl}
        <img data-qr alt="" width="160" height="160" />
      {:else}
        <div class="qr-placeholder">{t('proof.example')}<br />{t('proof.noLink')}</div>
      {/if}
    </div>
  </div>
</div>

<style>
  .exporter { display: grid; gap: var(--space-3); justify-items: start; }

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
    background: #070302;
    color: #fff8f4;
    border: 10px solid #ff5a1f;
    font-family: var(--font-sans);
  }
  .card-top { display: flex; justify-content: space-between; align-items: baseline; }
  .brand {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    letter-spacing: -0.03em;
    font-size: 34px;
  }
  .stamp { margin: 0; font-size: 18px; color: #bfb4ae; font-weight: 700; }
  .status {
    margin: 0;
    align-self: flex-start;
    padding: 8px 22px;
    background: #ff5a1f;
    border: 4px solid #ff5a1f;
    border-radius: 99px;
    color: #140800;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 30px;
  }
  .score { margin: 18px 0 0; font-size: 26px; color: #d4c9c3; font-weight: 700; }
  .sentence { margin: 18px 0 0; font-size: 32px; line-height: 1.3; max-width: 820px; font-weight: 700; }
  .card-bottom { display: flex; justify-content: space-between; align-items: flex-end; gap: 40px; }
  .disclaimer { margin: 0; font-size: 18px; color: #bfb4ae; max-width: 820px; line-height: 1.4; }
  .card img { background: #fff8f4; border: 3px solid #ff5a1f; border-radius: 8px; }
  .qr-placeholder {
    width: 160px;
    height: 160px;
    display: grid;
    place-items: center;
    border: 3px dashed #796154;
    border-radius: 8px;
    color: #bfb4ae;
    font-size: 14px;
    font-weight: 700;
    line-height: 1.4;
    text-align: center;
  }
</style>
