---
name: amba-forensics-explain
description: Explains AMBA's image-forensics (ELA) pillar to Indonesian users in plain language, and generates honest demo and Edukasi assets for it. Covers the ELA slider/hotspot viewer, bounding-box geometry, reliability by file format, downscale notice, the /edukasi 3-question quiz, and demo-cases data. Trigger when writing or reviewing ELA UI copy ("jelaskan ELA", "apa itu ELA", "kenapa kotak merah", "catatan keandalan", "gambar turun kualitas"), when implementing the SplitSliderViewer or ela.worker, when adding demo cases or quiz images, or when asked whether a heatmap claim is honest.
---

# AMBA Forensics Explain

Source of truth: `instruction/TECH_SPEC_AMBA.md` §3.1, §3.4, §8, §9; `instruction/PRD_AMBA_v10.md` F2, F5, F7, F8; `instruction/DESIGN.md` §3–§9; `instruction/CLAUDE.md` rules 5, 6, 7. Tech Spec wins on conflict.

AMBA's ELA pillar is **one indication among several**, and on its own it cannot reach TAMPERED: `ELA_SOLO_CAP` (TECH_SPEC §3.3 rule 4) caps a solo ELA result at 60, which is SUSPICIOUS. It can still decide a SUSPICIOUS verdict on its own.

## 1. ELA in plain Indonesian

What it actually does: AMBA saves the picture again as JPEG, then compares the two versions pixel by pixel. Areas that were edited often look different from their surroundings, so they stand out as bright spots ("hotspot").

| Length | Draft copy (Indonesian) | Where |
|---|---|---|
| 1 sentence | "AMBA menyimpan ulang gambar lalu membandingkan perbedaannya. Bagian yang pernah diedit sering tampak berbeda dari sekitarnya." | tooltip, `/edukasi` |
| 3 sentences | "AMBA menyimpan ulang gambar dalam format JPEG, lalu membandingkan hasilnya dengan gambar asli. Bagian yang pernah diedit sering berubah berbeda dari bagian lain, sehingga terlihat sebagai titik terang. Titik terang itu hanya petunjuk, bukan bukti gambar palsu." | `/hasil` tab "Timpa Teks" |
| Esensial mode | "AMBA membandingkan gambar asli dengan salinan barunya. Titik terang berarti bagian itu mungkin pernah diedit. Ini hanya petunjuk, bukan bukti." | Esensial, same screen |

These are draft copy, not spec strings. The exact spec strings that must not drift: "ELA kurang andal pada PNG/tangkapan layar", "Gambar ini terindikasi diedit. Jangan langsung percaya.", "Indikatif, bukan bukti pasti", "Berdasarkan: …", "Tidak dapat diperiksa: …".

Jargon rule: never write "DCT", "kuantisasi", "tabel kuantisasi", "artefak kompresi" without a one-line gloss, and never in Esensial mode. Say "cara JPEG menyimpan gambar" instead.

## 2. Reliability by format (TECH_SPEC §3.1, §3.4)

| Image source | `reliability` | What to tell the user |
|---|---|---|
| JPEG-origin | 1.0 | no note |
| PNG / WebP-lossless / screenshot-like | 0.5 | "ELA kurang andal pada PNG/tangkapan layar" + one line why |
| image < 64 px | 0.0 (pillar unavailable) | "Gambar terlalu kecil untuk diperiksa" |

Why, in plain words: a screenshot or PNG has been saved by the phone or app again, so it already contains changes everywhere. That noise hides the real edit, so the check counts less. Never present the reduced weight as a defect of the user.

Reliability note lives in the honest-limits panel; it must be rendered whenever `pillars.ela.reliability < 1` and must never be hidden in a tooltip alone.

## 3. Geometry contract

- Bounding boxes are normalized **0–1000** in order `[ymin, xmin, ymax, xmax]` (Gemini convention). Do not reorder to x,y.
- Convert with the **decoded image's own** width/height: `x% = xmin/10`, `y% = ymin/10`, `w% = (xmax-xmin)/10`, `h% = (ymax-ymin)/10`.
- The viewer container uses `aspect-ratio` from the decoded image. Boxes are percentages of that same box.
- Boxes render in a layer **outside** the `clip-path` slider layer, so they stay visible on both sides of the slider.
- Hotspot grid is 16x16 **cell means** from server ELA (`internal/forensics`), used only for corroboration overlap (IoU ≥ 0.1 or centre inside), not drawn as boxes.

