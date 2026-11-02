import { describe, expect, it } from 'vitest';
import { buildMockDossier, decodeMockToken, isMockToken, mockAuditStream } from './mock';

describe('buildMockDossier', () => {
  it('marks every preview result as a demo and signs a decodable preview token', () => {
    const dossier = buildMockDossier('text', {
      text: 'Bantuan Rp 500.000 dibagikan hari ini saja. Sebarkan ke 10 grup!'
    });

    expect(dossier.is_demo).toBe(true);
    expect(dossier.verdict.status).toBe('TAMPERED');
    expect(dossier.pillars.fact.status).toBe('CONTRADICTS');
    expect(dossier.tactics.length).toBeGreaterThan(0);
    expect(isMockToken(dossier.share_token)).toBe(true);
    expect(decodeMockToken(dossier.share_token)?.preview).toBe(true);
  });

  it('recognises a benign announcement without inventing a threat', () => {
    const dossier = buildMockDossier('text', {
      text: 'Jadwal posyandu bulan ini: Sabtu, 8 pagi di balai warga.'
    });

    expect(dossier.verdict.status).toBe('SAFE');
    expect(dossier.verdict.score).toBe(8);
    expect(dossier.pillars.fact.status).toBe('SUPPORTS');
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
      'text',
      { text: 'Akun Anda akan diblokir. Verifikasi sekarang.' },
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
      'text',
      { text: 'apa saja' },
      controller.signal,
      { step: 0, settle: 0 }
    )) {
      events.push(event);
    }

    expect(events).toEqual([]);
  });
});
