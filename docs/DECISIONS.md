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
