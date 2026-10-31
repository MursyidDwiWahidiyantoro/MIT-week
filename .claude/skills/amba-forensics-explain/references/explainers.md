# Explainers and asset recipe

Load when drafting ELA copy, producing a demo asset set, or designing the Edukasi quiz.
Sources: TECH_SPEC §3.1, §3.4, §8, §9; PRD F2, F7, F8; CLAUDE.md rules 5–7.

## 1. Explainer drafts (Indonesian, plain)

### 1.1 Tooltip / one line

> "AMBA menyimpan ulang gambar lalu membandingkan perbedaannya. Bagian yang pernah diedit sering tampak berbeda dari sekitarnya."

### 1.2 Three sentences (`/hasil`, tab Timpa Teks)

> "AMBA menyimpan ulang gambar dalam format JPEG, lalu membandingkan hasilnya dengan gambar asli. Bagian yang pernah diedit sering berubah berbeda dari bagian lain, sehingga terlihat sebagai titik terang. Titik terang itu hanya petunjuk, bukan bukti gambar palsu."

### 1.3 Esensial mode

> "AMBA membandingkan gambar asli dengan salinan barunya. Titik terang berarti bagian itu mungkin pernah diedit. Ini hanya petunjuk, bukan bukti."

Shorter sentences, one idea each. No jargon, no "DCT", no "artefak".

### 1.4 Legend text (under every heatmap)

> "Terang = bagian yang berbeda dari sekitarnya. Bisa karena diedit, bisa juga karena gambar diperkecil atau disimpan ulang."

### 1.5 Reliability note (PNG / screenshot)

> "ELA kurang andal pada PNG/tangkapan layar. Screenshot sudah disimpan ulang oleh aplikasi, jadi perbedaannya tersebar dan sulit dibaca."

### 1.6 Downscale notice

> "Gambar diperkecil dulu agar bisa diperiksa. Hasil ELA bisa berbeda dari gambar ukuran aslinya."

### 1.7 Too small

> "Gambar terlalu kecil untuk diperiksa."

### 1.8 Corroborated case (spec string, quote exactly)

> "Gambar ini terindikasi diedit. Jangan langsung percaya."

### 1.9 Anomaly list entry (text under the viewer)

> "Bagian teks di sekitar judul terlihat berbeda dari sekitarnya. Mungkin pernah diedit. Cek sumber aslinya sebelum percaya."

## 2. Demo asset recipe (one asset set)

Run inside `/home/riko/Projects/AMBA`. No real brands, faces, logos or publishers.

1. **Generate** a synthetic screenshot. `tools/gen_demo_assets.*` (Node or Python stdlib) draws a plain canvas: dark background, fake chat or headline layout, text from the case, watermark "CONTOH / SIMULASI" in a corner, `https://example.org/contoh-cek-fakta` as any URL. Save PNG at the target aspect (e.g. 720x1280 tall chat).
2. **Edit one region** on purpose (paste a flat patch or shift a block) so the ELA has something to find. This is the simulated edit, not a claim about a real person.
3. **Compute the heatmap with the server**, not a mock: call `internal/forensics` on the PNG (small Go program or a test that writes the heatmap PNG). Do not hand-paint a heatmap, and do not use a 1x1 placeholder (PRD v10 fix #10).
4. **Downscale and embed** the heatmap as WebP/PNG, target < 150 KB each, as a data URI in `backend/internal/demo/demo-cases.json`. Keep the source PNG out of the repo if it is large; the demo only needs the data URI.
5. **Fill the Dossier**: `is_demo: true`, `pillars.ela.reliability` matching the source format (PNG → 0.5), `reliability_note`, anomalies with boxes in `[ymin,xmin,ymax,xmax]` 0–1000, `tactics[].phrase` copied as a substring of the synthetic caption, `explanation_simple` from the spec template.
6. **Verify**: copy to `frontend/src/lib/data/demo-cases.json` (CI diff must be clean), validate against the Dossier JSON Schema in Go and Vitest, check each data URI size, and open the SPA offline to confirm zero network calls.

Case ids from PRD F8: `demo-tiktok`, `demo-berita-ig`, `demo-resmi`.

Checklist per asset:
- [ ] Synthetic only, watermark present.
- [ ] Heatmap computed from the actual image.
- [ ] Data URI < 150 KB.
- [ ] No real publisher, brand, face, or invented URL.
- [ ] Box order `[ymin,xmin,ymax,xmax]`, values 0–1000.
- [ ] Reliability matches format.
- [ ] Schema valid in Go and Vitest; demo JSON diff clean.

## 3. Edukasi quiz (3 questions, precomputed ELA)

Client-only, no live compute, no network. Each question shows a synthetic image and one concept; "Buka bedah forensik" reveals the precomputed ELA image plus the explanation.

| # | Concept | Image | Reveal copy |
|---|---|---|---|
| 1 | Hotspot is not hoax | synthetic headline screenshot with a pasted patch | "Ada titik terang di sini. Itu tanda bagian ini mungkin diedit, tapi belum tentu isinya bohong. Cek sumbernya dulu." |
| 2 | Screenshot weakens ELA | the same content saved as a PNG screenshot | "Gambar ini tangkapan layar, jadi ELA kurang andal. Perbedaannya tersebar dan sulit dibaca." |
| 3 | Compression creates hotspots | a clean image re-saved at low quality | "Titik terang muncul walau tidak ada yang diedit. Menyimpan ulang gambar juga bisa membuatnya." |

Rules:
- One concept per question; answer feedback explains the concept, not "benar/salah" only.
- Never say a person faked something, never call a hotspot "bukti".
- Assets follow §2 rules (synthetic, watermark, no real brands), but heatmaps are precomputed and stored as data URIs, not generated at runtime.
- Same information in both modes; Esensial shows the same image, ELA, and explanation with plainer labels and less decoration.
- No score, no percentage, no "keyakinan".
