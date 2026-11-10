import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import EvidenceAccordion from './EvidenceAccordion.svelte';
import { locale } from '$lib/i18n/index.svelte';
import type { DossierParsed } from '$lib/api/schemas';

type Pillars = DossierParsed['pillars'];

const src = (n: number, url = `https://turnbackhoax.id/a${n}`) => ({
  title: `[SALAH] Klaim ${n}`,
  publisher: 'turnbackhoax.id',
  url,
  match: 'CONTRADICTS' as const
});

function pillars(sources: ReturnType<typeof src>[], factAvailable = true): Pillars {
  return {
    ela: { available: false },
    ai: { available: false },
    link: { available: false },
    fact: { available: factAvailable, status: 'CONTRADICTS', claim: 'Klaim contoh', sources }
  } as unknown as Pillars;
}

beforeEach(() => locale.set('id'));

describe('EvidenceAccordion', () => {
  it('shows only the checks that ran and never says one could not be checked', () => {
    const { container } = render(EvidenceAccordion, { pillars: pillars([src(1)]) });
    expect(container.querySelectorAll('details')).toHaveLength(1);
    expect(container.textContent).not.toMatch(/tidak dapat diperiksa/i);
  });

  it('lists at most five fact-check sources and every one is a link', () => {
    const { container } = render(EvidenceAccordion, {
      pillars: pillars(Array.from({ length: 9 }, (_, i) => src(i + 1)))
    });
    const links = container.querySelectorAll('[data-testid="sources"] a');
    expect(links).toHaveLength(5);
    for (const a of links) expect(a.getAttribute('href')).toMatch(/^https:\/\//);
  });

  it('drops a source whose address is not http(s)', () => {
    const { container } = render(EvidenceAccordion, {
      pillars: pillars([src(1, 'javascript:alert(1)'), src(2)])
    });
    expect(container.querySelectorAll('[data-testid="sources"] a')).toHaveLength(1);
  });

  it('says plainly that no fact-check article covers the claim when the server sent none', () => {
    render(EvidenceAccordion, { pillars: pillars([]) });
    expect(screen.getByText('Belum ada artikel cek fakta yang membahas klaim ini.')).toBeTruthy();
  });
});
