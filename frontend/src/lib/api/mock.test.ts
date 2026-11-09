import { describe, expect, it } from 'vitest';
import { buildMockDossier, decodeMockToken, isMockToken, mockAuditStream } from './mock';

describe('buildMockDossier', () => {
  it('marks every preview result as a demo and signs a decodable preview token', () => {
    const dossier = buildMockDossier('url', { url: 'https://bit.ly/verifikasi-akun' });

    expect(dossier.is_demo).toBe(true);
    expect(isMockToken(dossier.share_token)).toBe(true);
    expect(decodeMockToken(dossier.share_token)?.preview).toBe(true);
  });

  it('does not invent a threat for an official link', () => {
    const dossier = buildMockDossier('url', { url: 'https://www.kemkes.go.id/' });

    expect(dossier.verdict.status).not.toBe('TAMPERED');
    expect(dossier.pillars.link.reasons.map((r) => r.code)).not.toContain('LOOKALIKE');
  });

  it('explains suspicious URL signals', () => {
    const dossier = buildMockDossier('url', {
      url: 'https://bit.ly/verifikasi-akun'
    });

    expect(dossier.pillars.link.available).toBe(true);
    expect(dossier.pillars.link.reasons.map((r) => r.code)).toContain('SHORTENER');
    expect(dossier.verdict.status).not.toBe('SAFE');
    expect(dossier.pillars.ela.available).toBe(false);
  });

  it('simulates a local ELA preview for an uploaded image', () => {
    const file = new File(['not-a-real-image'], 'tangkapan.png', { type: 'image/png' });
    const dossier = buildMockDossier('file', { file });

    expect(dossier.input_type).toBe('FILE');
    expect(dossier.pillars.ela.available).toBe(true);
    expect(dossier.pillars.ela.anomalies).toHaveLength(1);
    expect(dossier.pillars.ela.reliability).toBeLessThan(1);
  });
});

describe('mockAuditStream', () => {
  it('emits the same stage → partial → verdict → done order as the API', async () => {
    const controller = new AbortController();
    const events = [];
    for await (const event of mockAuditStream(
      'url',
      { url: 'https://bri-verifikasi.xyz/login' },
      controller.signal,
      { step: 0, settle: 0 }
    )) {
      events.push(event);
    }

    expect(events[0]).toMatchObject({ type: 'stage' });
    expect(events.some((event) => event.type === 'partial')).toBe(true);
    expect(events.some((event) => event.type === 'verdict')).toBe(true);
    expect(events.at(-1)).toEqual({ type: 'done' });
  });

  it('stops without yielding a verdict after cancellation', async () => {
    const controller = new AbortController();
    controller.abort();
    const events = [];
    for await (const event of mockAuditStream(
      'url',
      { url: 'https://example.org/' },
      controller.signal,
      { step: 0, settle: 0 }
    )) {
      events.push(event);
    }

    expect(events).toEqual([]);
  });
});
