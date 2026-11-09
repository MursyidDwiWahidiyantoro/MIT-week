import type { DossierParsed } from './api/schemas';

/**
 * The most recent verdict, held in memory for the /hasil page.
 *
 * Nothing is persisted (CLAUDE.md rule 1): a reload shows the "hasil tidak
 * tersimpan" state instead of a stale verdict. The uploaded image is kept only
 * as an object URL so the ELA viewer can render it, and is revoked the moment
 * a new audit starts or the result is cleared.
 */
export interface ResultEntry {
  dossier: DossierParsed;
  imageFile: File | null;
  imageUrl: string | null;
}

let current = $state<ResultEntry | null>(null);

function revoke(): void {
  const url = current?.imageUrl;
  if (!url) return;
  try {
    URL.revokeObjectURL(url);
  } catch {
    // Nothing to revoke in a worker or a torn-down document.
  }
}

export const result = {
  get value(): DossierParsed | null {
    return current?.dossier ?? null;
  },
  get imageUrl(): string | null {
    return current?.imageUrl ?? null;
  },
  get imageFile(): File | null {
    return current?.imageFile ?? null;
  },
  set(dossier: DossierParsed, imageFile: File | null = null): void {
    revoke();
    let imageUrl: string | null = null;
    if (imageFile) {
      try {
        imageUrl = URL.createObjectURL(imageFile);
      } catch {
        imageUrl = null;
      }
    }
    current = { dossier, imageFile, imageUrl };
  },
  clear(): void {
    revoke();
    current = null;
  }
};
