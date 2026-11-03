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

## Phase 3 — Frontend shell & flows

| Item | Status | Evidence |
|---|---|---|
| UI primitives (Button/Panel/Field/StatusBadge/Skeleton) | Done | 44px targets, focus rings |
| Presentation modes (Jelajah/Esensial) | Done | store, switch, neutral first-visit picker, calm motion |
| SSE client + AuditSession | Done | stage/partial/verdict/done, heartbeat ignored, one retry |
| Zod schemas + schema_version guard | Done | "Perbarui aplikasi" on a newer contract |
| Home (hero bar, problem cards, step flow) | Done | rendered at 375/1440, 0 console errors |
| `/deteksi` (tabs, dropzone, presets, radar) | Done | radar driven only by real stage events |
| `/hasil` (verdict, honest limits, demo badge) | Done | in-memory store, reload shows "tidak tersimpan" |

## Phase 2 — Pipeline & integrations

| Item | Status | Evidence |
|---|---|---|
| `internal/forensics` server ELA | Done | q90 re-encode diff, 16x16 grid, format reliability |
| `internal/linkcheck` | Done | 29 URL cases, redirect unrolling |
| `internal/llm` + mock mode | Done | schema-constrained findings, deterministic mock |
| `internal/grounding` | Done | allowlist + tool-result verification |
| `internal/pipeline` | Done | errgroup, sub-deadlines, degradation, hashed cache |
| `/audit/{url,file,text}` JSON + SSE | Done | stage→partial→verdict→done |
| `internal/tactics` | Done | NFKC + whitespace-normalised substring check |
| `/verify`, `/reply` | Done | token-only input, template fallback |

## Phase 4 — Results experience

| Item | Status | Evidence |
|---|---|---|
| `ela.worker.ts` | Done | off-thread heatmap, row-band yields, 4096 px cap |
| `SplitSliderViewer` | Done | boxes outside clip layer, keyboard range, Asli/ELA tabs |
| `EvidenceAccordion` | Done | 4 pillars, "Tidak dapat diperiksa" instead of omission |
| Share + `/verify` | Done | token in fragment, POST body, verified/invalid/unreachable states |
| `ProofCardExporter` | Done | 1200x630, QR, lazy-imported deps, disclaimer line |
| `HighlightedText` | Done | verified phrases only, label + explanation popover |
| `ReplyCard` | Done | tone chips, template-first, offline-safe |

## Phase 5 — Edukasi, About, Accessibility

| Item | Status | Evidence |
|---|---|---|
| `/tentang` | Done | privacy manifesto, limitations, disclaimer |
| `/edukasi` + quiz | Done | 3 modus + 3 questions with explanations |
| Read-aloud (id-ID) | Done | feature-detected, stops on unmount |
| Text size toggle | Done | 100/125/150, persisted |
| Focus management | Done | focus moves to h1 on route change |
| axe scan | Done | 0 violations on all 5 routes |

## Phase 6 — Hardening & delivery

| Item | Status | Evidence |
|---|---|---|
| axe on all routes | Done | 0 violations, 5 routes |
| Bundle budget | Done | 102.2 KB gz / 250 KB |
| Security audit | Done | no URL/IP/text logging, panic path removed from logs |
| README | Done | architecture, env table, limitations |
| Focus management | Done | route-change only, not initial load |

## Known open questions

Recorded in `docs/DECISIONS.md`:
- Link-only audits return `INSUFFICIENT` even at link score 85 (weight 0.20 <
  0.30 threshold). Literal spec behaviour; needs a product decision.
- Mock mode cannot reach TAMPERED via the fact pillar, because it cites
  example.org and grounding rejects it. TAMPERED is covered by URL audits.

## In progress
- Nothing. All phases implemented; open questions await a product decision.

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

## Infrastructure & Tooling — Docker Containerization (2026-11-02)

| Item | Status | Evidence |
|---|---|---|
| Multi-stage `backend/Dockerfile` | Done | Static Go build (`CGO_ENABLED=0`), Alpine 3.21 runtime, non-root user `10001:10001` |
| Multi-stage `frontend/Dockerfile` | Done | Node 22 base, static SPA export, Nginx 1.27 Alpine production image |
| Hot-reload Dev config (`docker-compose.dev.yml`) | Done | Air Go watcher (`/tmp/air`) + Vite HMR (`0.0.0.0:5173`) with anonymous `node_modules` |
| Production / Demo stack (`docker-compose.yml`) | Done | Nginx reverse-proxies `/api/` to backend with unbuffered SSE streaming support |
| BuildKit layer caching | Done | Separate `go mod download` & `npm ci` with cache mounts for 2-5s rebuilds |
| Team ergonomics (`Makefile`) | Done | `make up`, `make dev`, `make build`, `make logs`, `make down` |
| Docker documentation (`docs/DOCKER.md`) | Done | Cross-platform setup for Linux, macOS Apple Silicon, Windows WSL2 |
| Design specifications relocated | Done | `docs/design-spec.md`, `docs/waspada-design-spec.md` |

