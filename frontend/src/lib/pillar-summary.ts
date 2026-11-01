import type { PillarKey } from './audit.svelte';

/**
 * One honest line describing a pillar payload the server has already sent.
 *
 * Kept out of the Svelte component so the mapping can be tested without a
 * timing race: in mock mode the stream finishes in under a second, so an e2e
 * test can never reliably observe a mid-stream frame.
 *
 * Returns null when there is nothing truthful to say yet.
 */
export function pillarSummary(key: PillarKey, data: unknown): string | null {
  if (!data || typeof data !== 'object') return null;
  const d = data as Record<string, unknown>;
  if (d.available === false) return 'Tidak dapat diperiksa untuk konten ini.';

  switch (key) {
    case 'ai':
      return typeof d.score === 'number' ? `Skor indikatif ${d.score}.` : null;
    case 'link':
      return typeof d.score === 'number' ? `Skor tautan ${d.score}.` : null;
    case 'ela':
      return typeof d.reliability === 'number'
        ? `Keandalan ${Math.round(d.reliability * 100)}%.`
        : null;
    case 'fact': {
      if (typeof d.status !== 'string') return null;
      const map: Record<string, string> = {
        CONTRADICTS: 'Klaim dibantah sumber terverifikasi.',
        SUPPORTS: 'Klaim didukung sumber terverifikasi.',
        UNVERIFIED: 'Belum ada sumber terverifikasi.'
      };
      return map[d.status] ?? null;
    }
  }
}
