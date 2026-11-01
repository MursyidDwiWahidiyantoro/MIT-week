/// <reference lib="webworker" />
/**
 * Client-side ELA preview.
 *
 * The authoritative score comes from the server (internal/forensics); this
 * worker only produces a visual heatmap so the result page has something to
 * show before/without the server image. It never sets a score.
 */

const QUALITY = 0.9;
const MAX_SIDE = 4096;
const BAND_HEIGHT = 32;

export interface ElaRequest {
  bitmap: ImageBitmap;
}

export interface ElaResponse {
  type: 'done' | 'progress' | 'error';
  bitmap?: ImageBitmap;
  progress?: number;
  downscaled?: boolean;
  error?: string;
}

self.onmessage = async (event: MessageEvent<ElaRequest>) => {
  const { bitmap } = event.data;
  try {
    const scale = Math.min(1, MAX_SIDE / Math.max(bitmap.width, bitmap.height));
    const w = Math.max(1, Math.round(bitmap.width * scale));
    const h = Math.max(1, Math.round(bitmap.height * scale));
    const downscaled = scale < 1;

    const src = new OffscreenCanvas(w, h);
    const srcCtx = src.getContext('2d');
    if (!srcCtx) throw new Error('canvas unavailable');
    srcCtx.drawImage(bitmap, 0, 0, w, h);
    const original = srcCtx.getImageData(0, 0, w, h);

    // Re-encode as JPEG, then decode again: the difference is the ELA signal.
    const blob = await src.convertToBlob({ type: 'image/jpeg', quality: QUALITY });
    const reencodedBitmap = await createImageBitmap(blob);
    const reCanvas = new OffscreenCanvas(w, h);
    const reCtx = reCanvas.getContext('2d');
    if (!reCtx) throw new Error('canvas unavailable');
    reCtx.drawImage(reencodedBitmap, 0, 0);
    const reencoded = reCtx.getImageData(0, 0, w, h);

    const out = new OffscreenCanvas(w, h);
    const outCtx = out.getContext('2d');
    if (!outCtx) throw new Error('canvas unavailable');
    const heat = outCtx.createImageData(w, h);

    let processed = 0;
    for (let y = 0; y < h; y += BAND_HEIGHT) {
      const end = Math.min(h, y + BAND_HEIGHT);
      for (let i = y * w * 4; i < end * w * 4; i += 4) {
        const dr = Math.abs(original.data[i] - reencoded.data[i]);
        const dg = Math.abs(original.data[i + 1] - reencoded.data[i + 1]);
        const db = Math.abs(original.data[i + 2] - reencoded.data[i + 2]);
        const d = Math.min(255, (dr + dg + db) * 3);
        heat.data[i] = d;
        heat.data[i + 1] = Math.max(0, d - 90);
        heat.data[i + 2] = 0;
        heat.data[i + 3] = 255;
      }
      processed = end;
      (self as unknown as Worker).postMessage({
        type: 'progress',
        progress: Math.round((processed / h) * 100)
      } satisfies ElaResponse);
      // Yield so the worker stays responsive on large images.
      await new Promise((r) => setTimeout(r, 0));
    }

    outCtx.putImageData(heat, 0, 0);
    const resultBitmap = out.transferToImageBitmap();
    reencodedBitmap.close();
    bitmap.close();

    (self as unknown as Worker).postMessage(
      { type: 'done', bitmap: resultBitmap, downscaled } satisfies ElaResponse,
      [resultBitmap]
    );
  } catch (err) {
    (self as unknown as Worker).postMessage({
      type: 'error',
      error: err instanceof Error ? err.message : 'unknown'
    } satisfies ElaResponse);
  }
};
