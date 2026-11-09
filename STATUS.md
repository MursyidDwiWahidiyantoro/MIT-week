# STATUS.md — AMBA

> Dokumen ini untuk **anggota tim**, bukan untuk juri. Ditulis ulang 8 November
> 2026 setelah semua requirement PRD (P0, P1, P2) dibangun dan jalur deploy
> disiapkan. Sumber kebenaran tetap `instruction/TECH_SPEC_AMBA.md`; kalau
> dokumen itu dan STATUS.md berbeda, TECH_SPEC yang menang.

---

## 1. Ringkasan satu menit

| Hal | Kondisi |
|---|---|
| Backend Go | **Selesai & hijau**: 385 test lulus di 16 paket, `-race` |
| Frontend (6 halaman + alur hasil) | **Selesai**: audit nyata dengan SSE progresif, demo offline, EN/ID, mode, ukuran teks |
| PRD | **P0, P1, P2 terpasang** (rincian di bagian 4) |
| Kontrak API (OpenAPI ↔ JSON Schema ↔ Go ↔ Zod) | **Sinkron**, termasuk `source_metadata` dan `ocr_text` |
| Aksesibilitas & performa | axe 0 serious/critical (ID dan EN); Lighthouse mobile ≥ 90 di semua rute, a11y 100 |
| CI GitHub Actions | backend · frontend · lighthouse · contract · e2e |
| Siap produksi? | **Ya, setelah kunci diisi**: Vercel + Fly.io/Railway, panduan di `docs/DEPLOY.md` |

---

## 2. Cara menjalankan

### Paling cepat: Docker (lokal / LAN)

```bash
make up          # web http://localhost, API http://localhost:8080
make dev         # hot-reload: Go Air + Vite
make down
```

Panduan lengkap: `docs/DOCKER.md`. Produksi: `docs/DEPLOY.md`.

### Manual

```bash
cd backend  && go run ./cmd/server          # :8080, LLM_MOCK=true
cd frontend && npm install && npm run dev   # :5173 (mode mock)
```

**Penting:** `npm run dev` default `VITE_API_MODE=mock` (data contoh di
browser). Untuk backend sungguhan:

