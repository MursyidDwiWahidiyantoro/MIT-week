import { describe, expect, it } from 'vitest';
import { formatScore, statusLabel } from './format';

describe('statusLabel', () => {
  it('never returns a certainty claim', () => {
    const labels = (['SAFE', 'SUSPICIOUS', 'TAMPERED', 'INSUFFICIENT'] as const).map(statusLabel);
    for (const label of labels) {
      expect(label.toLowerCase()).not.toMatch(/pasti|terbukti|100%|dijamin|aman/);
    }
  });

  it('maps SAFE to the spec sentence', () => {
    expect(statusLabel('SAFE')).toBe('Tidak ditemukan tanda bahaya');
  });

  it('maps TAMPERED to the spec label', () => {
    expect(statusLabel('TAMPERED')).toBe('Terindikasi Hoax/Manipulasi');
  });

  it('maps INSUFFICIENT honestly', () => {
    expect(statusLabel('INSUFFICIENT')).toBe('Belum cukup data');
  });
});

describe('formatScore', () => {
  it('describes the score as risk, never confidence', () => {
    expect(formatScore(92)).toBe('Skor risiko 92 dari 100');
  });

  it('renders an em dash when no score exists', () => {
    expect(formatScore(null)).toBe('Belum ada skor');
  });
});
