# CLAUDE.md — AMBA build playbook

You are building **AMBA** (Aplikasi Mengenali & Bertindak Aman), a hoax/phishing checker web app for the MIT-Week 2026 competition.

## Workspace rules

Treat `/home/riko/Projects/AMBA` as your only workspace.

- Read, create, edit, and delete project files only within this directory (including `instruction/`).
- Do not access or modify files outside this directory, including global configurations (`~/.claude`), other projects, and the user's home-directory files.
- Do not run commands that modify files outside the workspace.
- If a task requires access outside this directory, stop and ask first.
- Keep all generated files and temporary project artifacts inside this directory.
- Before making changes, confirm the workspace path and inspect the project structure. Follow these rules throughout the entire session.

## Read order (before writing any code)
1. `PRD_AMBA_v10.md` — what/why, acceptance criteria. 2. `TECH_SPEC_AMBA.md` — contracts, scoring, security (source of truth). 3. `DESIGN.md` — visual and interaction system (authoritative for look-and-feel). 4. This file — how to work.
If anything conflicts with the old v9 PRD, the v10.1 docs win. If v10 docs conflict with each other, `TECH_SPEC_AMBA.md` wins; note the conflict in `docs/DECISIONS.md`.

## Stack & commands
```bash
# backend (Go 1.23, chi)
cd backend && go run ./cmd/server           # dev (LLM_MOCK=true by default)
cd backend && go test -race ./...           # must pass before any commit
cd backend && go vet ./... && golangci-lint run   # if installed
# frontend (Svelte 5 + SvelteKit SPA via adapter-static, Vite, TS, Tailwind)
cd frontend && npm i && npm run dev
cd frontend && npm run typecheck && npm run lint && npm run test
cd frontend && npm run test:e2e             # Playwright, uses LLM_MOCK backend
cd frontend && npm run build && npm run size   # fails if initial JS > 250 KB gz
```
Create these scripts in Phase 0 if missing.

## Hard rules (never violate)
1. **No accounts, no database, no persistence of user input** (images, text, URLs). In-memory TTL cache of final verdict JSON only. Single exception: the PWA share-target hand-off (Tech Spec §8.1) may hold a shared item in client Cache Storage for seconds, deleted on read and swept at boot.
2. **Never log** IPs, submitted URLs, text, titles, filenames. Only the anonymous line format in Tech Spec §7.6.
3. **Verdict is computed in Go** (`internal/scoring`). The LLM never outputs a score/status. Do not "simplify" this.
4. **Never serve golden/demo data for a live request.** Timeouts → degraded real verdict.
5. **Do not invent URLs, publishers, or sources** anywhere (code, demo data, copy). Demo sources are placeholders labelled "Sumber Contoh".
6. **All user-facing text is Indonesian**, plain and friendly; no jargon like "DCT", "ELA" without a one-line explanation.
7. **Never claim certainty.** Use "terindikasi", "tidak ditemukan tanda bahaya"; keep the disclaimer in footer + proof card.
8. Don't add scope outside the PRD (no auth, PDF analysis, video analysis, auto-report bots, new frameworks, any datastore/"Scam Radar"). **Do not do deployment, hosting, domain or Docker work** — local development only for now.
9. Don't hardcode secrets or the model name; use env. Never commit `.env`.
10. Don't weaken a security test to make it pass — fix the code.
11. **Tactics and replies are informational**: they never change the score; a tactic is shown only if its phrase is verified as a substring of the input; reply text may contain only verified-source URLs and the share link.
12. **Frontend is Svelte 5 + SvelteKit — not React.** Follow `DESIGN.md`: dark atmospheric technical look, restrained warm accent, **two presentation modes (Jelajah/Esensial) chosen by preference and never labelled by age/ability**; both modes expose identical information and functionality; reduced-motion honoured in both. Colour values in Tech Spec §8.0 are provisional until the team validates them against the approved visual reference.

