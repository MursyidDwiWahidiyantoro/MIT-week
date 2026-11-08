/**
 * Interface language: Indonesian (default) or English (PRD P1 EN/ID toggle).
 *
 * The Indonesian dictionary holds the original copy verbatim, so the default
 * experience is unchanged. English covers the whole interface; content that
 * comes from the user or the model (a claim, OCR text, a caption) stays in
 * its own language, and Balas Sopan replies stay Indonesian because they are
 * written for Indonesian family chats.
 */
import { id, type MessageKey } from './id';
import { en } from './en';

export type Locale = 'id' | 'en';
const STORAGE_KEY = 'amba.lang';
const DICTS: Record<Locale, Record<MessageKey, string>> = { id, en };

function initial(): Locale {
  if (typeof document !== 'undefined' && document.documentElement.lang === 'en') return 'en';
  try {
    return localStorage.getItem(STORAGE_KEY) === 'en' ? 'en' : 'id';
  } catch {
    return 'id';
  }
}

let current = $state<Locale>(typeof window === 'undefined' ? 'id' : initial());

export const locale = {
  get value(): Locale {
    return current;
  },
  set(next: Locale): void {
    current = next;
    if (typeof document !== 'undefined') document.documentElement.lang = next;
    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch {
      // Session-only when storage is blocked.
    }
  }
};

/**
 * Translates a key. `{name}` placeholders are filled from params. Reading
 * `current` here makes every call site reactive to a language switch.
 */
export function t(key: MessageKey, params?: Record<string, string | number>): string {
  let text = DICTS[current][key] ?? id[key] ?? key;
  if (params) {
    for (const [name, value] of Object.entries(params)) {
      text = text.replaceAll(`{${name}}`, String(value));
    }
  }
  return text;
}

/** BCP 47 tag for Intl formatting and speech synthesis. */
export function bcp47(): string {
  return current === 'en' ? 'en-US' : 'id-ID';
}

export type { MessageKey };
