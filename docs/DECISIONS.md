# Decisions

Record of choices made where the spec left room, plus any conflict found between docs.
Format: date · decision · why · what it would take to revisit.

## 2026-10-31 · Workspace scoped to /home/riko/Projects/AMBA
Project instructions now restrict all file access to this directory. Global
configuration is out of bounds unless the user explicitly approves a one-off.
Revisit: only if the user widens scope.

## 2026-10-31 · TypeScript 6.0.3 (not 5.x)
`@sveltejs/kit@3.1.1` declares a peer requirement of `typescript@^6.0.0`, so 5.9
fails installation. Revisit: when Kit relaxes the peer range.

## 2026-10-31 · Tailwind 3.4 (not 4.x)
Tailwind 4 moves configuration into CSS (`@theme`), which conflicts with the
token contract in TECH_SPEC §8.0 that expects `tailwind.config.ts` mapping to
CSS custom properties. Tailwind 3 keeps that mapping explicit and testable.
Revisit: if the team wants Tailwind 4, the token mapping must be rewritten and
`tools/check-contrast.mjs` re-pointed at the `@theme` block.

## 2026-10-31 · `@lucide/svelte` instead of `lucide-svelte`
`lucide-svelte@1.0.1` is deprecated upstream; `@lucide/svelte` is the maintained
package with the same icon set and stroke control. TECH_SPEC §8 only requires
one icon set with a consistent stroke weight, which this satisfies.

## 2026-10-31 · Fonts self-hosted via @fontsource-variable
TECH_SPEC §8 requires self-hosted, OFL-licensed fonts with `font-display: swap`.
Geist and Geist Mono (OFL) are installed as packages rather than fetched at
runtime, so no network call leaves the browser for fonts.

## 2026-10-31 · SPA mode via `ssr = false` + adapter-static fallback
Per TECH_SPEC §8 routing. `prerender = false` because every route reads
in-memory session state or a URL fragment; nothing is statically generated.

## 2026-10-31 · SvelteKit 3 removed `$lib` — use `#lib` (or a relative path)
`@sveltejs/kit@3` throws `module_removed_lib` on any `$lib/...` import. The
supported alias is `#lib`. Phase 0 uses a relative import for the one CSS
import; new modules should use `#lib/...`. Revisit: if the team prefers to keep
`$lib` for familiarity, add `alias: { '$lib': 'src/lib' }` to the `sveltekit()`
plugin config.

## 2026-10-31 · CSP is emitted by SvelteKit, not hardcoded in security-headers.json
SvelteKit injects an inline bootstrap script. A static `script-src 'self'`
blocked it and the SPA rendered a blank page with no console error beyond the
CSP violation. The policy now lives in the `sveltekit({ csp: { mode: 'hash' } })`
config so each build's inline script hash is included automatically.
`security-headers.json` keeps only the directives that must arrive as real HTTP
headers: `frame-ancestors 'none'` is ignored inside `<meta>` and therefore
cannot be part of the SvelteKit meta policy. Two policies on one response
intersect, so the header policy carries `frame-ancestors` alone — adding
`script-src` there would re-block the hashed inline script.

## 2026-10-31 · `tokens.css` must be imported before any consumer
Design tokens are CSS custom properties; a rule using `var(--x)` before the
definitions load silently falls back to browser defaults (serif text, 16 px
headings). The layout imports `tokens.css` first, then `app.css`.

## 2026-10-31 · Go module pinned to 1.24, not the toolchain's 1.26
Adding `golang.org/x/time` ran `go mod tidy` with a local Go 1.26 toolchain,
which rewrote the directive to `go 1.26.0` — newer than CI's pinned 1.24, so CI
would have failed on a clean runner. Pinned back to `go 1.24`; the module
builds and tests clean at that version. Revisit: when CI moves to 1.26.

