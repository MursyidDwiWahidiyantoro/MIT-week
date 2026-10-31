import type { DossierParsed } from './api/schemas';

/**
 * The most recent verdict, held in memory for the /hasil page.
 *
 * Nothing is persisted (CLAUDE.md rule 1): a reload shows the "hasil tidak
 * tersimpan" state instead of a stale verdict.
 */
let current = $state<DossierParsed | null>(null);

export const result = {
  get value(): DossierParsed | null {
    return current;
  },
  set(dossier: DossierParsed | null): void {
    current = dossier;
  },
  clear(): void {
    current = null;
  }
};
