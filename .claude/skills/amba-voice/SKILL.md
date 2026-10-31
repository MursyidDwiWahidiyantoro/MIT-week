---
name: amba-voice
description: Enforces AMBA's Indonesian plain-language voice and vocabulary contract for every user-facing string (buttons, errors, empty/loading states, 429, offline, degraded verdicts, verdict sentences, tactic explanations, Balas Sopan drafts, Edukasi and Tentang copy). Use when writing, reviewing or translating UI copy for AMBA, when a string must not overclaim, when choosing a status label, when naming a technical concept (ELA, OCR, hash, metadata), or when asked "apakah copy ini aman/benar". Triggers include "tulis copy", "perbaiki teks UI", "pesan error", "empty state", "label status", "balas sopan", "copy Indonesia", "cek nada", "vocabulary AMBA".
---

# AMBA Voice

Contract for all user-facing text in AMBA. Sources: `instruction/TECH_SPEC_AMBA.md` (§3.3, §6.1, §6.2, §8, §8.2), `instruction/PRD_AMBA_v10.md` (§3, §4, §10), `instruction/DESIGN.md`, `instruction/CLAUDE.md` (rules 5, 6, 7).

AMBA is a **literacy aid**, not a lie detector. The number is a **risk score 0–100** (higher = more risky), never a probability or confidence. Never write or imply "62% keyakinan".

## 1. Vocabulary contract

Full table with reasons and replacements: `references/lexicon.md` (load when choosing any verdict/status word).

| Allowed | Forbidden (overclaim) | Use instead |
|---|---|---|
| terindikasi | palsu | terindikasi tidak sesuai / perlu dicek |
| tidak ditemukan tanda bahaya | aman (as verdict) | tidak ditemukan tanda bahaya |
| belum cukup data | pasti / terbukti | belum cukup data / indikatif |
| indikatif, bukan bukti pasti | 100% / akurat / dijamin | indikatif |
| cek dulu sebelum menyebarkan | hoax (as certainty) | "Terindikasi Hoax/Manipulasi" (only the TAMPERED label) |

Rule: a forbidden word may appear **only** inside an exact spec string quoted in `references/lexicon.md`, never in new copy.

## 2. Status labels (TECH_SPEC §3.3, PRD §4)

| Status | Score | Label (exact) |
|---|---|---|
| SAFE | 0–20 | "Tidak ditemukan tanda bahaya" — never "pasti asli" |
| SUSPICIOUS | 21–60 | no label string defined in the spec; lead with `explanation_simple`. Do not invent one — record any new label in `docs/DECISIONS.md` |
| TAMPERED | 61–100 | "Terindikasi Hoax/Manipulasi" |
| INSUFFICIENT | — | "Belum cukup data" |

Every status is shown with icon + text label, never colour alone.

## 3. explanation_simple

Copy the six templates **verbatim** from `references/lexicon.md` §3 (they are generated in Go, not by the LLM). Never paraphrase them in UI code; never let the LLM write this sentence. It is the lead sentence of every result, in both modes.

## 4. Jargon

Jargon is allowed only with a one-line plain-Indonesian explanation, and **never in Mode Esensial labels**. Translations: `references/lexicon.md` §4.

## 5. Register

- Imperative or neutral second person ("Periksa tautan ini"), no honorific stacking.
- Short sentences, aim ≤ 20 words; one idea per sentence.
- No bureaucratic passive: avoid "dilakukan", "merupakan", "dipergunakan". Say the actor: "AMBA memeriksa…".
- Indonesian, not translated English ("unduh" not "download", "tautan" not "link" in prose; "tautan" may appear in UI labels).
- Warm but not chatty; never blame the user ("Anda salah mengunggah" → "Format file ini belum didukung").
- No emoji, no exclamation marks, no marketing superlatives.

## 6. Balas Sopan tones (TECH_SPEC §6.2)

| Tone | Voice | Example draft (≤ 280 chars) |
|---|---|---|
| hormat | formal, menyapa sopan, tanpa menyalahkan | "Mohon izin berbagi, Bapak/Ibu. Informasi ini belum terverifikasi. Saya cek lewat AMBA dan hasilnya indikatif: belum ditemukan sumber resmi. Mungkin kita cek dulu sebelum meneruskan." |
| santai | akrab, singkat, tetap sopan | "Eh, info ini kayaknya perlu dicek dulu. Aku coba cek pakai AMBA, hasilnya indikatif dan belum ada sumber resminya. Kita tahan dulu ya sebelum share." |
| singkat | satu baris, inti saja | "Info ini belum terverifikasi. Sebaiknya cek dulu sebelum diteruskan." |

Rules for all tones: talks about the content, never the sender; ≤ 280 chars; no certainty words; no invented facts; URLs only from verified sources or the AMBA share link (the server strips others). Drafts above are examples, not spec strings — the shipped text comes from `configs/reply_templates.json`.

## 7. State copy

Rules + one example per state (offline, 429 countdown, degraded, unsupported file, no checkable claim, platform-blocked, invalid token): `references/states.md` (load when writing any non-verdict state). Exact spec strings that must not drift:

- "Platform tidak mengizinkan pembacaan otomatis" + CTA "Unggah tangkapan layar"
- "Hasil tidak tersimpan — periksa lagi"
- "Perbarui aplikasi" (unknown `schema_version`)
- "Laporan tidak sah" (tampered token) / "Terverifikasi AMBA" (valid)
- "Berdasarkan: …" / "Tidak dapat diperiksa: …"
- "Indikatif, bukan bukti pasti" (AI pillar caveat)
- "ELA kurang andal pada PNG/tangkapan layar"
- "CONTOH / SIMULASI" + "Ini data contoh, bukan hasil audit sungguhan"

## 8. Pre-ship checklist (run by eye or grep)

- [ ] Banned-word grep clean: `grep -rnEi "\b(palsu|aman|pasti|terbukti|100%|akurat|dijamin)\b" src/` — every hit is an exact spec string or rewritten.
- [ ] Certainty grep: no "pasti", "terbukti", "dijamin", "100%" outside the disclaimer and the TAMPERED label.
- [ ] No percentage framing of the score ("62% keyakinan" is forbidden).
- [ ] Every jargon word has its one-line explanation on the same screen.
- [ ] No jargon in Essential Mode labels.
- [ ] Sentences ≤ 20 words; no "dilakukan"/"merupakan".
- [ ] Does it blame the user? If yes, rewrite to describe the content or the system.
- [ ] Every status has icon + label, not colour alone.
- [ ] All text Indonesian; no English UI strings.
- [ ] Error/empty copy still says what to do next.

## Do not

- Do not write "palsu", "aman", "pasti", "terbukti", "100%", "akurat", "dijamin" in new copy.
- Do not present the score as a probability or confidence.
- Do not paraphrase `explanation_simple` or generate it with the LLM.
- Do not invent a SUSPICIOUS label, a source name, a publisher, or a number.
- Do not use jargon without a one-line explanation, or any jargon in Essential Mode labels.
- Do not blame, shame or lecture the reader.
- Do not use emoji, exclamation marks, or marketing tone.
- Do not write English placeholder text ("Loading…", "Error").