## 2026-10-31 · TAMPERED needs a fallback explanation_simple
TECH_SPEC §8.2 defines `explanation_simple` only for the three floors,
SUSPICIOUS, INSUFFICIENT and SAFE. A weighted score can reach TAMPERED without
any floor firing (for example ELA 69 + AI 100 + UNVERIFIED + link 89), so
`scoring.ExplanationSimple` falls back to "Beberapa tanda bahaya ditemukan. Cek
dulu sebelum menyebarkan." — calm, non-certain, and consistent with the
vocabulary contract. Revisit: if the team prefers to define this case in the
spec, update §8.2 and the test table together.

## 2026-10-31 · ELA_SOLO_CAP only applies when ELA is available
TECH_SPEC §3.3 floor 4 reads literally as "if no other pillar scored >= 50 and
no floor 1-3 applied, cap at 60" — with no mention of ELA availability. Read
literally, an input where ELA could not run at all (AI 100 + link 49) would be
capped to 60 and downgraded to SUSPICIOUS, suppressing two real signals because
a third was missing.

The implementation guards the cap with `ELA.Available`, matching the floor's
name and its intent (stop ELA acting alone from reaching TAMPERED). Covered by
the test "no cap when ela is unavailable". Revisit: if the team wants the
literal reading, remove the guard and update §3.3 wording and that test.

## 2026-11-01 · ELA is blind to images already saved at the analysis quality
ELA compares the submitted image against a q=90 re-encode of itself. If the
input is already a q=90 JPEG the re-encode is effectively a no-op, every
difference rounds to zero, and the pillar reports a score of 0 even for a
visibly composited image. Real screenshots and re-saved photos carry a
different compression history, so this mostly affects synthetic test fixtures
— but it is a real ceiling on the technique.

Consequences: (a) unit tests must encode fixtures at a quality other than 90,
(b) the UI must never present a 0 ELA score as "clean" — reliability 0.0 and
the honest-limits panel exist for this, (c) the demo assets must be generated
with a realistic compression history or their heatmaps will be empty.

## 2026-11-01 · Fact status is derived from verified sources, never claimed by the model
The model reports a `fact_status`, but the pipeline only keeps it when at least
one source survived grounding with the matching verdict. A `CONTRADICTS` claim
backed solely by a non-allowlisted or invented URL is downgraded to
`UNVERIFIED`, which caps the fact pillar at 40 and prevents the
FACT_CONTRADICTED floor from firing.

Found by a live smoke test: a text audit scored 97 / TAMPERED while its only
"source" was `https://example.org/contoh-cek-fakta` marked UNVERIFIED. The
score now lands at 47 / SUSPICIOUS, which is what the evidence supports.

Consequence for mock mode: the mock cites `example.org`, so it can never reach
TAMPERED through the fact pillar. That is intentional — inventing a real
publisher URL to make a test look better would break CLAUDE.md rule 5. TAMPERED
coverage in e2e comes from a synthetic lookalike domain in a URL audit
(`https://bca-login.xyz/...`), where the LINK_MALICIOUS floor fires on real
heuristics rather than a fabricated citation.

Demo data still shows CONTRADICTS with `example.org` sources because TECH_SPEC
§9 requires demo sources to be placeholders labelled "Sumber Contoh"; demo
dossiers are pre-baked and never pass through grounding.

## 2026-11-01 · OPEN — link-only audits always return INSUFFICIENT
Verified behaviour: a URL audit whose only available pillar is Link returns
`INSUFFICIENT` with `score: null`, even when the link score is 85 and the
reasons include a lookalike bank brand, a suspicious TLD and a verification
path. Cause: Link carries base weight 0.20, and W = 0.20 < 0.30, so the
formula yields INSUFFICIENT before any floor can apply. Only a link score
>= 90 reaches the LINK_MALICIOUS floor (spec §3.3 floor 2).

This is literally what §3.2/§3.3 specify, so it is NOT a bug in the code. But
two product questions are unresolved and must be decided before Phase 6:

1. Should a URL audit with a link score in the 85-89 band produce a verdict?
   Options: lower the floor threshold to 85, add a weight floor for Link, or
   accept INSUFFICIENT. Any change must update TECH_SPEC §3.3 and the test
   vectors together.
