import { describe, it, expect, beforeEach } from 'vitest';
import { id } from './id';
import { en } from './en';
import { t, locale } from './index.svelte';

describe('i18n', () => {
  beforeEach(() => locale.set('id'));

  it('every key has non-empty copy in both languages', () => {
    for (const key of Object.keys(id) as Array<keyof typeof id>) {
      // The note introducing the Indonesian legal text only exists in English.
      if (key !== 'about.disclaimerNoteEn') expect(id[key], key).not.toBe('');
      expect(en[key], key).not.toBe('');
    }
    expect(Object.keys(en).sort()).toEqual(Object.keys(id).sort());
  });

  it('switches language and fills placeholders', () => {
    expect(t('score.of100', { score: 85 })).toBe('Skor risiko 85 dari 100');
    locale.set('en');
    expect(t('score.of100', { score: 85 })).toBe('Risk score 85 out of 100');
    expect(document.documentElement.lang).toBe('en');
    expect(localStorage.getItem('amba.lang')).toBe('en');
  });

  it('keeps placeholders consistent between languages', () => {
    const vars = (s: string) => (s.match(/\{\w+\}/g) ?? []).sort().join(',');
    for (const key of Object.keys(id) as Array<keyof typeof id>) {
      expect(vars(en[key]), key).toBe(vars(id[key]));
    }
  });

  it('never claims certainty in English copy either', () => {
    for (const [key, text] of Object.entries(en)) {
      if (key === 'disclaimer.full' || key.startsWith('edu.q1.') || key === 'about.p1') continue;
      expect(/\b(100%|guaranteed|proven)\b/i.test(text), `${key}: ${text}`).toBe(false);
    }
  });
});
