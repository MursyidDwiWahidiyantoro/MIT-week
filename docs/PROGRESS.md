# Progress

_Last updated: 2026-11-08_

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

---

## Realignment audit — PRD v9 vs repo v10.1 (2026-11-05)

The downloaded `PRD_AMBA.md` is **v9.0.0**. The repo already carries **v10.1**,
which explicitly supersedes it. v10 reverses several v9 decisions on purpose:
React 19 → Svelte 5/SvelteKit, blue glassmorphism → warm technical sticker
system, client-side ELA score → server-authoritative ELA, golden-cache
fallback → degraded real verdict, `/hasil/[id]` → signed fragment token.
Realigning to v9 would therefore have been a regression; the work below closes
the gap to **v10.1** instead, and records every deviation in `DECISIONS.md`.

### Fixed

| Item | Before | After |
|---|---|---|
| WCAG contrast gate | 4 failures (ink-on-lime 1.73:1 etc.) | 82 pairs pass in **both** themes |
| Status badges | dark text on dark fill in dark mode | semantic `--on-fill` / `--on-panel-dark` / `--on-brand` |
| Display font | `--font-display` pointed at Helvetica, Bricolage never used | Bricolage Grotesque Variable |
| Essential mode type | root font grew, H1 pinned by `vw` | every clamp has a `rem` term; e2e asserts it |
| Null arrays in JSON | `anomalies/redirect_chain/reasons/sources` = `null` | normalised to `[]`; Go test scans the wire format |
| `Retry-After` | unreadable cross-origin | `Access-Control-Expose-Headers` + test |
| ELA corroboration | hotspot grid computed then discarded | wired into `scoring.Input`; anomaly boxes attached after `Wait` |
| Injection guard | flag hardcoded `false` | pipeline carries `findings.InjectionSuspected` |
| Redirect chain | one-element echo | real unroll via the SSRF-safe client |
| Text audit link pillar | never ran | first URL in the message is analysed (F4) |
| Hero link bar (F1) | CTA only | real `GET /deteksi?q=` form |
| Mode + text-size controls | components existed, never imported | wired into the header; picker after hero |
| About team (F10) | missing | four-member team section |
| Proof card palette | old indigo hexes | implemented warm/dark tokens |
| PWA manifest colours | old indigo | implemented palette |
| axe gate | claimed in docs, dependency never imported | Playwright gate: 6 surfaces × 2 themes |
| `/reply` fallback | ignored verdict status | status × tone matrix, 4 statuses |
| `trusted_sources.yaml` | referenced by spec, absent | present, with a Go parity test |
| `ModePicker` / `TextSize` / `ModeSwitch` | dead components | live |

### Still open

- `/reply` uses deterministic templates, not an LLM draft.
- Link-only audits can still return `INSUFFICIENT` at score 85 (product decision).
- Safe Browsing and metrics env vars are read but unused.
- Demo case IDs differ from PRD F8 (`contoh-*` vs `demo-tiktok` etc.); recorded
  in `DECISIONS.md`.

### Verification (2026-11-05)

```
backend : go vet ./...                 no issues
backend : go test -race ./...          264 tests pass, 14 packages
frontend: svelte-check                 0 errors, 0 warnings
frontend: eslint                       no issues
frontend: vitest                       56 tests pass, 9 files
frontend: vite build                   OK, adapter-static
frontend: contrast                     82 pairs, both themes
frontend: size                         120.2 KB gz / 250 KB
frontend: playwright e2e               14/14 (incl. axe on 6 surfaces)
demo diff                              in sync
```

## Production readiness — PRD P0/P1/P2 complete (2026-11-08)

Goal: every PRD requirement built and the app deployable. Delivered as a
commit series on top of `1bef567` (`git log 1bef567..`).

| Area | Before | Now |
|---|---|---|
| Link-only verdict | `INSUFFICIENT` even for a clear phishing link | `LINK_STRONG` floor (team decision), vectors 9–12 |
| Lookalike detection | Levenshtein on the whole host | token-based, Cyrillic/Greek fold, official domains, `.go.id`-style suffixes |
| Safe Browsing | key read, never used | checked on the whole redirect chain; a hit sets `LINK_MALICIOUS` |
| Fact grounding | model-cited sources | Google Fact Check Tools search in Go, second classify call, fail closed |
| Link content (F1) | not read | oEmbed/OpenGraph preview; BLOCKED/UNAVAILABLE explained, screenshot route offered |
| Screenshots (F2) | ELA only | image sent to the model, OCR text, URL in the OCR gets the link pillar, OCR tactics |
| SSE | verdict at the end | stage/partial per pillar as it finishes, heartbeat, per-step budgets |
| Balas Sopan (F12) | templates only | validated model draft, template fallback |
| Demo (F8) | `contoh-*`, hand-written | PRD IDs + `demo-tautan`, generated with the real scorer, real screenshot |
| Edukasi (F7) | 3 modus + 3 questions | + "Bedah forensik" quiz with real ELA heat maps |
| EN/ID toggle (P1) | — | 500+ keys, Indonesian default, English verdict rebuilt from codes |
| Metrics (P1) | env var read, unused | anonymous Prometheus counters on an internal listener |
| Telegram (P2) | — | webhook/polling bot on the same pipeline |
| Performance | CLS 0.134, no offline shell | CLS 0, fonts preloaded, offline app shell, Lighthouse gate |
| Deployment | out of scope | Vercel + Fly.io/Railway configs, `docs/DEPLOY.md`, production boot checks |
| Operations | silent upstream failures | reason + status logged, never URLs or keys |

