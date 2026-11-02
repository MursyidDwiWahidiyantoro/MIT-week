import { describe, expect, it } from 'vitest';
import {
  firstHttpUrl,
  isAllowedImageType,
  SHARE_MAX_BYTES,
  validateShareFile
} from './share-target';

describe('firstHttpUrl', () => {
  it('extracts a URL from shared text', () => {
    expect(firstHttpUrl('Cek ini https://example.org/artikel dulu ya')).toBe(
      'https://example.org/artikel'
    );
  });

  it('trims trailing punctuation', () => {
    expect(firstHttpUrl('Lihat https://example.org/artikel.')).toBe('https://example.org/artikel');
  });

  it('returns null when there is no link', () => {
    expect(firstHttpUrl('Tidak ada tautan di sini')).toBeNull();
  });
});

describe('shared image validation', () => {
  it('accepts only the documented image types', () => {
    expect(isAllowedImageType('image/png')).toBe(true);
    expect(isAllowedImageType('image/jpeg')).toBe(true);
    expect(isAllowedImageType('image/webp')).toBe(true);
    expect(isAllowedImageType('image/gif')).toBe(false);
    expect(isAllowedImageType('application/pdf')).toBe(false);
  });

  it('rejects files over the 10 MB limit', () => {
    const file = new File(['x'], 'besar.png', { type: 'image/png' });
    Object.defineProperty(file, 'size', { value: SHARE_MAX_BYTES + 1 });
    expect(validateShareFile(file)).toContain('10 MB');
  });

  it('accepts a normal PNG', () => {
    const file = new File(['x'], 'gambar.png', { type: 'image/png' });
    expect(validateShareFile(file)).toBeNull();
  });
});
