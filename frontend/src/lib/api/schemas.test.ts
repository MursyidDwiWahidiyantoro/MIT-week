import { describe, expect, it } from 'vitest';
import { parseDossier, SchemaVersionError } from './schemas';

function minimalDossier(overrides: Record<string, unknown> = {}) {
  return {
    schema_version: 2,
    id: 'AMBA-20261030-7F3K',
    timestamp: '2026-10-30T07:22:15Z',
    input_type: 'URL',
    source_reference: 'contoh',
    degraded: false,
    is_demo: true,
    verdict: {
      score: 78,
      status: 'TAMPERED',
      short_title: 'contoh',
      explanation: 'Penjelasan contoh.',
      explanation_simple: 'Kabar ini dibantah sumber terpercaya. Jangan disebarkan.',
      basis: ['ai', 'fact'],
      not_checked: ['ela', 'link'],
      floors_applied: ['FACT_CONTRADICTED']
    },
    pillars: {
      ela: { available: false, score: 0, reliability: 0, reliability_note: null, anomalies: [] },
      ai: { available: true, score: 82, indicators: [], caveat: 'Indikatif, bukan bukti pasti' },
      fact: { available: true, claim: 'x', status: 'CONTRADICTS', sources: [] },
      link: { available: false, score: 0, final_url: '', redirect_chain: [], reasons: [] }
    },
    tactics: [],
    share_token: '',
    ...overrides
  };
}

describe('parseDossier', () => {
  it('accepts a valid dossier', () => {
    expect(parseDossier(minimalDossier()).verdict.status).toBe('TAMPERED');
  });

  it('rejects a score above 100', () => {
    const bad = minimalDossier();
    bad.verdict.score = 120;
    expect(() => parseDossier(bad)).toThrow();
  });

  it('rejects an unknown status', () => {
    const bad = minimalDossier();
    bad.verdict.status = 'PROBABLY_FINE';
    expect(() => parseDossier(bad)).toThrow();
  });

  it('allows a null score for INSUFFICIENT', () => {
    const base = minimalDossier();
    const d: Record<string, unknown> = {
      ...base,
      verdict: { ...base.verdict, score: null, status: 'INSUFFICIENT' }
    };
    expect(parseDossier(d).verdict.score).toBeNull();
  });

  it('raises SchemaVersionError for a newer contract', () => {
    expect(() => parseDossier(minimalDossier({ schema_version: 3 }))).toThrow(SchemaVersionError);
  });

  it('raises SchemaVersionError with the update message', () => {
    try {
      parseDossier(minimalDossier({ schema_version: 99 }));
    } catch (err) {
      expect((err as Error).message).toBe('Perbarui aplikasi');
    }
  });
});
