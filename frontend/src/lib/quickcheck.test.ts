import { describe, expect, it } from 'vitest';
import {
  evaluateQuickCheck,
  familyWordCount,
  generateFamilyCode,
  quickCheckQuestions
} from './quickcheck';
import words from './data/family-words.json';

describe('evaluateQuickCheck', () => {
  it('returns Rendah when no warning sign is present', () => {
    const result = evaluateQuickCheck({});
    expect(result.level).toBe('low');
    expect(result.score).toBe(0);
    expect(result.steps).toHaveLength(3);
  });

  it('returns Waspada in the 3–5 band', () => {
    const result = evaluateQuickCheck({ money: true, urgent: true });
    expect(result.score).toBe(5);
    expect(result.level).toBe('wary');
  });

  it('returns Tinggi at 6 or more', () => {
    const result = evaluateQuickCheck({ money: true, urgent: true, secret: true });
    expect(result.score).toBe(7);
    expect(result.level).toBe('high');
  });

  it('forces Tinggi for any OTP/PIN request, even alone', () => {
    const result = evaluateQuickCheck({ otp: true });
    expect(result.score).toBe(4);
    expect(result.level).toBe('high');
    expect(result.forcedByOtp).toBe(true);
  });

  it('uses the six questions and their documented weights', () => {
    expect(quickCheckQuestions).toHaveLength(6);
    const total = quickCheckQuestions.reduce((sum, q) => sum + q.weight, 0);
    expect(total).toBe(14);
  });
});

describe('Sandi Keluarga', () => {
  it('has at least 256 unique common words', () => {
    expect(familyWordCount).toBeGreaterThanOrEqual(256);
    expect(new Set(words).size).toBe(words.length);
  });

  it('generates two distinct words', () => {
    const code = generateFamilyCode(() => 0);
    const parts = code.split('-');
    expect(parts).toHaveLength(2);
    expect(parts[0]).not.toBe(parts[1]);
  });

  it('uses the injected random source instead of Math.random', () => {
    const calls: number[] = [];
    const code = generateFamilyCode(() => {
      calls.push(1);
      return 0.5;
    });
    expect(calls.length).toBeGreaterThan(0);
    expect(code).toMatch(/^[a-z]+-[a-z]+$/);
  });
});
