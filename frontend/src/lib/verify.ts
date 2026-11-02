/**
 * Share-token verification.
 *
 * The token lives in the URL *fragment* (`/hasil#t=...`), so it never reaches a
 * server log. Verification POSTs it in a body — never a path or query string.
 */

import { API_BASE, isMockMode } from './api/config';
import { decodeMockToken, isMockToken } from './api/mock';

export interface VerifiedPayload {
  id: string;
  ts: string;
  status: string;
  score: number | null;
  title?: string;
  claim?: string;
  srcs?: Array<{ pub: string; url: string; m: string }>;
  /** True for a locally generated preview token, never for a signed report. */
  preview?: boolean;
}

/**
 * `invalid` means the token itself did not verify: the report was altered or
 * the link is incomplete. `unavailable` means we could not get an answer at
 * all (rate limit, server error). Only the first may ever be worded as
 * "Laporan tidak sah".
 */
export type VerifyResult =
  | { valid: true; payload: VerifiedPayload }
  | { valid: false; reason: 'invalid' | 'unreachable' | 'unavailable'; retryAfter?: number };

/** Reads the token from the current fragment, if present. */
export function tokenFromHash(hash: string): string | null {
  const raw = hash.startsWith('#') ? hash.slice(1) : hash;
  if (!raw || raw.startsWith('?')) return null;
  const params = new URLSearchParams(raw);
  const token = params.get('t');
  return token && token.length > 0 ? token : null;
}

/** Builds the share URL with the token in the fragment. */
export function buildShareUrl(origin: string, token: string): string {
  return `${origin.replace(/\/$/, '')}/hasil#t=${encodeURIComponent(token)}`;
}

/** Verifies a token against the API. Never throws: an invalid token is a state. */
export async function verifyToken(token: string, signal?: AbortSignal): Promise<VerifyResult> {
  // Preview tokens are deliberately unsigned and only exist while the backend
  // is stubbed. They are still checked structurally, and the result page labels
  // them "Laporan contoh" instead of "Terverifikasi AMBA".
  if (isMockToken(token)) {
    const payload = decodeMockToken(token);
    return payload
      ? { valid: true, payload: { ...payload, preview: true } }
      : { valid: false, reason: 'invalid' };
  }

  if (isMockMode) {
    // A real signed report cannot be checked without the API. That is a
    // connectivity limitation, not evidence that the report was forged.
    return { valid: false, reason: 'unreachable' };
  }

  try {
    const res = await fetch(`${API_BASE}/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token }),
      signal
    });
    if (res.status === 429) {
      const retry = Number(res.headers.get('Retry-After') ?? '0');
      return { valid: false, reason: 'unavailable', retryAfter: retry || undefined };
    }
    if (res.status >= 500) {
      // The server failed; that says nothing about the report's authenticity.
      return { valid: false, reason: 'unavailable' };
    }
    if (!res.ok) return { valid: false, reason: 'invalid' };
    const body = (await res.json()) as { valid?: boolean; payload?: VerifiedPayload };
    if (!body.valid || !body.payload) return { valid: false, reason: 'invalid' };
    return { valid: true, payload: body.payload };
  } catch {
    // A network/CORS failure is NOT proof the report is forged. Saying so
    // would tell the user their evidence was tampered with when we simply
    // could not reach the server.
    return { valid: false, reason: 'unreachable' };
  }
}