## Visual redesign — neo-brutalist sticker system (2026-11-03)

Applied `docs/waspada-design-spec.md` to the whole SvelteKit frontend. The
design record is `docs/DESIGN-SYSTEM.md`; this table is the delivery log.

| Item | Status | Evidence |
|---|---|---|
| Token layer rewritten (sticker palette, type, hard shadows, motion) | Done | `src/lib/styles/tokens.css` |
| Sticker primitives (`.card`, `.btn`, `.pill`, `.lift`, `.sticker`, `.marquee`, `.nav-link`) | Done | `src/app.css` |
| Typeface swap: Bricolage Grotesque + Atkinson Hyperlegible, self-hosted | Done | `@fontsource` deps, CSP stays `font-src 'self'` |
| Inline stroke icon set (single source, no icon font, no emoji) | Done | `src/lib/components/ui/Icon.svelte` |
| Header: brand tile, one-line nav, text-size + mode tools | Done | `src/routes/+layout.svelte` |
| Home: hero + scanner example, marquee, steps, scam school bento, accessibility, CTA | Done | `src/routes/+page.svelte` + `src/lib/components/home/` |
| `/deteksi`, `/hasil`, `/edukasi`, `/cek-cepat`, `/tentang`, 404 restyled | Done | route files |
| Contrast gate extended to the sticker palette, `var()` aliases resolved | Done | 41 pairs, `npm run contrast` |
| Esensial mode collapses every flex row to one column | Done | `app.css`, verified at 150% text |
| First-visit mode picker made inline (no focus trap, no autofocus) | Done | `ModePicker.svelte` |
| axe: 0 violations on all 5 routes + demo result page | Done | `@axe-core/playwright`, WCAG 2.1 AA |
| Keyboard: skip link first, 14/14 stops with visible 4px ring | Done | Playwright tab walk |
| Hover: every primary target moves; zero colour-only primary targets | Done | Playwright hover delta |
| Tap targets: 0 below 44px at 390px | Done | Playwright measurement |
| Vitals: LCP 464 ms, CLS 0 on mobile viewport | Done | Playwright |
| No horizontal overflow at 360 / 390 / 768 / 1280 | Done | Playwright |
| e2e against the real Go backend | Done | 8/8 Playwright |

### Defects found by actually rendering it

1. **Mode picker displaced the hero.** It rendered above the hero and pushed
   the H1 below the fold at 390px. Moved after the hero, where it belongs as a
   secondary choice.
2. **Mode picker trapped Tab.** It was a `role="dialog"` with an autofocus and a
   focus loop, so the skip link and the whole page were unreachable by keyboard,
   and Tab cycled inside three buttons forever. It is an inline, non-modal
   choice; the trap and the autofocus are gone.
3. **Nav wrapped to two rows at 360px.** "Tentang" fell to its own line. Nav
   padding and size are tuned so all four links hold one line at 360px.
4. **`/deteksi` tabs wrapped 2+1 on phones.** `flex-basis: 0` alone cannot
   shrink below intrinsic width; `min-width: 0` was missing.
5. **Quiz number badges straddled the card border.** `legend` inside a bordered
   `fieldset` floats onto the border edge. Both quizzes now use a labelled
   `role="radiogroup"` / `role="group"` with the badge inside the card.
6. **Marquee text reported as 1.37:1 by axe.** An artifact of sampling a
   transformed track; the band is opaque ink and the real ratio is 16.9:1. The
   band colour is now restated on the animated subtree so the computed backdrop
   is unambiguous.
7. **`/hasil` eyebrow repeated its own H1** word for word.
8. **Pre-existing e2e ambiguity:** the demo banner and the proof-card stamp both
   read "CONTOH / SIMULASI", so `getByText` matched two nodes. The assertion now
   scopes to the banner (`role="note"`); the app is unchanged.

### Verification

```
frontend: svelte-check        0 errors, 0 warnings
frontend: eslint              no issues
frontend: vitest              56 tests pass
frontend: vite build          OK, adapter-static
frontend: contrast            41 pairs pass
frontend: size                119.3 KB gz / 250 KB
frontend: playwright e2e      8/8 against the real Go API
frontend: axe (WCAG 2.1 AA)   0 violations, 7 surfaces
frontend: mobile LCP 464 ms   CLS 0, no overflow at 360/390/768/1280
```

### Note

The scanner on the home page is a labelled worked example, not a live scan.
Badge collection is illustrative: there is no badge storage, by design (no
accounts, no persistence).
