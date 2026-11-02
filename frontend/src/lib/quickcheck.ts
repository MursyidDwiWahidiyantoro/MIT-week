import data from './data/quickcheck.json';
import words from './data/family-words.json';

export type QuickCheckLevel = 'high' | 'wary' | 'low';

export interface QuickCheckQuestion {
  id: string;
  text: string;
  weight: number;
}

export interface QuickCheckResult {
  level: QuickCheckLevel;
  label: string;
  summary: string;
  steps: string[];
  score: number;
  forcedByOtp: boolean;
}

interface LevelData {
  min: number;
  label: string;
  summary: string;
  steps: string[];
}

const LEVELS = data.levels as Record<QuickCheckLevel, LevelData>;
const OTP_ID = data.otpQuestionId;

/**
 * Scores the six-question flow. A "yes" to OTP/PIN is always Tinggi, even if
 * the numeric total would land in a lower band (TECH_SPEC §8.5).
 */
export function evaluateQuickCheck(answers: Record<string, boolean>): QuickCheckResult {
  const questions = data.questions as QuickCheckQuestion[];
  let score = 0;
  for (const question of questions) {
    if (answers[question.id]) score += question.weight;
  }

  const forcedByOtp = answers[OTP_ID] === true;
  const level: QuickCheckLevel = forcedByOtp
    ? 'high'
    : score >= LEVELS.high.min
      ? 'high'
      : score >= LEVELS.wary.min
        ? 'wary'
        : 'low';

  return {
    level,
    score,
    forcedByOtp,
    label: LEVELS[level].label,
    summary: LEVELS[level].summary,
    steps: LEVELS[level].steps
  };
}

/** Two distinct words, joined for easy reading and printing. */
export function generateFamilyCode(
  random: () => number = secureRandom
): string {
  const list = words as string[];
  if (list.length < 2) throw new Error('Daftar kata tidak cukup');
  const first = pick(list.length, random);
  let second = pick(list.length, random);
  // A repeated word is valid but less memorable; retry a bounded number of
  // times so a mocked or unlucky generator still terminates.
  for (let attempt = 0; second === first && attempt < 8; attempt++) {
    second = pick(list.length, random);
  }
  if (second === first) second = (first + 1) % list.length;
  return `${list[first]}-${list[second]}`;
}

function pick(length: number, random: () => number): number {
  // Rejection sampling avoids modulo bias when the range is not a power of 2.
  const limit = Math.floor(0x100000000 / length) * length;
  for (let attempt = 0; attempt < 32; attempt++) {
    const value = Math.floor(random() * 0x100000000);
    if (value < limit) return value % length;
  }
  return Math.floor(random() * length) % length;
}

function secureRandom(): number {
  const buffer = new Uint32Array(1);
  crypto.getRandomValues(buffer);
  return buffer[0] / 0x100000000;
}

export const quickCheckQuestions = data.questions as QuickCheckQuestion[];
export const familyWordCount = (words as string[]).length;