## Working method
- **Contract first:** Phase 0 produces `api/openapi.yaml` + Dossier JSON Schema; Go structs and Zod schemas must match it. Any contract change updates all three in the same commit.
- Work phase by phase. Finish a phase's acceptance checks before starting the next. Commit per logical task: `feat(scoring): …`, `fix(ssrf): …`, `test(link): …`, `docs: …`.
- Keep a running `docs/PROGRESS.md` (done / in progress / blocked / decisions) — update at the end of every task.
- Prefer small pure functions with table tests. Write the test first for scoring, SSRF, link heuristics, token.
- Use sub-agents for independent tracks once Phase 0 is merged: **backend track** (Phases 1–2) and **frontend track** (Phases 3–5) can run in parallel against the mock API; an **explore/review agent** may audit security and a11y at Phase 6. Give each agent the exact phase section and the files it owns; avoid two agents editing the same file.
- If blocked by a genuinely open question (PRD §9), choose the safest default, record it in `docs/DECISIONS.md`, continue.

## Backlog

### Phase 0 — Scaffold & contract
- [ ] Monorepo layout per Tech Spec §1; `README.md` with run instructions; `.env.example`; `.gitignore`.
- [ ] `api/openapi.yaml` (routes, SSE events, error envelope, Dossier v2) + JSON Schema.
- [ ] Go module, chi router skeleton, `/healthz`, config loader (fail fast on bad env).
- [ ] SvelteKit (Svelte 5, TS, `adapter-static` SPA fallback) + Tailwind app with routes `/`, `/deteksi`, `/hasil`, `/edukasi`, `/tentang` (placeholders); `src/lib/styles/tokens.css` per `DESIGN.md` §7 / Tech Spec §8.0 and `tools/check-contrast.mjs` wired into CI.
- [ ] CI (GitHub Actions): go test -race, FE typecheck/lint/test/build+size, demo-json diff check.
**Done when:** both apps boot; CI green; OpenAPI validates.

### Phase 1 — Backend core (pure + security)
- [ ] `internal/scoring` per Tech Spec §3 + all vectors §3.5 (recompute arithmetic; correct the spec table if wrong).
- [ ] `internal/security`: SSRF client (§7.1) with ≥15-case tests; upload validation (§7.2); rate limiter + proxy trust (§7.3); API headers + CORS (§7.4).
- [ ] `internal/token` (§7.5) with tamper tests.
- [ ] `internal/demo` + embedded `demo-cases.json` per §9; `GET /demo/{id}` with regex param validation.
- [ ] Error envelope + panic-recovery middleware (no 500 leaks, no stack in body).
**Done when:** `go test -race ./...` green; demo route returns valid Dossier < 50 ms.

### Phase 2 — Pipeline & integrations
- [ ] `internal/forensics` server ELA + hotspot grid (§3.4).
- [ ] `internal/linkcheck` rules + redirect unrolling + optional Safe Browsing (§5), ≥25 URL tests.
- [ ] `internal/llm`: 9router client, JSON-schema findings, tool loop, **mock mode**, injection flags (§6).
- [ ] `internal/grounding` allowlist verification + `configs/trusted_sources.yaml`.
- [ ] `internal/pipeline` errgroup w/ sub-deadlines, degradation, cache (§4).
- [ ] Handlers `/audit/url|file|text` with JSON and SSE modes; `/verify`.
- [ ] `internal/tactics` (enum, normalized-substring verification) and deterministic `explanation_simple` templates (§6.1, §8.2) with tests.
- [ ] `internal/reply` + `POST /reply` (§6.2): token-only input, URL/length validation, template fallback, `configs/reply_templates.json`.
**Done when:** with `LLM_MOCK=true` all three audit routes return correct Dossiers; SSE emits stage→partial→verdict→done in order; killing the mock LLM yields `degraded:true`, never HTTP 500.