### Defects found while building

- Telegram accepted `http://127.0.0.1/...` as a link to check: `ValidateFetchURL`
  now refuses internal targets before any request.
- Empty text was answered with the invalid-URL message.
- CSP `connect-src` was hard-coded to `http://localhost:8080`, which would have
  blocked every API call in production.
- The mobile header wrapped into three rows once the language switch was
  added: the display controls now fold behind one toggle.
- The screenshot quiz explanation described an even glow while the heat map
  shows the yellow text brightest; the copy now explains that contrast, not
  editing, makes it glow.

### Manual checklist (cannot be automated)

- [ ] Android Chrome: install the PWA from the production URL, share a link
      from another app to AMBA, confirm the audit starts and the shared item is
      gone from Cache Storage afterwards.
- [ ] Android Chrome: share a screenshot to AMBA.
- [ ] iOS Safari: paste button reads the clipboard after a tap.
- [ ] Telegram: `/start`, `/cek <link>`, a forwarded text, a photo; the reply
      link opens the signed report.
- [ ] Real keys: the boot `readiness` log shows `llm_model_listed=true`,
      `fact_search=true`, `safe_browsing=true`.

### Verification (2026-11-08)

```
backend : gofmt -l .                   clean
backend : go vet ./...                 no issues
backend : go test -race ./...          385 tests pass, 16 packages
backend : go run ./cmd/gendemo         deterministic, no diff
frontend: svelte-check                 0 errors, 0 warnings
frontend: eslint                       no issues
frontend: vitest                       60 tests pass, 10 files
frontend: vite build                   OK, adapter-static
frontend: contrast                     82 pairs, both themes
frontend: headers                      vercel.json matches security-headers.json
frontend: size                         142.0 KB gz / 250 KB (all JS)
frontend: playwright e2e               32/32 (axe ID + EN, mode parity, i18n)
frontend: lighthouse (mobile, median 5) perf 93-100, a11y 100, BP 96-100, SEO 100, CLS 0
demo diff                              in sync
```

## Real-sample evaluation (2026-11-09)

Run from a throwaway harness on the pillars that work without external
services (the model, Fact Check Tools and Safe Browsing are unreachable from
the development environment). Details and data sources in
`docs/DECISIONS.md` ("Link heuristics tuned on real phishing and legitimate
URLs").

| Set | Size | Warning verdict before | After tuning |
|---|---|---|---|
| Indonesian-themed phishing links | 1,053 | 25.4 % | 83.6 % |
| Phishing links on `.id` domains | 10,406 | 2.2 % | 13.1 % |
| Random global phishing links | 3,000 | 2.4 % | 4.5 % |
| Top 10,000 domains (legitimate) | 10,000 | 0 | 4 |
| Official Indonesian URLs (legitimate) | 60 | 0 | 0 |

| Image | Truth | ELA score |
|---|---|---|
| TruFor pristine1.jpg / pristine2.jpg | authentic | 0 / 0 |
| CASIA Sp_S_NNN_A_ani0041 (JPEG) | edited | 2 |
| CASIA Tp_S_NNN_S_N_ani20009 (lossless) | edited | 5 |
| TruFor tampered1.png / tampered2.png | edited | 5 / 14 |
| CAT-Net example.jpg | edited | 0 |

Still to run with real keys: the model pillar, fact checking and Safe
Browsing (`docs/DEPLOY.md` §6).

## Scam rules (2026-11-09)

`internal/scam` adds a deterministic `SCAM_PATTERN` floor (TECH_SPEC §6.3,
`docs/DECISIONS.md`). Each set below was written before the rules it led to
and measured once on first sight:

| Set | Scams caught | False alarms in ordinary messages |
|---|---|---|
| Documented scams and quoted warnings (built with the rules) | 29 / 29 | 0 / 35 |
| Holdout 1 | 13 / 20 | not recorded |
| Holdout 2 | 13 / 15 | 0 / 15 |
| Set 3: account threats, new fees, customs | 0 / 15 | 5 / 25 |
| Set 4 | 11 / 20 | 0 / 20 |
| Set 5 | 10 / 20 | 0 / 20 |
| Set 6 | 14 / 25 | 1 / 25 |

All sets pass now and stay as regression tests; 3 messages are still left
to the model (a fake OJK officer, a "finance department" transfer request,
an English account notice).

## One-sided ELA and the AI rubric (2026-11-09)

- ELA counts only from 50 up; a low ELA no longer lowers a score or helps an
  image reach SAFE, and SAFE needs a fact check (vectors 15-17).
- The model's `ai_score` is a manipulation-risk rubric; the pillar is shown as
  "Penilaian AI" / "AI assessment".
- The demo news screenshot (ELA 3) now scores 90 instead of 85: its low ELA
  no longer dilutes the fact check.
