# Honesty map — labels, notes and anti-patterns

Load when implementing or reviewing the honest-limits panel, degraded banner, ELA reliability note, floor explanation, or the status/score display.

Sources: TECH_SPEC_AMBA.md §3.3, §8.0-§8.2, §9; PRD_AMBA_v10.md §4, F5, F6, F8.

## 1. `verdict.basis` → "Berdasarkan:" labels

| basis value | Indonesian label |
|---|---|
| `ela` | "Analisis gambar (ELA)" |
| `fact` | "Pengecekan fakta" |
| `ai` | "Deteksi jejak AI" |
| `link` | "Pemeriksaan tautan" |
| `llm_untrusted` | "Catatan: teks dari AI tidak dipakai sebagai bukti" (not a pillar; appears when `INJECTION_GUARD` fired) |

Render as: `Berdasarkan: Analisis gambar (ELA), Pengecekan fakta.`

## 2. `verdict.not_checked` → "Tidak dapat diperiksa:" labels

| not_checked value | Indonesian label |
|---|---|
| `ela` | "Analisis gambar (ELA)" |
| `fact` | "Pengecekan fakta" |
| `ai` | "Deteksi jejak AI" |
| `link` | "Pemeriksaan tautan" |

Render as: `Tidak dapat diperiksa: Deteksi jejak AI.` Never append a reason the backend did not send; the pillar payload's `available:false` is the reason.

## 3. `verdict.floors_applied` → short explanation

Floors are why a score can rise above the weighted average. Show them only when present; the user does not need the code name.

| floor code | plain sentence (usulan copy) |
|---|---|
| `FACT_CONTRADICTED` | "Kabar ini dibantah sumber terpercaya, jadi skor risiko dinaikkan." |
| `LINK_MALICIOUS` | "Tautan ini cocok dengan pola penipuan yang diketahui, jadi skor risiko dinaikkan." |
| `ELA_CORROBORATED` | "Gambar terindikasi diedit dan temuan ini saling menguatkan, jadi skor risiko dinaikkan." |
| `ELA_SOLO_CAP` | "Hanya analisis gambar yang bisa dipakai, jadi skor risiko dibatasi." |
| `INJECTION_GUARD` | "Teks dari AI tidak dipakai sebagai bukti, jadi skor risiko punya batas bawah." |

Floor codes themselves are internal; never print `FACT_CONTRADICTED` in the UI.

## 4. ELA reliability note

| `pillars.ela.reliability` | `available` | UI |
|---|---|---|
| 1.0 | true | score + anomalies, no note |
| 0.5 | true | score + note: "Gambar ini jenisnya membuat analisis kurang meyakinkan." |
| 0.0 | false | "Tidak dapat diperiksa" — no score, no `0` |

Use `reliability_note` from the Dossier when it is non-null; otherwise the sentence above.

## 5. State-specific copy

| state | lead line | extra |
|---|---|---|
| SAFE | "Tidak ditemukan tanda bahaya, tapi tetap bijak sebelum menyebarkan." | never "pasti asli" |
| SUSPICIOUS | "Belum jelas benar atau tidak. Cek dulu sebelum menyebarkan." | (from `explanation_simple`) |
| TAMPERED | floor-specific sentence from §8.2 | label "Terindikasi Hoax/Manipulasi" |
| INSUFFICIENT | "Informasinya belum cukup untuk dipastikan." | next step, not retry-only |
| degraded:true | usulan "Sebagian pemeriksaan tidak berjalan." | list `not_checked` |
| is_demo:true | "Ini data contoh, bukan hasil audit sungguhan." | watermark "CONTOH / SIMULASI" |
| invalid token | "Laporan tidak sah." | no payload |

`explanation_simple` templates are fixed in Go (TECH_SPEC §8.2); the frontend renders them verbatim and never rewrites them.

## 6. Anti-patterns that destroy trust

| Anti-pattern | Why it breaks trust | Instead |
|---|---|---|
| Huge red number, no sentence | reads as a verdict of guilt / a probability | score + label + icon + `explanation_simple` |
| `92%` or "62% keyakinan" | score is risk, not confidence | "Skor risiko 92 dari 100" |
| ELA `0` shown as "bersih" | reliability 0.0 means not measurable | "Tidak dapat diperiksa" |
| Hidden `degraded` flag | presents a partial run as complete | banner + list of skipped checks |
| INSUFFICIENT styled as an error/red card | punishes an honest answer | neutral card + next step |
| Confetti / "Aman!" / lone green tick on SAFE | overclaims verification | "Tidak ditemukan tanda bahaya" |
| "Bukti hukum" / "bisa dipakai di pengadilan" | AMBA is not legal evidence | "INDIKATIF — bukan alat bukti hukum" |
| Urgency/fear copy ("bahaya!", "segera sebarkan") | mirrors the tactics being detected | calm, factual tone |
| Demo shown without watermark | demo masquerades as a live result | "CONTOH / SIMULASI" + banner |
| Invalid token rendering "0" or empty sources | leaks the shape of the report | "Laporan tidak sah" only |
| Source list without publisher or with invented URL | fabricates authority | publisher + real URL; demo uses "Sumber Contoh" |
| Tooltip-only caveat on AI pillar | caveat becomes optional | visible line "Indikatif, bukan bukti pasti" |
| Colour-only status | inaccessible | icon + text label always |
| Score in Explore mode but not Essential | breaks mode parity | same `data-info-key="verdict.score"` in both |