Failure modes to check:
- Letterbox mismatch: container aspect ratio differs from the image → boxes drift.
- Boxes clipped by the slider: box layer nested inside the clipped element.
- Wrong aspect ratio: box math uses natural size while CSS uses `object-fit: contain` frame size.
- Order slip: box drawn as `[x1,y1,x2,y2]` → box transposed.
- Tiny viewport: box outline < 2 px, invisible.

## 4. Honest framing

An ELA hotspot is an **indication of editing, never proof of a hoax**. Editing is normal: crops, resizing, re-compression, screenshots and stickers all create hotspots. A person who edited a screenshot has not necessarily lied.

Phrasing pattern (use, adapt, never strengthen):
- "Bagian ini terlihat berbeda dari sekitarnya — mungkin pernah diedit."
- "Ini petunjuk untuk diperiksa lebih lanjut, bukan bukti."
- "Bisa juga muncul karena gambar diperkecil atau disimpan ulang."
- For a corroborated case only: quote the spec string "Gambar ini terindikasi diedit. Jangan langsung percaya."

Never affirm certainty: "gambar ini palsu", "terbukti diedit", "orang ini memalsukan", "ini bukti", "pasti palsu". Negated forms are required, not banned: keep "bukan bukti" and "bukan bukti pasti", and keep the spec string "Indikatif, bukan bukti pasti" verbatim. Never colour the whole image red or label a hotspot "guilty".

## 5. Downscale notice

The server rejects uploads above 4096 px longest side or > 16 MP (PRD F2 decode check), so no server ELA runs on them. Separately, `ela.worker.ts` downscales above 4096 px for display only (TECH_SPEC §8); the notice belongs to that display downscale, not to server ELA. The user must be told when it applies, because the result changes:

> "Gambar diperkecil dulu agar bisa diperiksa. Hasil ELA bisa berbeda dari gambar ukuran aslinya."

Show it next to the viewer, persistent (not a toast), in both modes, whenever `ela.worker.ts` downscales for display.

## 6. Demo and Edukasi assets

Rules: synthetic screenshots only; heatmaps really computed from them; embed as WebP/PNG data URI < 150 KB each; no real brands, logos, faces, publishers; placeholder URLs `https://example.org/...`; publisher "Sumber Contoh"; watermark "CONTOH / SIMULASI" + banner "Ini data contoh, bukan hasil audit sungguhan"; `is_demo: true`; `tactics[].phrase` must be a substring of the synthetic caption/OCR text.

Step-by-step recipe (generate → server ELA → embed → verify) and the Edukasi quiz design: `references/explainers.md` (load when producing an asset or a quiz question).

## 7. Pre-ship checklist

- [ ] Every heatmap has a legend (what bright = what it does not mean) and a text list of anomalies.
- [ ] Reliability note visible when `< 1`; < 64 px handled as unavailable.
- [ ] Downscale notice present when the image was resized.
- [ ] Box math tested against 9:16, 16:9, 1:1, tall chat screenshot.
- [ ] Boxes visible on both slider sides; alt tabs Asli / ELA exist; anomaly list in text.
- [ ] No affirming certainty word ("ini bukti", "terbukti", "pasti palsu"); negated "bukan bukti" / "bukan bukti pasti" kept where a caveat is due; score not framed as confidence.
- [ ] Demo assets watermarked, synthetic, `< 150 KB`, schema-valid in Go and Vitest.
- [ ] Quiz reveals precomputed ELA; no live compute, no network.
- [ ] Works in both modes; no jargon in Esensial labels.

## Do not

- Do not show a heatmap with no legend, no anomaly list, or as the only evidence.
- Do not call a heatmap "bukti" or say a person faked something.
- Do not use red-only meaning; pair colour with icon and text.
- Do not use a real news screenshot, brand, logo or face in demo or quiz data.
- Do not invent a publisher, URL, percentage or reliability number.
- Do not let the client-computed ELA score drive the verdict; server score only.
- Do not hide the reliability or downscale note.
- Do not use "DCT", "kuantisasi", "artefak kompresi" unglossed or in Esensial mode.
- Do not add live ELA computation to `/edukasi`; assets are precomputed.