```bash
VITE_API_MODE=http VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

Backend **tidak membaca `.env`** (tanpa `godotenv`); ekspor variabel di shell.
Daftar lengkap variabel ada di `.env.example`.

**Versi:** `go.mod` dipatok **Go 1.24** dan CI memakai 1.24. Jangan `go mod tidy`
dengan toolchain lebih baru tanpa memeriksa direktif `go`. Node ≥ 22.

---

## 3. Pemeriksaan (jalankan sebelum bilang "selesai")

```bash
cd backend  && gofmt -l . && go vet ./... && go test -race ./...
cd backend  && go run ./cmd/gendemo && git status --porcelain   # harus kosong
cd frontend && npm run typecheck && npm run lint && npm run test
cd frontend && npm run build && npm run size && npm run contrast && npm run headers
cd frontend && npm run test:e2e
cd frontend && npm run lighthouse      # dari folder frontend, setelah build
bash tools/check-demo-diff.sh
```

Hasil terakhir (8 Nov 2026):

| Pemeriksaan | Hasil |
|---|---|
| `gofmt` + `go vet` + `go test -race` | bersih; 385 test lulus, 16 paket |
| `gendemo` | deterministik, nol diff |
| `svelte-check` | 0 error, 0 warning |
| `eslint` | bersih |
| `vitest` | 60 test lulus di 10 file |
| `vite build` | sukses, adapter-static |
| bundle | 142,0 KB gz / batas 250 KB |
| contrast | 82 pasang, dua tema |
| headers | `vercel.json` = `security-headers.json` |
| Playwright e2e | 32/32 (axe ID dan EN, paritas mode, i18n, tautan, demo) |
| Lighthouse (mobile, median 5) | performa 93–100, a11y 100, best practices 96–100, SEO 100, CLS 0 |
| demo diff | sinkron |

---

## 4. Fitur: apa yang benar-benar ada

### Backend

| Bagian | Status | Bukti |
|---|---|---|
| `internal/scoring` | ADA | bobot ELA .30 / AI .15 / Fact .35 / Link .20; 6 floor termasuk `LINK_STRONG`; vektor 1–12 + 11b |
| `internal/security` | ADA | SSRF `Dialer.Control` + penolakan target internal, upload, rate limit dengan identitas dari edge host, header, CORS |
| `internal/token` | ADA | HMAC share token, tamper, rotasi |
| `internal/forensics` | ADA | ELA q90, grid 16×16, reliabilitas per format, `Hotspots`, gambar untuk model (JPEG ≤ 1600 px) |
| `internal/linkcheck` | ADA | brand lookalike berbasis token, domain resmi, unroll redirect, `SHORTENER_HIDES`, Safe Browsing; diuji pada data phishing nyata (83,6 % tautan phishing bertema Indonesia diberi peringatan, 4 dari 10.060 situs sah salah ditandai) |
| `internal/metadata` | ADA | oEmbed (YouTube, TikTok, X) dan OpenGraph lewat klien aman; BLOCKED/UNAVAILABLE jujur |
| `internal/llm` | ADA | analisis + klasifikasi dengan `json_schema` strict, input gambar, deteksi injeksi, mock deterministik; bawaan deploy: OpenRouter `nvidia/nemotron-3.5-super-vl:free` + cadangan `openrouter/free` |
| `internal/grounding` | ADA | Google Fact Check Tools + allowlist; hanya URL hasil pencarian AMBA yang bisa terverifikasi |
| `internal/pipeline` | ADA | SSE progresif, anggaran per langkah, cache LRU 500, satu event akhir per pilar |
| `internal/tactics` | ADA | NFKC + verifikasi substring (teks, caption, atau OCR) |
| `internal/reply` | ADA | draf model tervalidasi (≤ 280, URL sumber saja, kata terlarang) + template status × nada |
| `internal/telegram` | ADA | webhook ber-secret atau polling, `/cek`, foto, batas per chat |
| `internal/metrics` | ADA | penghitung anonim Prometheus di listener internal |
| `internal/demo` | ADA | 4 kasus PRD F8 (+ alias `contoh-*`), dihasilkan `cmd/gendemo` |

Endpoint: `GET /healthz` · `POST /api/v1/audit/{url,file,text}` · `POST /api/v1/verify` ·
`POST /api/v1/reply` · `GET /api/v1/demo/{id}` · `POST /telegram/webhook`.
SSE: `stage → partial → verdict → done`, dengan heartbeat `: ping`.

### Frontend

| Halaman / fitur | Status | Catatan |
|---|---|---|
| `/` beranda | ADA | hero + bar tautan (F1), scanner contoh, sekolah scam, CTA |
| `/deteksi` | ADA | tab, dropzone, demo, radar dari SSE asli, `?tab=` dan `?q=` |
| `/hasil` | ADA | verdict, batas jujur, slider ELA, taktik (teks/OCR), pratinjau tautan, sumber, bagikan, balas, kartu bukti |
| `/edukasi` | ADA | 3 modus + 3 soal + **bedah forensik** (3 gambar, ELA asli) |
| `/cek-cepat` | ADA | 6 pertanyaan + sandi keluarga, offline |
| `/tentang` | ADA | tim, privasi (apa yang dikirim ke mana), batas, disclaimer |
| EN/ID | ADA | default Indonesia; pilihan tersimpan, dipasang sebelum paint pertama |
| Disclaimer PRD §10 | ADA | verbatim, berbahasa Indonesia, di footer setiap halaman |
| Mode Jelajah/Esensial | ADA | paritas informasi diuji e2e |
| Ukuran teks, read-aloud, auto-read | ADA | suara mengikuti bahasa (id-ID / en-US) |
| Kartu bukti 1200×630 | ADA | QR + baris INDIKATIF |
| PWA + share target | ADA | app shell offline; tombol tempel dan bot Telegram untuk iOS |

---

## 5. Risiko yang masih terbuka

1. **Belum diuji dengan kunci sungguhan.** Model, Fact Check Tools, Safe
   Browsing, dan Telegram teruji dengan server palsu di unit test, tetapi belum
   pernah dipanggil dengan kunci asli (lingkungan pengembangan ini tidak bisa
   menjangkau OpenRouter). Langkah uji ada di `docs/DEPLOY.md` §6.
2. **Kuota model gratis kecil.** OpenRouter gratis: 20 permintaan per menit
   dan 50 per hari (1.000 per hari setelah top-up $10). Satu audit memakai
   sampai dua panggilan model.
3. **Skor ELA adalah rata-rata global** (TECH_SPEC §3.4). Suntingan kecil
   terlihat jelas di peta panas, tetapi skornya rendah, jadi `ELA_CORROBORATED`
   jarang terpicu. Opsi perbaikan di `docs/DECISIONS.md`.
4. **Share target Android** belum dicek manual di ponsel (Playwright tidak bisa
   membuka menu Share OS).
5. **`$lib` masih lewat `alias` di `vite.config.ts`**, yang sudah deprecated di
   SvelteKit 3 (muncul peringatan saat build). Migrasi ke `#lib/*` sudah
   disiapkan di `package.json`, tetapi belum dijalankan.
