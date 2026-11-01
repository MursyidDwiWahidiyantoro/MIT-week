import { describe, expect, it, vi, beforeEach } from 'vitest';
import { speechAvailable, speak, stopSpeaking } from './a11y.svelte';

describe('speech helpers', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('reports availability from the environment', () => {
    expect(typeof speechAvailable()).toBe('boolean');
  });

  it('does nothing when synthesis is unavailable', () => {
    const original = window.speechSynthesis;
    // @ts-expect-error simulate a browser without speech synthesis
    delete window.speechSynthesis;
    expect(() => speak('halo')).not.toThrow();
    expect(() => stopSpeaking()).not.toThrow();
    if (original) window.speechSynthesis = original;
  });
});