### Phase 3 — Frontend shell & flows
- [ ] Layout (Navbar/Footer with disclaimer), Home with `HeroInstantBar`, `ProblemCards`, `StepFlow`.
- [ ] `/deteksi`: `ModeSelectorTabs`, `DropzoneUploader` (type/size validation, friendly errors), `JurorPresetBar`, `ScanningRadar` driven by real SSE stage events (not fake timers).
- [ ] `AuditSession` (`$lib/audit.svelte.ts`, Svelte 5 runes) SSE client + Zod schemas + error/empty/offline/429 states; announce stage changes through a polite live region.
- [ ] Demo buttons work **offline** from bundled JSON.
- [ ] `frontend/security-headers.json` + Vite plugin applying it in dev/preview (no host-specific files yet).
- [ ] PWA manifest + `sw.js` share target + clipboard-paste fallback (§8.1); unit-test the SW handler with fake `FormData`; add manual Android checklist to `docs/PROGRESS.md`.
- [ ] Presentation modes (Jelajah/Esensial): `mode.svelte.ts`, `ModeSwitch`, neutral first-visit picker, `data-info-key` on every information item, Playwright **parity test**, calm-motion handling (§8.0, §8.2).
**Done when:** link/file/text flows reach `/hasil` against the mock backend; demo works with network disabled.

### Phase 4 — Results experience
- [ ] `ela.worker.ts` (§8) + object-URL cleanup; downscale notice.
- [ ] `SplitSliderViewer` with fixes (boxes outside clip layer, accessible range, keyboard, alt tabs) + anomaly text list.
- [ ] `VerdictCard`, honest-limits panel, `EvidenceAccordion` (4 pillars, progressive skeletons).
- [ ] Share: signed-token link in fragment, `/verify` call, valid/invalid states; WhatsApp share button.
- [ ] `ProofCardExporter` 1200×630 PNG with QR mini-token, offline-safe fonts, disclaimer line.
- [ ] `HighlightedText` tactic highlights with static `tactics.json` labels (§8.3).
- [ ] `ReplyCard` Balas Sopan: tone chips, template-first then LLM swap, Copy + WhatsApp, offline demo templates (§8.4).
**Done when:** Playwright covers upload→slider→export; tampered token shows "Laporan tidak sah".

### Phase 5 — Edukasi, About, Accessibility
- [ ] `/tentang` (team, privacy manifesto stating exactly what is/isn't sent to AI gateway, limitations, disclaimer).
- [ ] `/edukasi` catalog + 3-question quiz with synthetic images and precomputed ELA reveal (P1; do after a11y).
- [ ] (P1) `/cek-cepat` + Sandi Keluarga printable card (§8.5): client-only, offline, with tests.
- [ ] Read-aloud (id-ID), text-size toggle (persisted), reduced-motion, focus management on route change, skip-link.
- [ ] Copy pass: every error/empty state in warm plain Indonesian; plain, easy-to-read wording.
**Done when:** axe: 0 serious/critical on all routes; manual keyboard-only run of main flow succeeds.

### Phase 6 — Hardening & delivery
- [ ] E2E suite per Tech Spec §11; Lighthouse CI ≥ 90 all categories; bundle ≤ 250 KB gz.
- [ ] Design review against `DESIGN.md` §9 in both modes at phone, tablet and desktop widths (spawn a separate review agent).
- [ ] Security review pass (spawn a separate review agent): SSRF, upload, CSP config, CORS, rate-limit bypass attempts, `/reply` URL stripping, share-target cleanup, log audit (grep for URL/IP logging).
- [ ] `README.md`: architecture diagram, env table, local run instructions, pitch demo script, known limitations (e.g. iOS has no Web Share Target). Deployment docs are out of scope for now.
- [ ] `docs/DECISIONS.md` and `docs/PROGRESS.md` finalized.
**Done when:** all Done-When boxes above are checked and `docs/PROGRESS.md` lists no open P0 items.

## Definition of done (every task)
- Tests added/updated and passing (`go test -race`, `npm run test`).
- Types/lint clean; no `any` without a comment; no `TODO` without a `docs/PROGRESS.md` entry.
- No new network call from the browser except to `VITE_API_BASE_URL`.
- Copy is Indonesian; a11y: focus order, labels, contrast ≥ 4.5:1, targets ≥ 44 px.
- Contracts (OpenAPI ↔ Go ↔ Zod) in sync.
- UI changes pass the `DESIGN.md` §9 checklist in **both** presentation modes.

## Kickoff prompt (paste to start)
> Read CLAUDE.md, PRD_AMBA_v10.md, TECH_SPEC_AMBA.md and DESIGN.md. Execute Phase 0 completely, commit, update docs/PROGRESS.md, then report. After my approval, run the backend and frontend tracks in parallel with sub-agents.
