import { bcp47 } from './i18n/index.svelte';

/**
 * Accessibility preferences that are independent of presentation mode:
 * text size and read-aloud. Both persist to localStorage and both degrade
 * safely when storage is unavailable.
 */

export type TextScale = 100 | 125 | 150;

const SCALE_KEY = 'amba.textScale';

function readScale(): TextScale {
  try {
    const raw = Number(localStorage.getItem(SCALE_KEY));
    return raw === 125 || raw === 150 ? raw : 100;
  } catch {
    return 100;
  }
}

let scale = $state<TextScale>(100);
let initialised = $state(false);

export const textSize = {
  get value(): TextScale {
    return scale;
  },
  init(): void {
    if (initialised) return;
    scale = readScale();
    apply(scale);
    initialised = true;
  },
  set(next: TextScale): void {
    scale = next;
    apply(next);
    try {
      localStorage.setItem(SCALE_KEY, String(next));
    } catch {
      // Session-only when storage is blocked.
    }
  }
};

function apply(value: TextScale): void {
  if (typeof document === 'undefined') return;
  // Root font size scales every rem-based token at once. Set as a factor so
  // Essential Mode's --mode-scale still applies (see tokens.css).
  document.documentElement.style.setProperty('--text-scale', String(value / 100));
}

/** Read-aloud uses the browser's own speech synthesis, Indonesian voice. */
export function speechAvailable(): boolean {
  return typeof window !== 'undefined' && 'speechSynthesis' in window;
}

export function speak(text: string): void {
  if (!speechAvailable()) return;
  window.speechSynthesis.cancel();
  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = bcp47();
  utterance.rate = 0.95;
  window.speechSynthesis.speak(utterance);
}

export function stopSpeaking(): void {
  if (!speechAvailable()) return;
  window.speechSynthesis.cancel();
}

/**
 * Auto-read: when on, a new result speaks its lead sentence and explanation
 * as soon as it appears (PRD F14). Independent of presentation mode and off
 * by default: unexpected audio is not acceptable (DESIGN.md).
 */
const AUTO_READ_KEY = 'amba.autoRead';
let autoReadOn = $state(false);
let autoReadLoaded = false;

export const autoRead = {
  get value(): boolean {
    if (!autoReadLoaded) {
      autoReadLoaded = true;
      try {
        autoReadOn = localStorage.getItem(AUTO_READ_KEY) === '1';
      } catch {
        autoReadOn = false;
      }
    }
    return autoReadOn;
  },
  set(next: boolean): void {
    autoReadOn = next;
    autoReadLoaded = true;
    try {
      localStorage.setItem(AUTO_READ_KEY, next ? '1' : '0');
    } catch {
      // Session-only when storage is blocked.
    }
  }
};
