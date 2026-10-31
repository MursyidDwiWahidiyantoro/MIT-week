---
name: amba-sse-ux
description: Builds AMBA's streaming and progressive UI from real server events only. Covers the SSE stage -> partial -> verdict -> done contract, the 1.5 s / 3.5 s / 8 s / 12 s latency budget, per-pillar sub-deadline degradation with layout-stable skeletons, aria-live coalescing and verdict focus, AbortController cleanup, and the 429 / offline / degraded / 504-timeout / unknown-schema failure states. Trigger when building or reviewing /deteksi scanning feedback (ScanningRadar, progress list), /hasil progressive reveal, AuditSession SSE code, stage announcements, skeleton loaders, cancellation on route leave, or when asked "why does the radar keep spinning", "show progress without faking it", "streaming results", "progres pemindaian", "kenapa loading lama", "skeleton", "aria-live tahap".
---

# AMBA SSE & Progressive UX

Progress shown to the user must be **measured**, never simulated. The server emits events; the UI renders them. Nothing else.

## 1. Event contract (TECH_SPEC §2)

Endpoint: `POST /audit/{url|file|text}` with `Accept: text/event-stream`. Use `fetch` + `ReadableStream`; **not** `EventSource` (it cannot POST or carry an abort signal).

| Event | Payload | UI effect |
|---|---|---|
| `stage` | `{"pillar":"ela\|ai\|fact\|link","status":"running\|done\|skipped\|failed"}` | set that pillar's row to running / done / skipped / failed |
| `partial` | `{"pillar":"fact","data":{…pillar payload…}}` | render that pillar's real content, replacing its skeleton |
| `verdict` | full `Dossier` (signed `share_token` included) | render verdict, move focus, stop the scanner |
| `error` | error envelope | show the mapped failure state; keep input |
| `done` | `{}` | close the stream, stop timers, mark session complete |

Rules:
- Order is `stage → partial → verdict → done`. Per pillar, its `stage:running` then `stage:done|skipped|failed` precede its `partial`.
- `: ping` comment every 10 s is a heartbeat, not a stage. Never surface it.
- The radar animation, the progress list, and every percentage come **only** from `stage`/`partial` counts. Four pillars → "2 dari 4 bagian selesai", not "47%".
- Demo (`/demo/{id}`) renders a complete bundled Dossier; it does not fake SSE.
- `verdict` may arrive with `degraded:true` — that is a real result, not an error.

## 2. Latency budget (PRD §5)

| Moment | Budget | UI consequence |
|---|---|---|
| first `partial` | ≤ 1.5 s p50 | by 1.5 s show the first pillar's real content; never leave a spinner-only screen when a partial exists |
| final `verdict` | ≤ 3.5 s p50, ≤ 8 s p95 | no "almost done" copy, no invented ETA |
| hard timeout | 12 s → `degraded:true` verdict **or** `504 UPSTREAM_TIMEOUT` | at 12 s the stream must reach a terminal state, not hang |

- Show named stages ("Memeriksa tautan…") immediately; a named step is honest, a percentage is not.
- Past 8 s without a verdict: keep the stage list visible and show "Masih memeriksa. Kamu boleh tetap di halaman ini." Do **not** add an estimated time.
- Past 12 s with no `verdict`: show the terminal state with a retry, not an endless radar. Two distinct outcomes, never merged:
  - server returned a 200 Dossier with `degraded:true` -> render that verdict plus the skipped-pillar list;
  - server returned `504 UPSTREAM_TIMEOUT` (error envelope, **no** verdict) -> show the retry state only. Do not invent a verdict.

## 3. Per-pillar degradation

Pillars resolve independently (errgroup, sub-deadlines: ELA 0.8 s file-only, link 2.5 s, LLM 9 s). Each pillar has three UI states:

| Pillar state | Render |
|---|---|
| pending | skeleton in the **final** box |
| resolved | real pillar content |
| unavailable | "Tidak dapat diperiksa" — only for pillars listed in `verdict.not_checked`; never a `0` score |

- One slow pillar never blocks the page. Verdict + resolved pillars render as they arrive.
- **Skeleton box parity:** the skeleton occupies the same width, height and margins as the final pillar block (reserve space with `min-height` / `aspect-ratio`). No layout shift when content lands (PRD: CLS 0).
- A `partial` replaces only its own pillar. Do not re-render or reset pillars already shown.
- Never show a partial pillar's number as if it were the final score. Pillar scores are labeled by pillar; the overall verdict only exists after `verdict`.

