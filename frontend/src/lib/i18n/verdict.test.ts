import { describe, it, expect, beforeEach } from 'vitest';
import { simpleSentence, aiLines, explanation } from './verdict';
import { locale } from './index.svelte';
import type { DossierParsed } from '$lib/api/schemas';

/** Just enough of a dossier for the lead sentence: floors, status, video. */
function dossier(floors: string[], video: boolean): DossierParsed {
  return {
    verdict: {
      score: 52,
      status: 'SUSPICIOUS',
      short_title: 'contoh',
      explanation: 'Penjelasan contoh.',
      explanation_simple: 'Kalimat dari server.',
      basis: ['ai', 'link'],
      not_checked: ['ela', 'fact'],
      floors_applied: floors
    },
    video: video ? ({ status: 'OK' } as DossierParsed['video']) : null
  } as DossierParsed;
}

describe('lead sentence for video floors', () => {
  beforeEach(() => locale.set('en'));

  it('says the caption does not match the footage for VIDEO_CONTEXT_MISMATCH', () => {
    expect(simpleSentence(dossier(['VIDEO_CONTEXT_MISMATCH'], true))).toBe(
      'The caption does not match what the footage shows. Check before you share.'
    );
  });

  it('says the video looks edited to mislead for VIDEO_MANIPULATED, ahead of a context mismatch', () => {
    expect(simpleSentence(dossier(['VIDEO_CONTEXT_MISMATCH', 'VIDEO_MANIPULATED'], true))).toBe(
      'This video looks edited to mislead. Check before you share.'
    );
  });

  it('speaks of the video, not an image, when AI_LIKELY comes from a video', () => {
    expect(simpleSentence(dossier(['AI_LIKELY', 'VIDEO_MANIPULATED'], true))).toBe(
      'This video looks AI-generated. Check before you share.'
    );
    expect(simpleSentence(dossier(['AI_LIKELY'], false))).toBe('This image looks AI-generated. Check before sharing.');
  });

  it('keeps the stronger floors first and the server wording in Indonesian', () => {
    expect(simpleSentence(dossier(['VIDEO_CONTEXT_MISMATCH', 'LINK_STRONG'], true))).toBe(
      'This link shows scam signs. Better not click it.'
    );
    locale.set('id');
    expect(simpleSentence(dossier(['VIDEO_CONTEXT_MISMATCH'], true))).toBe('Kalimat dari server.');
  });
});

describe('Jejak AI card for a video link', () => {
  beforeEach(() => locale.set('id'));

  const ai: DossierParsed['pillars']['ai'] = {
    available: true,
    score: 82,
    indicators: [],
    caveat: 'Indikatif, bukan bukti pasti',
    provenance: null,
    detector: null
  };

  it('names the video frames as what was checked, not the preview image', () => {
    expect(aiLines(ai, 'URL', true)).toEqual([
      'Yang diperiksa: beberapa frame dari video tautan ini.',
      'Kemungkinan dibuat dengan AI: tinggi.'
    ]);
    expect(aiLines(ai, 'URL')[0]).toBe('Yang diperiksa: gambar pratinjau tautan ini.');
  });
});

describe('explanation for a link whose preview is blocked', () => {
  beforeEach(() => locale.set('en'));
  const blocked = (video: boolean) =>
    ({ ...dossier([], video), source_metadata: { status: 'BLOCKED' } }) as unknown as DossierParsed;

  it('mentions the blocked preview only when the video was not analysed either', () => {
    expect(explanation(blocked(false))).toContain('preview is closed');
    expect(explanation(blocked(true))).not.toContain('preview is closed');
  });
});

describe('INSUFFICIENT and the not-checked sentence', () => {
  beforeEach(() => locale.set('id'));

  it('shows an old INSUFFICIENT payload as a neutral conclusion, not as missing data', () => {
    const d = dossier([], false);
    d.verdict.status = 'INSUFFICIENT';
    d.verdict.score = null;
    const sentence = simpleSentence(d);
    expect(sentence).toBe('Tidak ada tanda bahaya yang ditemukan, tetapi datanya terbatas.');
    expect(sentence.toLowerCase()).not.toContain('belum cukup');
  });

  it('drops the server sentence that lists what could not be checked', () => {
    const d = dossier([], false);
    d.verdict.explanation =
      'Diperiksa: jejak AI, keamanan tautan. Tidak dapat diperiksa: teks ditimpa (ELA), cek fakta. Hasil ini indikatif, bukan bukti pasti.';
    expect(explanation(d)).toBe('Diperiksa: jejak AI, keamanan tautan. Hasil ini indikatif, bukan bukti pasti.');
  });
});
