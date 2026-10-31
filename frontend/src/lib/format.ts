import type { Status } from './api/types';

/**
 * Indonesian status labels. Exact wording comes from TECH_SPEC §3.3.
 * These never claim certainty: SAFE is "no danger signs found", not "safe".
 */
const LABELS: Record<Status, string> = {
  SAFE: 'Tidak ditemukan tanda bahaya',
  SUSPICIOUS: 'Perlu diperiksa',
  TAMPERED: 'Terindikasi Hoax/Manipulasi',
  INSUFFICIENT: 'Belum cukup data'
};

export function statusLabel(status: Status): string {
  return LABELS[status];
}

/**
 * The score is a 0-100 RISK score, not a probability or confidence.
 * Never render it with a percent sign.
 */
export function formatScore(score: number | null): string {
  return score === null ? 'Belum ada skor' : `Skor risiko ${score} dari 100`;
}
