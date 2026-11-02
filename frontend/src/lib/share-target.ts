/**
 * PWA share-target handoff.
 *
 * The service worker writes one shared item into Cache Storage and redirects
 * to /deteksi?share=<id>. The page reads and deletes it immediately. This is
 * the only place user content is ever written to client storage, and it lives
 * for seconds, never more than five minutes.
 */
export const SHARE_CACHE = 'amba-share-v1';
export const SHARE_MAX_AGE_MS = 5 * 60 * 1000;
export const SHARE_MAX_BYTES = 10 * 1024 * 1024;
export const SHARE_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp'] as const;

export interface SharedItem {
  url?: string;
  text?: string;
  file: File | null;
}

export interface ShareMeta {
  ts: number;
  url?: string;
  text?: string;
  title?: string;
  hasFile: boolean;
  fileName?: string;
  fileType?: string;
}

export function firstHttpUrl(text: string): string | null {
  const match = text.match(/https?:\/\/[^\s<>"']+/i);
  if (!match) return null;
  return match[0].replace(/[.,;:!?)]+$/, '');
}

export function isAllowedImageType(type: string): boolean {
  return (SHARE_IMAGE_TYPES as readonly string[]).includes(type);
}

export function validateShareFile(file: File): string | null {
  if (!isAllowedImageType(file.type)) {
    return 'Jenis berkas ini belum didukung. Gunakan JPG, PNG, atau WebP.';
  }
  if (file.size > SHARE_MAX_BYTES) {
    return 'Berkas terlalu besar. Maksimal 10 MB.';
  }
  return null;
}

export function shareKey(id: string): string {
  return new URL(`/__share/${id}`, globalThis.location?.origin ?? 'http://localhost').href;
}

export function mediaKey(id: string): string {
  return `${shareKey(id)}/media`;
}

/** Reads and deletes one shared item. Returns null when it is gone. */
export async function readSharedItem(id: string): Promise<SharedItem | null> {
  if (!/^[a-zA-Z0-9-]{8,64}$/.test(id) || !('caches' in globalThis)) return null;

  try {
    const cache = await caches.open(SHARE_CACHE);
    const key = shareKey(id);
    const response = await cache.match(key);
    if (!response) return null;

    const meta = (await response.json()) as ShareMeta;
    let file: File | null = null;
    if (meta.hasFile) {
      const media = await cache.match(mediaKey(id));
      if (media) {
        const blob = await media.blob();
        file = new File([blob], meta.fileName ?? 'gambar', {
          type: meta.fileType ?? blob.type
        });
      }
    }

    await cache.delete(key);
    await cache.delete(mediaKey(id));
    return { url: meta.url, text: meta.text, file };
  } catch {
    return null;
  }
}
