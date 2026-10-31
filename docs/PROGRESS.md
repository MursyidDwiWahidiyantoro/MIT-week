# Progress

_Last updated: 2026-10-31_

## Phase 0 — Scaffold & contract

| Item | Status | Evidence |
|---|---|---|
| Monorepo layout (Tech Spec §1) | Done | `api/ backend/ frontend/ configs/ docs/ tools/` |
| `README.md` with run instructions | Done | `README.md` |
| `.env.example` | Done | mirrors Tech Spec §10 |
| `.gitignore` | Done | ignores `.env`, build output, artifacts |
| `api/openapi.yaml` | Done | 7 paths, 11 schemas, parses clean |
| `api/dossier.schema.json` | Done | JSON Schema 2020-12, generated from OpenAPI |
| Go module + chi router + `/healthz` | Done | `go test -race ./...` green |
| Config loader, fail fast | Done | `internal/config` + 6 tests |
| Error envelope + panic recovery | Done | `internal/httpapi` + 4 tests |
| SvelteKit 3 / Svelte 5 / TS / Tailwind SPA | Done | `npm run build` green |
| Routes `/ /deteksi /hasil /edukasi /tentang` | Done (placeholders) | build output |
| `tokens.css` per DESIGN §7 | Done | `src/lib/styles/tokens.css` |
| `tools/check-contrast.mjs` in CI | Done | 22 pairs pass |
| Bundle size gate (≤250 KB gz) | Done | 42.1 KB gz |
| CI: go test -race, FE gates, contract, demo diff | Done | `.github/workflows/ci.yml` |

**Phase 0 done-when:** both apps boot ✔ · CI green (not yet run on GitHub) · OpenAPI validates ✔

## Bugs found and fixed during Phase 0 verification

Three real defects, each caught by actually rendering the built app:

1. **Blank page (SPA never hydrated).** SvelteKit injects an inline bootstrap
   script; the hardcoded `script-src 'self'` blocked it. No visible error, just
   an empty screen. Fixed by moving CSP into `sveltekit({ csp: { mode: 'hash' } })`
   so each build's script hash is included. `security-headers.json` keeps only
   `frame-ancestors` (ignored in `<meta>`) plus the non-CSP headers.
2. **Unstyled page.** `tokens.css` was never imported, so every `var(--…)` fell
   back to browser defaults (serif body, 16 px h1, no nav gap). Fixed by
   importing tokens before `app.css`.
3. **Build failure on a clean checkout.** `tsconfig.json` extended
   `.svelte-kit/tsconfig.json`, which SvelteKit 3 no longer generates; the
   supported form is `extends: "$app/tsconfig"`.

## Verification run locally

```
backend : go test -race ./...            PASS
backend : curl /healthz                  {"status":"ok"} + security headers
frontend: svelte-check                   0 errors, 0 warnings
frontend: eslint                         0 errors
frontend: vitest                         6 tests pass
frontend: vite build                     OK, adapter-static "build"
frontend: contrast                       22 pairs pass
frontend: size                           42.1 KB gz / 250 KB
frontend: preview + CSP headers          served 200, headers applied
frontend: hydrated render check          H1 renders, 0 console errors
frontend: computed styles                Geist loaded, bg #0b0f14, h1 52px,
                                         nav gap 24px, CTA min-height 44px
contract: api/dossier.schema.json        draft 2020-12, validates
```

## In progress
- Phase 1 backend: scoring, security, token, demo — done locally, pending full gate run.
- Phase 3 frontend: not started.

## Phase 1 — Backend core

| Item | Status | Evidence |
|---|---|---|
| `internal/scoring` | Done | 68 tests, all 8 §3.5 vectors |
| `internal/security` (SSRF) | Done | 20 blocked-address cases + redirect/DNS-rebinding tests |
| `internal/security` (upload) | Done | magic bytes, dims, animation, size — 13 cases |
| `internal/security` (ratelimit) | Done | burst, refill, proxy trust, sweep, in-flight cap |
| `internal/security` (headers/CORS) | Done | allowlist only, no credentials, preflight |
| `internal/token` | Done | round-trip, tamper, truncation, rotation, size |
| `internal/demo` + `/demo/{id}` | Done | 3 cases, schema-validated, example.org only |

## Phase 0 close-out
All Phase 0 items are done and verified locally. CI has not yet run on GitHub
(the repo lives inside the user's `~/Projects` git root; no remote configured).

## Blocked
- None.

## Next (Phase 1)
Scoring (`internal/scoring`) with the §3.5 vectors, SSRF client, token signing,
demo data, then Phase 2 pipeline.

## Decisions
See `docs/DECISIONS.md`.