2. Trust UX: "we checked the link and found nothing wrong" currently renders
   as "Belum cukup data", which reads like a failure rather than a completed
   check. `amba-trust-ux` should get an explicit INSUFFICIENT-because-single-
   pillar case before the copy pass.

## 2026-11-01 · Mock mode cannot reach TAMPERED via the fact pillar
Follows from the verified-source rule above. Mock cites example.org, grounding
rejects it, fact stays UNVERIFIED. TAMPERED coverage comes from a URL audit
against a synthetic lookalike domain. Do not "fix" this by allowlisting
example.org or inventing a real publisher.

## 2026-11-01 · Session paused at Phase 2 complete, Phase 4 not started
Backend through the pipeline and all audit/verify/reply routes is implemented
and green. Frontend is at Phase 3. Next: Phase 4 (results experience), then
the open questions above.

## 2026-11-01 · OPEN · Server serialises empty slices as null, breaking the contract

Found while verifying the live audit flow, not by a test — which is why the
frontend now has a Playwright suite.

`backend/internal/models/dossier.go` declares `Anomalies []Anomaly`,
`RedirectChain []string`, `Reasons []LinkReason` and `Sources []Source` as
slices. Go marshals a nil slice as `null`, so every live audit emits:

```json
"ela":  {"anomalies": null}
"link": {"redirect_chain": null, "reasons": null}
"fact": {"status": "", "sources": null}
```

`api/dossier.schema.json` and `frontend/src/lib/api/schemas.ts` both require
arrays and a `FactStatus` enum, so `parseDossier` threw on every real audit and
the UI showed "Pemeriksaan terhenti sebelum selesai" — a network message for a
contract bug.

**Frontend mitigation shipped:** Zod now accepts `null` and normalises to `[]`,
and `""` to `UNVERIFIED`; `parseDossier` throws a distinct `ContractError` so
the UI no longer blames the network. This is a safety net, not a fix.

**Owner: backend.** The real fix is to initialise these slices to empty values
before marshalling (or mark them `omitempty` and make the contract optional),
then tighten the Zod schema back to plain arrays. Until then the contract
document and the wire format disagree, and any client that trusts the schema
strictly will break.

Do not "fix" this by loosening the JSON Schema alone — that hides the drift
instead of removing it.

## 2026-11-01 · OPEN · Cross-origin `Retry-After` is unreadable until the API exposes it

`internal/httpapi/audit.go` sends `Retry-After` on a 429, and the frontend reads
it with `response.headers.get('Retry-After')`. Cross-origin, that header is
**not readable** unless the response carries `Access-Control-Expose-Headers:
Retry-After` — `security/headers.go` currently sets only `Allow-Origin`,
`Allow-Methods`, `Allow-Headers` and `Max-Age`.

Consequence: the countdown in "Terlalu banyak pemeriksaan. Coba lagi dalam N
detik." falls back to the default 30 seconds, which may be wrong.

**Frontend behaviour is correct either way** — it degrades to a sensible
default rather than showing nothing — so this is not a blocker.

**Owner: backend.** Add `Access-Control-Expose-Headers: Retry-After` to the
CORS middleware and a test asserting it, then tighten the share spec to assert
the exact countdown instead of only the retry button.

## 2026-11-02 · Multi-stage Dockerization and Nginx Reverse Proxy Architecture

Containerized full stack for cross-platform team usage (Linux x86_64, macOS Apple Silicon arm64, Windows WSL2) with fast rebuilds and zero-setup onboarding.

**Decisions made:**
1. **Multi-stage Dockerfiles:**
   - `backend/Dockerfile`: Multi-stage Go build (`base` -> `dev` with Air hot-reload -> `build` with `CGO_ENABLED=0` static binary -> `prod` on `alpine:3.21` with `ca-certificates`, `tzdata`, and non-root user `10001:10001`).
   - `frontend/Dockerfile`: Multi-stage Node build (`base` with `node:22-alpine` -> `dev` with Vite dev server -> `builder` with static SPA export -> `prod` on `nginx:1.27-alpine`).
