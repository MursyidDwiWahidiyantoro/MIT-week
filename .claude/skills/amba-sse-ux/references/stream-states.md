# Stream states — event→UI table and failure matrix

Load when wiring the SSE client, mapping `stage`/`partial`/`verdict`/`error` to UI, or implementing any failure state on /deteksi or /hasil.

Sources: TECH_SPEC_AMBA.md §2, §4, §8, §8.0-§8.2; PRD_AMBA_v10.md §5, F1-F5, F9, F14; DESIGN.md §5, §6.

## 1. Event → UI state

| # | Event received | State transition | Screen |
|---|---|---|---|
| 0 | submit accepted | `idle → scanning` | input collapses to a summary; stage list appears with all 4 pillars pending |
| 1 | `stage` `running` (pillar P) | `P: pending → running` | row P: icon + "Memeriksa …"; announce if text changed and coalesce window passed |
| 2 | `stage` `done` (P) | `P: running → done` | row P: check + "selesai"; do **not** reveal content yet |
| 3 | `stage` `skipped` / `failed` (P) | `P → unavailable` | row P: "Tidak dapat diperiksa"; keep the box (skeleton parity) |
| 4 | `partial` (P) | `P: done → resolved` | replace P's skeleton with real pillar content; keep other pillars untouched |
| 5 | `verdict` | `scanning → complete` | render Dossier; stop radar; focus verdict heading (guarded); start calm/static state |
| 6 | `done` | `complete → closed` | close reader, clear timers, release abort controller |
| 7 | `error` | `scanning → failed` | mapped failure state (§2); input preserved |
| 8 | stream closes without `verdict` | `scanning → failed` | "Pemeriksaan terhenti sebelum selesai. Coba lagi." |
| 9 | abort (route leave / Batalkan) | `scanning → cancelled` | no UI write after unmount; session cleared |

Pillar display names (from the honesty map, keep identical everywhere):

| pillar | label |
|---|---|
| `ela` | "Analisis gambar (ELA)" |
| `ai` | "Deteksi jejak AI" |
| `fact` | "Pengecekan fakta" |
| `link` | "Pemeriksaan tautan" |

Derived progress: `resolved_or_unavailable / 4` of pillars. Text only: "2 dari 4 bagian selesai". Never a percentage.

## 2. Failure-state matrix

| Trigger | HTTP / condition | What stays on screen | Primary action | Indonesian copy (usulan unless quoted) |
|---|---|---|---|---|
| Rate limited | `429` + `Retry-After` | input, disabled submit, live countdown | auto re-enable at 0 | "Terlalu banyak pemeriksaan. Coba lagi dalam {n} detik." |
| Offline | `navigator.onLine === false` or fetch TypeError | input, demo cases, retry | "Coba lagi" | "Kamu sedang offline. Contoh kasus masih bisa dibuka. Periksa lagi saat koneksi kembali." |
| Partial failure | `degraded:true` (HTTP 200 Dossier) | verdict + resolved pillars + skipped list | view result | "Sebagian pemeriksaan tidak berjalan." + "Hasil ini tetap berguna, tetapi belum lengkap." |
| Upstream timeout | `504 UPSTREAM_TIMEOUT` (error envelope, no Dossier) | input + retry; **no verdict to render** | "Coba lagi" | "Pemeriksaan belum selesai. Coba lagi." |
| Hard deadline hit | 12 s, no `verdict` | stage list, no spinner-only | "Coba lagi" | "Pemeriksaan terhenti sebelum selesai. Coba lagi." |
| Bad input | `400 INVALID_URL` / `415` / `413` | form + inline field error | fix field | use server `message`; never the code |
| Blocked target | `400 BLOCKED_TARGET` | form + generic note | edit input | generic message only — never which rule fired |
| Unknown schema | `schema_version !== 2` | one line, no partial payload | update app | "Perbarui aplikasi" |
| Reloaded live result | no `AuditSession`, no `#t=` | empty state | "Periksa lagi" | "Hasil tidak tersimpan — periksa lagi" |
| Invalid share token | `/verify` `valid:false` | nothing else | "Periksa ulang kontennya" | "Laporan tidak sah" |
| LLM down, rest fine | `degraded:true`, `not_checked:["ai","fact"]` | full verdict minus AI and Fact pillars — the LLM pillar group (ai, fact, claim, anomalies) fails together | — | "Deteksi jejak AI" and "Pengecekan fakta" listed under "Tidak dapat diperiksa:" |
| Share-target entry | PWA share hand-off | auto-started scan + "Batalkan" | cancel aborts | — |

Invariants for every row:
- Screen is never blank; the user's typed text / chosen file survives.
- Exactly one next action, or a plain reason why none is possible.
- No HTTP code, no error code, no stack trace, no English.
- Failure copy never blames the user and never claims certainty about the content.

## 3. Timing rules

| Time | Must be true |
|---|---|
| 0 ms | a named stage is visible (no empty spinner) |
| ≤ 1.5 s | first real pillar content is on screen if the server sent `partial` |
| ≤ 8 s (p95) | verdict normally present |
| 12 s | resolve to degraded verdict or a terminal failure state |

- Heartbeat `: ping` every 10 s: reset any client idle detector, render nothing.
- No client-side timer may **create** progress. Timers may only: re-enable a 429 button, expire a countdown, or close a stalled stream at the hard deadline.

## 4. Announcement rules

| Announce | Do not announce |
|---|---|
| stage text change (coalesced ~500 ms) | heartbeats, repeated identical stages |
| pillar becomes unavailable | each partial payload field |
| verdict ready ("Hasil sudah siap") | animation frames, countdown every second |

One polite live region for the whole scan. The verdict itself is announced by focusing its heading, not by duplicating the full text in the live region.