## 4. Accessibility

- Stage changes go through one **polite** `aria-live` region (`aria-atomic="false"`). Coalesce: announce at most once per ~500 ms, and only when the stage text actually changes. Never announce every frame, tick, or heartbeat.
- The live region text is the current step, e.g. "Memeriksa tautan." then "Pengecekan fakta selesai."
- On `verdict`, move focus to the verdict heading (`<h2 tabindex="-1">`) **only if** focus is still on the submit control or body; never steal focus from a field the user is typing in.
- The result must be fully keyboard reachable: scanner is non-interactive, verdict and pillars are in DOM order.
- Calm motion: `html[data-motion="calm"]` is forced when `prefers-reduced-motion: reduce` **and** always in Mode Esensial. In calm motion, `ScanningRadar` is replaced by a **static labelled progress list** (`stage` rows, no animation). Explore Mode may animate; it must stop on `verdict`/`done` and never loop after completion.
- Status is never colour-only: each row has icon + text label.

## 5. Cancel, abort, unmount

- `AuditSession` (`$lib/audit.svelte.ts`, runes) owns one `AbortController` per audit, created on start and aborted in `$effect` cleanup.
- On abort: cancel the reader, drop the reference, mark the session cancelled. **No state write after unmount** — guard every handler with a `cancelled`/`disposed` flag.
- One retry on network error (not on 4xx/429/504), then a friendly failure with a "Coba lagi" action. Never a retry loop.
- Route leave during a scan aborts and clears timers; returning does not resume a zombie stream.
- Keep the user's typed input / chosen file in `AuditSession` across the abort so retry needs no re-typing.

## 6. Failure states

Each keeps the screen non-blank and preserves typed input. Full matrix: `references/stream-states.md`.

| Trigger | What stays on screen | Copy |
|---|---|---|
| `429` + `Retry-After` | input + countdown, submit disabled until 0 | "Terlalu banyak pemeriksaan. Coba lagi dalam {n} detik." |
| offline | input + demo cases + retry | "Kamu sedang offline. Contoh kasus masih bisa dibuka." |
| `degraded:true` (HTTP 200) | verdict + resolved pillars + skipped list | "Sebagian pemeriksaan tidak berjalan." |
| `504 UPSTREAM_TIMEOUT` | input + retry (no verdict exists) | "Pemeriksaan belum selesai. Coba lagi." |
| `schema_version` unknown | one line, no payload | "Perbarui aplikasi" |
| reload of live result | nothing persisted → recovery CTA | "Hasil tidak tersimpan — periksa lagi" |
| stream ends without `verdict` | stage list + retry | "Pemeriksaan terhenti sebelum selesai. Coba lagi." |

## 7. Review checklist

- [ ] Radar/progress derives from `stage`/`partial` only; `grep -rn "setInterval\|setTimeout" src/lib/components/Scanning* src/lib/audit*` returns nothing that feeds progress.
- [ ] No percentage anywhere that was not computed from resolved-pillar count.
- [ ] Event order asserted in a test: `stage → partial → verdict → done`, and `done` last.
- [ ] Abort on route leave asserted; no post-unmount state write.
- [ ] `aria-live` region is polite, single, and coalesced; a test asserts N rapid stages produce fewer announcements.
- [ ] Skeleton and final pillar have identical box dimensions (no CLS).
- [ ] `prefers-reduced-motion` and Mode Esensial both render the static list, not the radar.
- [ ] Focus lands on the verdict heading on arrival, and not while the user is typing.
- [ ] Every failure state leaves the input and offers one next action.
- [ ] Works in both presentation modes with identical information keys (`data-info-key`).

## Do not

- Do not fake progress: no timer-driven radar, no synthetic bar, no "estimasi 5 detik".
- Do not use an indeterminate spinner as the only feedback past the first stage event.
- Do not animate a percentage you did not measure.
- Do not treat `partial` as final or show pillar scores as the verdict.
- Do not block the whole page on the slowest pillar.
- Do not announce every SSE frame or heartbeat to assistive tech.
- Do not let motion continue under `prefers-reduced-motion` or in Mode Esensial.
- Do not use `EventSource` (no POST, no abort) or leave a stream open after route change.
- Do not blank the page, clear the input, or auto-retry in a loop on failure.
- Do not show raw error codes (`UPSTREAM_TIMEOUT`, `BLOCKED_TARGET`) or stack text.
- Do not add an estimated time, a fake "AI sedang berpikir" delay, or a minimum spinner duration to look busy.