6. **Cache dan rate limit per mesin.** Sesuai aturan tanpa database; kalau
   API diskalakan ke beberapa mesin, batasnya berlaku per mesin.
7. **`instruction/CLAUDE.md` dan `.claude/` di-gitignore.** Rekan yang clone tidak
   mendapat playbook. Kalau disengaja, sebutkan di README.

---

## 6. Keputusan desain yang sudah diambil

Semua detail di `docs/DECISIONS.md`. Yang paling sering ditanya:

1. **Skor hanya dihitung di Go.** Model tidak pernah mengembalikan skor/status.
2. **Status fakta dari sumber terverifikasi**, bukan klaim model; hanya URL dari
   pencarian AMBA sendiri yang bisa terverifikasi.
3. **ELA sendirian tidak bisa TAMPERED** (cap 60), dan cap hanya bila ELA tersedia.
4. **Bukti tautan kuat cukup** (`LINK_STRONG`): skor tautan ≥ 45 tanpa bukti lain
   menjadi skor, maksimal 85.
5. **Produksi menolak model tiruan**, tanpa kunci tanda tangan, atau tanpa https.
6. **Token di fragment URL**, dikirim lewat body `POST /verify`.
7. **CSP dihasilkan SvelteKit** (`csp.mode: 'hash'`); `connect-src` dari
   `VITE_API_BASE_URL` saat build.
8. **Bahasa Inggris dibangun dari kode verdict**, bukan diterjemahkan mesin.
9. **Warna memakai peran semantik** (`--on-fill`, `--on-ink`, `--on-brand`,
   `--on-panel-dark`), bukan pasangan `ink`/`white` yang tetap.

---

## 7. Alur kerja tim

- Branch kerja `main`; commit/push milik pemilik repo.
- Prefix: `feat(scoring):`, `fix(ssrf):`, `test(link):`, `docs:`.
- **Kontrak lebih dulu:** perubahan kontrak mengubah `api/openapi.yaml` →
  `api/dossier.schema.json` → `backend/internal/models/dossier.go` →
  `frontend/src/lib/api/schemas.ts` dalam satu commit.
- Teks antarmuka baru masuk ke **dua** kamus: `src/lib/i18n/id.ts` dan `en.ts`
  (TypeScript dan vitest menolak kunci yang hilang).
- Data demo diubah lewat `backend/cmd/gendemo`, bukan dengan menyunting JSON.

### Aturan yang tidak boleh dilanggar

