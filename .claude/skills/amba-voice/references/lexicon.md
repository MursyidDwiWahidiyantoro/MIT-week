# Lexicon — AMBA word and template tables

Load when choosing any verdict/status word, translating jargon, or quoting a spec string.

## 1. Allowed vs forbidden

| Word | Status | Why | Replacement |
|---|---|---|---|
| terindikasi | allowed | hedged finding | — |
| indikatif | allowed | signals only suggest | — |
| tidak ditemukan tanda bahaya | allowed | absence of evidence, not proof | — |
| belum cukup data | allowed | too little evidence (INSUFFICIENT) | — |
| bukan bukti pasti | allowed | honest-limits caveat | — |
| cek dulu sebelum menyebarkan | allowed | recommended action | — |
| palsu | forbidden | verdict-like certainty; defamation risk | "terindikasi tidak sesuai", "belum terverifikasi" |
| aman | forbidden as verdict | implies guarantee; "no evidence of risk" ≠ verified safe | "tidak ditemukan tanda bahaya" |
| pasti | forbidden | certainty | "terindikasi", "kemungkinan" |
| terbukti | forbidden | legal/evidentiary claim; AMBA is not legal evidence (disclaimer) | "terindikasi" |
| 100% | forbidden | certainty | "indikatif" |
| hoax | forbidden as a certainty claim | the word is a conclusion | only the TAMPERED label "Terindikasi Hoax/Manipulasi" |
| akurat | forbidden | claims accuracy AMBA cannot guarantee | "indikatif" |
| dijamin | forbidden | guarantee | remove |

## 2. Status labels (TECH_SPEC §3.3, PRD §4)

| Status | Score | Exact label | Note |
|---|---|---|---|
| SAFE | 0–20 | "Tidak ditemukan tanda bahaya" | never "pasti asli"; requires W ≥ 0.50 and a Fact or ELA+Link pillar |
| SUSPICIOUS | 21–60 | none defined | do not invent; lead with `explanation_simple`; log any new label in `docs/DECISIONS.md` |
| TAMPERED | 61–100 | "Terindikasi Hoax/Manipulasi" | only place "Hoax" may appear |
| INSUFFICIENT | — | "Belum cukup data" | grey; never "SAFE" |

## 3. explanation_simple templates (TECH_SPEC §8.2) — verbatim

| Trigger | Exact string |
|---|---|
| `FACT_CONTRADICTED` | "Kabar ini dibantah sumber terpercaya. Jangan disebarkan." |
| `LINK_MALICIOUS` | "Tautan ini mirip penipuan. Jangan diklik." |
| `ELA_CORROBORATED` | "Gambar ini terindikasi diedit. Jangan langsung percaya." |
| other `SUSPICIOUS` | "Belum jelas benar atau tidak. Cek dulu sebelum menyebarkan." |
| `INSUFFICIENT` | "Informasinya belum cukup untuk dipastikan." |
| `SAFE` | "Tidak ditemukan tanda bahaya, tapi tetap bijak sebelum menyebarkan." |

Generated deterministically in Go. ≤ 140 chars. Never LLM-written.

## 4. Jargon translation

| Term | Plain Indonesian (one line) |
|---|---|
| DCT | "Cara gambar disimpan dalam blok kecil; perbedaan blok bisa jadi tanda edit." |
| ELA | "Pemeriksaan tingkat kompresi gambar untuk melihat bagian yang mungkin diedit." |
| OCR | "Membaca teks yang ada di dalam gambar." |
| LLM | "Model bahasa AI yang membantu membaca isi konten." |
| SSRF | never user-facing; internal security term |
| hash | "Kode unik dari isi file, dipakai tanpa menyimpan isinya." |
| metadata | "Informasi tambahan di file, seperti waktu atau alat pembuatnya." |
| bounding box | "Kotak penanda bagian gambar yang dicurigai." |

Rule: jargon only with the one-line explanation on the same screen; never in Mode Esensial labels.

## 5. Fixed tactic labels (TECH_SPEC §6.1)

Labels and "kenapa mencurigakan" lines live in static `frontend/src/lib/data/tactics.json`; do not retype from LLM output.

| Code | Label |
|---|---|
| URGENCY | Mendesak |
| REWARD_BAIT | Iming-iming hadiah |
| FALSE_AUTHORITY | Mengaku otoritas |
| FEAR | Menakut-nakuti |
| SECRECY | Minta dirahasiakan |
| OUTRAGE_BAIT | Memancing emosi |
| IMPERSONATION | Menyamar sebagai orang/lembaga |
| FAKE_SOCIAL_PROOF | Klaim "sudah dipercaya banyak orang" |
