/// <reference lib="webworker" />
/**
 * Minimal service worker for the PWA share target.
 *
 * It does not precache the app (the SPA is small and static hosting already
 * caches assets). Its only job is to accept the OS share POST, hold the item
 * for a few seconds in Cache Storage, and redirect into /deteksi.
 */
import {
  SHARE_CACHE,
  SHARE_MAX_AGE_MS,
  firstHttpUrl,
  mediaKey,
  shareKey,
  validateShareFile,
  type ShareMeta
} from './lib/share-target';

const sw = self as unknown as ServiceWorkerGlobalScope;

sw.addEventListener('install', (event) => {
  event.waitUntil(sw.skipWaiting());
});

sw.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      await sw.clients.claim();
      await sweep();
    })()
  );
});

sw.addEventListener('message', (event) => {
  if (event.data?.type === 'amba-sweep-share') event.waitUntil(sweep());
});

sw.addEventListener('fetch', (event) => {
  const request = event.request;
  if (request.method !== 'POST') return;

  const url = new URL(request.url);
  if (!url.pathname.endsWith('/share-target')) return;

  event.respondWith(handleShare(request));
});

async function handleShare(request: Request): Promise<Response> {
  try {
    const form = await request.formData();
    const title = String(form.get('title') ?? '').slice(0, 300);
    const text = String(form.get('text') ?? '').slice(0, 5000);
    const url = String(form.get('url') ?? '').slice(0, 2048);

    const media = form.get('media');
    const file = media instanceof File ? media : null;
    if (file) {
      const problem = validateShareFile(file);
      if (problem) return redirect('invalid');
    }

    const resolvedUrl = url || firstHttpUrl(text) || '';
    if (!resolvedUrl && !text && !file) return redirect('empty');

    const id = crypto.randomUUID();
    const meta: ShareMeta = {
      ts: Date.now(),
      url: resolvedUrl || undefined,
      text: text || undefined,
      title: title || undefined,
      hasFile: !!file,
      fileName: file?.name,
      fileType: file?.type
    };

    const cache = await caches.open(SHARE_CACHE);
    await cache.put(
      shareKey(id),
      new Response(JSON.stringify(meta), { headers: { 'Content-Type': 'application/json' } })
    );
    if (file) {
      await cache.put(
        mediaKey(id),
        new Response(file, { headers: { 'Content-Type': file.type } })
      );
    }

    return redirect(`deteksi?share=${encodeURIComponent(id)}`);
  } catch {
    return redirect('invalid');
  }
}

function redirect(path: string): Response {
  return Response.redirect(new URL(path, sw.registration.scope).href, 303);
}

async function sweep(): Promise<void> {
  try {
    const cache = await caches.open(SHARE_CACHE);
    const keys = await cache.keys();
    const ids = new Set<string>();

    for (const request of keys) {
      const match = new URL(request.url).pathname.match(/\/__share\/([^/]+)$/);
      if (match) ids.add(match[1]);
    }

    for (const id of ids) {
      const response = await cache.match(shareKey(id));
      if (!response) {
        await cache.delete(mediaKey(id));
        continue;
      }
      const meta = (await response.json()) as ShareMeta;
      if (Date.now() - meta.ts > SHARE_MAX_AGE_MS) {
        await cache.delete(shareKey(id));
        await cache.delete(mediaKey(id));
      }
    }
  } catch {
    // A failed sweep must never break the app.
  }
}
