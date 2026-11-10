/**
 * API mode switch.
 *
 * The Go API is the real backend, but frontend work must not depend on it
 * running. `VITE_API_MODE=mock` (the default in `vite dev`) routes audits,
 * verification and replies through `lib/api/mock.ts`, which builds clearly
 * labelled example dossiers in the browser.
 *
 * A production build defaults to `http`; set `VITE_API_MODE=mock` explicitly
 * only for a frontend-only preview build.
 */
/**
 * Production builds default to a same-origin `/api/v1`, so a static host that
 * proxies the API (or a single-origin Docker deploy) works with no rebuild.
 * Development keeps pointing at the Go server on :8080.
 */
export const API_BASE =
  import.meta.env.VITE_API_BASE_URL ??
  (import.meta.env.DEV ? 'http://localhost:8080/api/v1' : '/api/v1');

const requested = import.meta.env.VITE_API_MODE;

export const API_MODE: 'mock' | 'http' =
  requested === 'mock' || requested === 'http'
    ? requested
    : import.meta.env.DEV
      ? 'mock'
      : 'http';

export const isMockMode = API_MODE === 'mock';