2. **Relative API Base & Nginx Reverse Proxy:**
   - Production frontend static build injects `VITE_API_BASE_URL=/api/v1` (relative).
   - Nginx reverse-proxies `/api/` to `http://backend:8080/api/`.
   - Why: Keeps frontend and API on the same origin (`http://localhost` or LAN IP). Completely eliminates CORS configuration errors and CSP `connect-src` violations when team members test on local networks or mobile devices.
   - SSE streaming: Nginx location block explicitly disables buffering (`proxy_buffering off; proxy_cache off; chunked_transfer_encoding off;`) and extends read timeout to 600s so real-time audit streaming works without delay.
3. **Hot-Reload Dev Architecture (`docker-compose.dev.yml`):**
   - Backend mounts `./backend:/app` and runs Air. Air's temporary build directory is pinned to `/tmp/air` in container filesystem, preventing root-owned build artifacts from polluting the host repository.
   - Frontend mounts `./frontend:/app` with an anonymous volume for `/app/node_modules`. This prevents host OS node binaries (e.g., Darwin or Windows) from overriding container Linux binaries.
   - Sets `CHOKIDAR_USEPOLLING=true` for reliable file-change detection across WSL2 and macOS VirtioFS virtualized mounts.
4. **BuildKit Layer Caching for Fast Rebuilds:**
   - Dependency manifests (`go.mod`, `go.sum`, `package*.json`) are copied and installed with cache mounts (`/go/pkg/mod`, `/root/.cache/go-build`, `/root/.npm`) before application source code.
   - Code changes rebuild in 2-5 seconds without re-downloading packages from the network.
5. **Team Command Ergonomics:**
   - Provided root `Makefile` (`make up`, `make dev`, `make build`, `make logs`, `make down`).
   - Standardized on `docker-compose.yml` (demo/prod-like stack) and `docker-compose.dev.yml` (dev override).

Revisit: if deploying to Kubernetes or multi-node clusters, replace docker-compose with Helm charts / K8s manifests and add ingress TLS termination.

## 2026-11-05 · v10.1 is the contract; the downloaded v9 PRD is superseded

The file supplied for this realignment is PRD **v9.0.0**. `instruction/PRD_AMBA_v10.md`
says it supersedes v9.0.0 and v10.0, and it reverses several v9 choices
deliberately: Svelte 5/SvelteKit instead of React 19, a warm technical sticker
system instead of blue glassmorphism, server-authoritative ELA instead of a
client score, degraded real verdicts instead of a golden-cache fallback, and a
signed fragment token instead of `/hasil/[id]`.

Decision: keep the repo on **v10.1** and close the gap to v10.1. Reverting to
v9 would undo working, tested architecture. Revisit: only if the competition
rules are changed to require the v9 stack.

## 2026-11-05 · Colour is role-based, never ink/white pairs

`--ink` is the text colour and flips between themes. Using it as a background
produced near-white panels with dark text in dark mode, and pairing it with
`--lime`/`--yellow`/`--coral` produced 1.5–1.7:1 contrast.

Decision: every surface maps to one of four roles — `--on-fill` (pastel fills),
`--on-panel-dark` (deliberately dark panels), `--on-brand` (brand fills),
`--on-ink` (the few light ink fills). `tools/check-contrast.mjs` checks both
themes. Revisit: if a fifth surface family is introduced, add a role rather
than a one-off hex.

## 2026-11-05 · Demo case IDs differ from PRD F8

PRD F8 names `demo-tiktok`, `demo-berita-ig`, `demo-resmi`. The implementation
ships `contoh-hoax-bantuan`, `contoh-tautan-mencurigakan`, `contoh-informasi-aman`
(backend and frontend copies byte-identical, CI-diffed). The behaviour the
acceptance criteria describe is met: three synthetic cases, `is_demo: true`,
visible "CONTOH / SIMULASI" banner, zero network calls from the SPA, valid
against the Dossier schema.

