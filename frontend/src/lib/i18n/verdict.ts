/**
 * Server-written copy, localised. The API speaks Indonesian (TECH_SPEC §2);
 * every sentence it writes is derived from codes in the Dossier, so the
 * English interface rebuilds the same sentence from the same codes. In
 * Indonesian the server's own text is shown unchanged (it can be richer,
 * e.g. the demo cases' explanations).
 */
import type { DossierParsed } from '$lib/api/schemas';
import { t, locale, type MessageKey } from './index.svelte';

const has = (key: string): key is MessageKey => t(key as MessageKey) !== key;

export function shortTitle(d: DossierParsed): string {
  return locale.value === 'id' ? d.verdict.short_title : t(`status.${d.verdict.status}`);
}

/** TECH_SPEC §8.2 deterministic lead sentence (scoring.ExplanationSimple). */
export function simpleSentence(d: DossierParsed): string {
  if (locale.value === 'id') return d.verdict.explanation_simple;
  const floors = d.verdict.floors_applied;
  for (const floor of ['FACT_CONTRADICTED', 'LINK_MALICIOUS', 'SCAM_PATTERN', 'LINK_STRONG', 'ELA_CORROBORATED']) {
    if (floors.includes(floor)) return t(`simple.${floor}` as MessageKey);
  }
  return t(`simple.${d.verdict.status}` as MessageKey);
}

/** Mirrors the server's explanation(): honest limits in plain words. */
export function explanation(d: DossierParsed): string {
  if (locale.value === 'id') return d.verdict.explanation;
  const parts: string[] = [];
  const md = d.source_metadata;
  if (md?.status === 'BLOCKED') parts.push(t('expl.blocked'));
  if (md?.status === 'UNAVAILABLE') parts.push(t('expl.unavailable'));
  const scam = d.verdict.scam_pattern;
  if (scam && has(`scam.${scam}`)) parts.push(t(`scam.${scam}` as MessageKey));
  const basis = d.verdict.basis.filter((b) => b !== 'llm_untrusted');
  if (basis.length) parts.push(t('expl.checked', { list: basis.map((b) => t(`expl.pillar.${b}` as MessageKey)).join(', ') }));
  if (d.verdict.not_checked.length) {
    parts.push(t('expl.notChecked', { list: d.verdict.not_checked.map((b) => t(`expl.pillar.${b}` as MessageKey)).join(', ') }));
  }
  parts.push(t('expl.indicative'));
  return parts.join(' ');
}

export function linkReason(code: string, serverText: string): string {
  if (locale.value === 'id') return serverText;
  const key = `reason.${code}`;
  return has(key) ? t(key) : t('reason.default');
}

export function caveat(serverText: string): string {
  return locale.value === 'id' ? serverText : t('ai.caveat');
}

export function reliabilityNote(serverText: string | null | undefined): string {
  if (!serverText) return '';
  return locale.value === 'id' ? serverText : t('ela.reliabilityNote');
}