1. Tanpa akun, tanpa database, tanpa menyimpan input pengguna.
2. Jangan pernah mencatat IP, URL, teks, judul, atau nama file.
3. Vonis dihitung di Go.
4. Data demo tidak pernah dipakai untuk permintaan nyata.
5. Jangan mengarang URL, penerbit, atau sumber.
6. Semua teks antarmuka berbahasa Indonesia secara default dan mudah dipahami;
   bahasa Inggris hanya pilihan tambahan.
7. Jangan mengklaim kepastian.

---

## 8. Rambu teknis

- **`$lib` bawaan SvelteKit 3 sudah dihapus**; repo ini menjaganya lewat `alias`
  (lihat risiko 5). Modul baru boleh memakai `#lib/...`.
- **CSP `script-src 'self'` memblokir bootstrap** → halaman kosong. Jangan hardcode.
  Script pre-paint di `app.html` di-hash otomatis saat build.
- **`tokens.css` wajib diimpor sebelum `app.css`.**
- **`go mod tidy` dengan toolchain lebih baru** menulis ulang direktif `go`.
- **Kontras** diperiksa `tools/check-contrast.mjs` (82 pasang, dua tema).
- **Budget bundle** 250 KB gz; sekarang 142,0 KB (total semua JS; JS awal per halaman lebih kecil).
- **Lighthouse** dijalankan dari folder `frontend` setelah `npm run build`.
- **Jangan memakai emoji Unicode**; ikon dari `@lucide/svelte`.
- **Jangan mencetak `err` dari `http.Client.Do`** untuk layanan luar: pesannya
  memuat URL, termasuk kunci API dan kueri pengguna.

---

## 9. Peta file

```
api/                 openapi.yaml · dossier.schema.json
backend/
  cmd/server/        entrypoint, Telegram, metrik, graceful shutdown
  cmd/gendemo/       generator demo + gambar kuis forensik
  fly.toml · railway.json · Dockerfile
  internal/
    scoring/         satu-satunya tempat skor dihitung
    security/        ssrf · upload · ratelimit · headers
    forensics/       ELA server + hotspot grid + gambar untuk model
    linkcheck/       heuristik URL, unroll, Safe Browsing
    metadata/        oEmbed / OpenGraph
    llm/             klien OpenAI-compatible + mock
    grounding/       allowlist + Fact Check Tools
    pipeline/        orkestrasi, SSE progresif, cache
    reply/           Balas Sopan
    telegram/        bot
    metrics/         penghitung anonim
    token/           HMAC share token
    tactics/         verifikasi frasa
    demo/            demo-cases.json (embed)
    httpapi/         router, handler, SSE, error envelope
    models/          Dossier (cermin kontrak)
frontend/
  src/routes/        +page · deteksi · hasil · edukasi · cek-cepat · tentang
  src/lib/
    i18n/            id.ts · en.ts · t()
    audit.svelte.ts  sesi audit + SSE + AbortController
    api/schemas.ts   cermin Zod dari kontrak
    components/      ui · deteksi · hasil · home · mode · a11y · share
    data/            demo-cases · forensic-quiz · tactics · reply-templates
  static/            demo/ · edukasi/ · manifest
  tools/             size · contrast · headers · lighthouse · preload-fonts
  vercel.json
configs/             reply_templates.json · trusted_sources.yaml
docs/                DEPLOY · DOCKER · DECISIONS · PROGRESS · DESIGN-SYSTEM
tools/               check-demo-diff.sh
```

---

## 10. Langkah berikutnya

1. Deploy dengan kunci asli mengikuti `docs/DEPLOY.md`, lalu jalankan uji §6.
2. Cek share target di Android Chrome (PWA terpasang) dan bot Telegram.
3. Putuskan perbaikan skor ELA (risiko 3) sebelum mengandalkan `ELA_CORROBORATED`.
4. Migrasi `$lib` → `#lib` dalam satu commit terpisah.
