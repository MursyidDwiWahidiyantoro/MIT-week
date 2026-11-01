import { describe, expect, it } from 'vitest';
import { buildShareUrl, tokenFromHash } from './verify';

describe('tokenFromHash', () => {
  it('reads the token from a fragment', () => {
    expect(tokenFromHash('#t=v1.abc.def')).toBe('v1.abc.def');
  });

  it('accepts a hash without the leading marker', () => {
    expect(tokenFromHash('t=v1.abc.def')).toBe('v1.abc.def');
  });

  it('returns null when there is no token', () => {
    expect(tokenFromHash('')).toBeNull();
    expect(tokenFromHash('#')).toBeNull();
    expect(tokenFromHash('#other=1')).toBeNull();
  });

  it('does not treat a query string as a token source', () => {
    expect(tokenFromHash('?t=v1.abc.def')).toBeNull();
  });
});

describe('buildShareUrl', () => {
  it('puts the token in the fragment, never the query', () => {
    const url = buildShareUrl('https://amba.example', 'v1.abc.def');
    expect(url).toBe('https://amba.example/hasil#t=v1.abc.def');
    expect(url).not.toContain('?');
  });

  it('trims a trailing slash from the origin', () => {
    expect(buildShareUrl('https://amba.example/', 'x')).toBe('https://amba.example/hasil#t=x');
  });

  it('encodes characters that would break the fragment', () => {
    const url = buildShareUrl('https://a.example', 'a+b/c=');
    expect(url).toContain('a%2Bb%2Fc%3D');
  });
});
