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
  for (const floor of [
    'FACT_CONTRADICTED',
    'LINK_MALICIOUS',
    'LINK_STRONG',
    'AI_PROVENANCE',
    'ELA_CORROBORATED',
    'AI_EDITED',
    'AI_LIKELY',
    'VIDEO_MANIPULATED',
    'VIDEO_CONTEXT_MISMATCH'
  ]) {
    if (!floors.includes(floor)) continue;
    // A social video link reaches AI_LIKELY through its frames, not an image.
    if (floor === 'AI_LIKELY' && d.video) return t('simple.AI_LIKELY_VIDEO');
    return t(`simple.${floor}` as MessageKey);
  }
  return t(`simple.${d.verdict.status}` as MessageKey);
}

/** Mirrors the server's explanation(): honest limits in plain words. */
export function explanation(d: DossierParsed): string {
  if (locale.value === 'id') return d.verdict.explanation;
  const parts: string[] = [];
  const md = d.source_metadata;
  // A blocked preview matters only when the video was not analysed either.
  if (md?.status === 'BLOCKED' && d.video?.status !== 'OK') parts.push(t('expl.blocked'));
  if (md?.status === 'UNAVAILABLE') parts.push(t('expl.unavailable'));
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

type AIPillar = DossierParsed['pillars']['ai'];

/** The AI pillar's score bands, shared with the Telegram summary. */
export function aiLevel(score: number): 'high' | 'medium' | 'low' {
  return score >= 70 ? 'high' : score >= 40 ? 'medium' : 'low';
}

/**
 * The Jejak AI card in plain words: which image was checked, what the file
 * declares about its origin (or how likely it looks AI-made), and the
 * dedicated detector's answer when one ran. Built from codes, so both
 * languages say the same thing. `fromVideo`: the score came from the frames
 * of an analysed social video, not from the link's preview image.
 */
export function aiLines(ai: AIPillar, inputType: DossierParsed['input_type'], fromVideo = false): string[] {
  const lines: string[] = [];
  if (inputType === 'URL') lines.push(t(fromVideo ? 'video.aiChecked' : 'ai.preview'));
  const p = ai.provenance;
  if (p && (p.kind === 'AI_GENERATED' || p.kind === 'AI_EDITED')) {
    const made = p.kind === 'AI_EDITED' ? 'edited' : 'generated';
    const by = p.generator ? ` (${p.generator})` : '';
    const key = `ai.prov.${made}.${p.source}`;
    lines.push(has(key) ? t(key, { by }) : t(`ai.prov.${made}.xmp` as MessageKey, { by }));
  } else {
    lines.push(t(`ai.level.${aiLevel(ai.score)}` as MessageKey));
    if (p?.kind === 'CAMERA') {
      lines.push(p.camera ? t('ai.prov.camera', { camera: p.camera }) : t('ai.prov.cameraC2pa'));
    } else if (p?.c2pa) {
      lines.push(t('ai.prov.c2paNoAI'));
    } else if (p) {
      lines.push(t('ai.prov.none'));
    }
  }
  if (ai.detector) {
    const name = ai.detector.name.charAt(0).toUpperCase() + ai.detector.name.slice(1);
    lines.push(t('ai.detector', { name, n: ai.detector.score }));
  }
  return lines;
}
