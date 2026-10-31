/**
 * Presentation mode: Jelajah (explore) or Esensial (essential).
 *
 * Modes are chosen by preference only — never labelled by age, ability or
 * skill (CLAUDE.md rule 12). Both modes expose identical information and
 * functionality; they differ in decoration density, motion and type size.
 */
export type Mode = 'explore' | 'essential';

const STORAGE_KEY = 'amba.mode';

function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined' || !window.matchMedia) return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

function readStored(): Mode | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw === 'explore' || raw === 'essential' ? raw : null;
  } catch {
    // Private mode or blocked storage: fall through to the default.
    return null;
  }
}

function applyToDocument(mode: Mode | null): void {
  if (typeof document === 'undefined') return;
  const root = document.documentElement;
  if (mode) {
    root.dataset.mode = mode;
  } else {
    delete root.dataset.mode;
  }
  // Essential Mode always runs calm; reduced-motion forces calm in both modes.
  const calm = mode === 'essential' || prefersReducedMotion();
  root.dataset.motion = calm ? 'calm' : 'full';
}

/** Null means the visitor has not chosen yet, so the neutral picker shows. */
let current = $state<Mode | null>(null);
let initialised = $state(false);
let dismissed = $state(false);

export const mode = {
  get value(): Mode | null {
    return current;
  },
  get resolved(): Mode {
    return current ?? (prefersReducedMotion() ? 'essential' : 'explore');
  },
  /**
   * The picker is a first-visit affordance only. Once the visitor has chosen
   * (or a stored choice was found) it never shows again; without a stored
   * choice the resolved mode already applies, so re-prompting on every route
   * would nag rather than help.
   */
  get needsPicker(): boolean {
    return initialised && current === null && !dismissed;
  },
  init(): void {
    if (initialised) return;
    current = readStored();
    applyToDocument(current);
    initialised = true;
  },
  dismiss(): void {
    dismissed = true;
  },
  set(next: Mode): void {
    current = next;
    dismissed = true;
    applyToDocument(next);
    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch {
      // Storage unavailable: the choice still applies for this session.
    }
  }
};

/** Every information item carries one of these keys in both modes. */
export const INFO_KEYS = [
  'verdict.score',
  'verdict.sentence',
  'verdict.explanation',
  'verdict.limits',
  'pillar.ela',
  'pillar.ai',
  'pillar.fact',
  'pillar.link',
  'sources',
  'tactics',
  'reply',
  'share',
  'proofcard',
  'readaloud'
] as const;

export type InfoKey = (typeof INFO_KEYS)[number];