Decision: keep the Indonesian IDs — they are the user-facing labels and the
demo is synthetic, so the v9 media-platform names would overclaim. Revisit: if
the jury deck or PRD must match literally, rename in one commit across
`backend/internal/demo/demo-cases.json`, `frontend/src/lib/data/demo-cases.json`,
and `backend/internal/httpapi/demo_test.go`.

## 2026-11-05 · Batch files are stored verbatim (`-text`), not normalised

Windows has no `make`, so `start.bat` / `start-dev.bat` / `stop.bat` mirror the
Makefile targets. They must reach a teammate as CRLF: `cmd.exe` expects it and
can mis-parse parenthesised `if errorlevel` blocks on bare LF.

`.gitattributes` therefore declares `*.bat -text`, i.e. **no conversion at
all**. `text eol=crlf` would store LF in the repository and convert only on
checkout, so any path that reads the blob without a checkout (an archive, a
copy from the web UI) would hand over LF. With `-text` the committed bytes are
the bytes every consumer receives.

`whitespace=cr-at-eol` is needed alongside it: without it `git diff --check`
reports every CR as trailing whitespace and the pre-commit gate fails.

Revisit: if the team standardises on a Windows-only toolchain, `eol=crlf`
would also work; `-text` stays correct either way.

## 2026-11-06 · Maru is the engine; the model never supplies a source

The gateway is 9router with the `Maru` combo. Four behaviours found by probing
it directly shaped the client, and each one is now covered by a test:

1. **9router streams by default.** A request without `stream:false` returns
   `text/event-stream`, which the old client could not parse. Every call now
   sends `stream:false` explicitly *and* decodes an SSE body if the gateway
   streams anyway.
2. **Maru ignores `response_format`.** It returns its own keys (`patterns`,
   `reasoning`, `skor`) and sometimes wraps JSON in a ```json fence. The
   decoder accepts both the requested schema and these aliases; only a missing
   numeric score fails the pillar. A strict unmarshal discarded a usable
   assessment on every image audit.
3. **Reasoning is always on.** A one-word answer costs ~5–14 s at default
   effort. `reasoning_effort:"low"` brings a realistic audit to 2–4 s, so it
   is the default and `HARD_TIMEOUT_MS` rose from 12 s to 20 s to leave image
   audits room. Images are downscaled to 1600 px for the model (`internal/imaging`);
   ELA still runs on the original bytes.
4. **`fact_status` was removed from the model's output.** The model cannot
   verify anything, and asking it for a verdict produced confident
   `CONTRADICTS` with an invented URL. Retrieval now owns the fact pillar:
   a fixed news endpoint (Bing News RSS) → allowlist filter → claim/number
   match → deterministic debunk/support markers. The model contributes only a
   search query and a claim.

A retrieved item must also share a **number** with the claim. Without that
gate, a fact-check about "bantuan Rp 900.000" was cited as contradicting a
claim about "Rp 500.000" — a false citation with a real, allowlisted URL.

## 2026-11-06 · Mock mode reports an honest UNVERIFIED fact pillar

Retrieval is skipped in mock mode (previews and e2e must stay offline). The
fact pillar is still `available:true` with `status:"UNVERIFIED"` and no
sources: the claim exists, nothing was checked. Marking the pillar unavailable
instead would have pushed text audits to `INSUFFICIENT` and broken the e2e
contract test for no product benefit.

## 2026-11-06 · Every unconfigured integration is a labelled placeholder

`GET /api/v1/ops/status` reports each dependency as live, disabled or
placeholder, with no key, host or model output. Safe Browsing without
`GOOGLE_SAFE_BROWSING_KEY` is a placeholder; retrieval without a live model is
a placeholder; mock mode labels the AI pillar a placeholder. The alternative —
silently skipping a check — is the dishonesty the trust contract forbids.

## 2026-11-06 · `/reply` drafts with the model, validates, then falls back

`internal/reply` asks Maru for a ≤ 280-character draft using only the verdict
summary already in the share token (never raw user content). The draft is
rejected wholesale if it contains certainty words or shaming, and every URL
outside `{token sources, share link}` is stripped. A rejected or unavailable
draft returns the deterministic status × tone template, and `source` says
which one the user received. The template matrix has a parity test against
`configs/reply_templates.json`.

## 2026-11-06 · Netlify for static hosting, not Vercel

The frontend is a static SPA (`adapter-static`); either host can serve it. The
decision is licence-based, not technical:

- **Netlify Free/Starter** permits commercial use, so the project does not need
  a paid plan if it later ships as a product.
- **Vercel Hobby** is restricted to non-commercial personal projects. Moving to
  Pro to satisfy that is a cost the project does not need.

The Go API is **not** deployed to either: it holds 20 s SSE streams and makes
outbound fetches, which do not fit serverless function limits. `netlify.toml`
therefore ships the SPA only, and the API base is configurable at build time
(`VITE_API_BASE_URL`). A preview-only deployment can build with
`VITE_API_MODE=mock`; it is honest because the app labels every mock result.

Two deploy bugs were fixed on the way: `connect-src` hardcoded
`http://localhost:8080` (a deployed build would block every request), and the
production API base defaulted to localhost instead of same-origin `/api/v1`.

