---
name: amba-trust-ux
description: Enforces AMBA's calibrated-uncertainty and honest-limits rules on every result surface. Covers the 0-100 RISK score (never confidence), status bands, the "Berdasarkan / Tidak dapat diperiksa" panel, ELA reliability notes, degraded and INSUFFICIENT states, demo watermarks, invalid share tokens, and the SAFE asymmetry. Trigger when writing or reviewing any verdict UI, score display, /hasil page, proof card, share or verify screen, degraded/error copy, demo badge, or when asked "is this score confident", "how do we show uncertainty", "what does degraded mean", "why is score null", "make the verdict honest", "tampilkan skor", "panel batas", "hasil tidak sah", "skor keyakinan".
---

# AMBA Trust UX

Every result surface states what AMBA knows **and what it does not**. The verdict is evidence, not proof.

## 1. Score is RISK, not confidence

`verdict.score` is 0-100 **risk** (higher = more risky). It is not a probability, not a confidence, not a percentage.

| score | status | label to show | tone token | icon |
|---|---|---|---|---|
| 0-20 | SAFE | "Tidak ditemukan tanda bahaya" | `--ok` | shield-check |
| 21-60 | SUSPICIOUS | usulan: "Perlu diperiksa" | `--warn` | alert-triangle |
| 61-100 | TAMPERED | "Terindikasi Hoax/Manipulasi" | `--err` | alert-octagon |
| null | INSUFFICIENT | "Belum cukup data" | neutral | help-circle |

Rules:
- Always render score + status label + icon + `explanation_simple` **together**. Never a bare number, never a number in a card without the sentence.
- Always write the unit: "Skor risiko 92 dari 100". Never `92%`, `92/100 keyakinan`, `peluang 92%`.
- Never the words `%`, `keyakinan`, `probabilitas`, `peluang`, `yakin`, `pasti`, `terbukti` near the score.
- Score stays visible in both presentation modes (`data-info-key="verdict.score"`); Esensial may reduce its size, never remove it.
- `explanation_simple` is the lead sentence in every mode and the first thing read aloud.

## 2. Honest-limits panel (mandatory under the verdict)

Render two lines from the Dossier, plus a note when relevant:

- `Berdasarkan: …` from `verdict.basis`
- `Tidak dapat diperiksa: …` from `verdict.not_checked`
- ELA reliability note when `pillars.ela.reliability < 1`
- Floor explanation when `verdict.floors_applied` is non-empty (why the score rose or was capped)

Empty `basis` → "Belum ada pemeriksaan yang berhasil." Never hide the panel because it is empty.

Label mapping for every `basis`, `not_checked`, reliability and floor code: load `references/honesty-map.md` when implementing or reviewing this panel. Never print a raw enum code to the user.

## 3. Reliability semantics (ELA)

| reliability | meaning | what the UI says |
|---|---|---|
| 1.0 | JPEG-origin, analysis trustworthy | show ELA score, no note |
| 0.5 | PNG/WebP-lossless or screenshot-like | show score **plus** "Gambar ini jenisnya membuat analisis kurang meyakinkan." |
| 0.0 | image < 64 px | pillar is unavailable: `available:false`, say "Tidak dapat diperiksa", show **no** number |

Never show a fake `0` ELA score. "Not measurable" is not "clean".

## 4. DEGRADED state

`degraded:true` means one or more pillars failed (timeout, LLM down, site unreachable).

- Banner above the verdict: usulan "Sebagian pemeriksaan tidak berjalan".
- List what did not run, from `not_checked` (same labels as the honest-limits panel).
- Keep the verdict and the action; add usulan "Hasil ini tetap berguna, tetapi belum lengkap."
- Never say "gagal" in a way that blames the user; never present a degraded result as complete; never hide the flag to look polished.

## 5. INSUFFICIENT state

`W < 0.30` → `status:"INSUFFICIENT"`, `score:null`. This is an **honest answer, not an error**.

- Show "Belum cukup data" + the exact sentence "Informasinya belum cukup untuk dipastikan."
- Never render `0`, never a spinner, never a red error card, never "coba lagi" as the only action.
- Give a next step (usulan): "Coba sertakan tautan aslinya, tempel teks lengkapnya, atau unggah tangkapan layar yang lebih jelas."

## 6. Demo watermark

`is_demo:true` → watermark "CONTOH / SIMULASI" always visible (also on the proof card, share preview and small screens), plus banner "Ini data contoh, bukan hasil audit sungguhan".

- Demo never shows the green "Terverifikasi AMBA" mark and never masquerades as a live verdict.
- Demo share/verify screens keep the watermark.

## 7. Invalid / tampered share token

`/verify` returns `valid:false` (bad signature, truncated, rotated key) → show "Laporan tidak sah".

- Render **no** partial payload: no score placeholder, no source list, no "0".
- No retry loop that hints at tampering; offer only "Periksa ulang kontennya".

## 8. SAFE asymmetry

SAFE requires `W >= 0.50` **and** an available Fact pillar or ELA+Link pillars. Otherwise the verdict downgrades to INSUFFICIENT.

Copy consequence: "Tidak ditemukan tanda bahaya" — absence of evidence is not verification. Never "pasti asli", "terverifikasi aman", "100% aman", "verified safe".

## 9. Review checklist

| Surface | Must be visible | Must never be claimed |
|---|---|---|
| Overall verdict | score (int), status label, icon, `explanation_simple`, honest-limits panel | certainty, probability, "terbukti" |
| ELA pillar | score, reliability note when < 1, anomalies list, "Tidak dapat diperiksa" when `available:false` | clean when reliability 0.0 |
| AI pillar | score, caveat "Indikatif, bukan bukti pasti" | proof that content is AI-made |
| Fact pillar | claim, status, sources with publisher, placeholder label "Sumber Contoh" for demo | invented sources or URLs |
| Link pillar | `S_Link`, final URL, redirect chain, reasons | legal proof of fraud |
| Degraded | banner + list of skipped checks | complete analysis |
| Demo | "CONTOH / SIMULASI" watermark + banner | live result, "Terverifikasi AMBA" |
| Invalid token | "Laporan tidak sah" only | any recovered data |

Full anti-pattern table and exact Indonesian strings: `references/honesty-map.md`.

## Do not

- Do not show `score%` or any percentage, confidence or probability next to the score.
- Do not render a number without its status label, icon and sentence.
- Do not show ELA `0` when reliability is `0.0` or the pillar is unavailable.
- Do not hide `degraded`, `not_checked` or `is_demo` to make the screen look finished.
- Do not celebrate SAFE (confetti, "Aman!", green checkmark alone) — absence of evidence is not verification.
- Do not write "pasti", "terbukti", "dijamin", "100% aman", "verified safe", "bukti hukum".
- Do not treat INSUFFICIENT as an error, a retry prompt or a failure of the user.
- Do not render any payload when a share token is invalid.
- Do not use urgency or fear copy ("segera sebarkan", "bahaya!"); AMBA must not mirror the tactics it detects.
- Do not invent sources, publishers or URLs; demo uses "Sumber Contoh" and `example.org` placeholders only.
- Do not show one-decimal scores; the score is an integer 0-100.
