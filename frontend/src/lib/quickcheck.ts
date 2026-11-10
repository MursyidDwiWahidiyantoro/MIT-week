import data from './data/quickcheck.json';

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

export const quickCheckQuestions = data.questions as QuickCheckQuestion[];
