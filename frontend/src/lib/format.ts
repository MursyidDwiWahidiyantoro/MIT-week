import type { Status } from './api/types';
import { t } from './i18n/index.svelte';

/**
 * Status labels. The Indonesian wording comes from TECH_SPEC §3.3. They never
 * claim certainty: SAFE is "no danger signs found", not "safe".
 */
export function statusLabel(status: Status): string {
  return t(`status.${status}`);
}

/**
 * The score is a 0-100 RISK score, not a probability or confidence.
 * Never render it with a percent sign.
 */
export function formatScore(score: number | null): string {
  return score === null ? t('score.none') : t('score.of100', { score });
}
