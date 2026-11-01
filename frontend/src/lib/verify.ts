/**
 * Share-token verification.
 *
 * The token lives in the URL *fragment* (`/hasil#t=...`), so it never reaches a
 * server log. Verification POSTs it in a body — never a path or query string.
 */

export interface VerifiedPayload {
  id: string;
  ts: string;
  status: string;
  score: number | null;
  title?: string;
  claim?: string;
  srcs?: Array<{ pub: string; url: string; m: string }>;
}

export type VerifyResult =
  | { valid: true; payload: VerifiedPayload }
  | { valid: false; reason: 'invalid' | 'unreachable' };

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1';

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
  try {
    const res = await fetch(`${API_BASE}/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token }),
      signal
    });
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
