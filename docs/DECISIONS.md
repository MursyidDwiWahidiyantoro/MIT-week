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