## 2026-11-06 · Palette: three distinct hues, no pink

`--coral` was `#ff879c` (pink) and `--lime` was `#ffad87` (peach). The home
page also ran a `--copal-1/2/3` ramp that turned lime, amber and vermilion into
three shades of orange. Result: every accent collapsed into one peach family
and the danger state read as decoration.

Restored the spec's intent — a strong warm trio — while keeping the warm-orange
concept: `--lime #c8f751`, `--yellow #ffc53d`, `--coral #ff4b3e`, `--brand
#ff5a1f`. All 82 contrast pairs pass in both themes. Layout is untouched.

## 2026-11-06 · Home page: fabricated progress and badges removed

The "Pelajaran singkat" card showed made-up completion percentages (100/60/25/5)
for lessons that did not exist, and "Kumpulkan lencananya" advertised badges
with no storage. Both are gone: lessons are links to real `/edukasi` anchors,
and badges are real next-step links. "Buka panduan keluarga" now opens a new
`/panduan-keluarga` page with a client-only safe word and a printable card
(PRD F15), instead of pointing at the six-question self-check.

## 2026-11-06 · Safe accent: warm ivory instead of green

The electric lime (`#c8f751`) used for SAFE status, the hero highlight and
step 2 clashed with the copal orange. Replaced with warm ivory
(`--safe #efe0c9` dark / `#ecd9b8` light) and renamed the token from `--lime`
to `--safe`: a cream token called "lime" would be a lie, and the name now says
what it means.

Ivory was chosen over teal and cobalt because it keeps the entire palette warm
— no cool hue is left to clash with copal. Trade-off: SAFE reads calmer than a
saturated green, which is acceptable because status is never colour alone
(icon + label always ship).

Cream is a **surface, never text**: cream on `--brand` is 2.5:1. Three
hover/link rules that used it as text were moved to `--on-brand` / `--yellow`,
and `.btn--safe` was removed from the brand-filled control list.

## 2026-11-06 · Rebuilding one compose service recreates the other

Running `docker compose build frontend && docker compose up -d frontend` after
editing `docker-compose.yml` recreated `backend` with the file's defaults,
which are `LLM_MOCK=true` and an empty key. The app silently dropped to mock
mode. When restarting either service, pass the live env inline (or export it)
so the backend keeps `LLM_MOCK=false` and `NINEROUTER_API_KEY`:

```bash
LLM_MOCK=false NINEROUTER_API_KEY=... docker compose up -d --force-recreate backend
```

`GET /api/v1/ops/status` reports `llm.mode`, so this is detectable rather than
silent.

**Fix:** a gitignored root `.env` now holds the live values. Docker Compose
reads it automatically, so any `docker compose up` / `restart` keeps
`LLM_MOCK=false` and the key. `chmod 600`; it is never committed
(`.gitignore` line 2).
