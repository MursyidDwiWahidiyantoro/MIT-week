# STATUS.md — AMBA

> Dokumen ini untuk **anggota tim**, bukan untuk juri. Ditulis ulang 5 November
> 2026 setelah audit menyeluruh. Sumber kebenaran tetap
> `instruction/TECH_SPEC_AMBA.md`; kalau dokumen itu dan STATUS.md berbeda,
> TECH_SPEC yang menang.

---

## 1. Ringkasan satu menit

| Hal | Kondisi |
|---|---|
| Backend Go (scoring, security, pipeline, token) | **Selesai & hijau** — 264 test lulus, 14 paket |
| Frontend (6 halaman + alur hasil) | **Berfungsi** — audit nyata, unggah gambar, demo offline, mode, ukuran teks |
| Kontrak API (OpenAPI ↔ Go ↔ Zod) | **Sinkron** — server tidak lagi mengirim `null` untuk array; ada test penjaga |
| Aksesibilitas | **Tergerbang** — axe 0 serious/critical di 6 permukaan, dua tema (e2e), kontras 82 pasang di 2 tema |
| CI GitHub Actions | **Hijau** (backend, frontend, contract, e2e) |
| Bisa dipakai untuk demo juri? | **Ya** — `make up`, lalu `/deteksi` → `/hasil` bekerja |

---

## 2. Cara menjalankan

### Paling cepat: Docker

```bash
make up          # web http://localhost, API http://localhost:8080
make dev         # hot-reload: Go Air + Vite
make down
```

Panduan lengkap: `docs/DOCKER.md`.

### Manual

```bash
cd backend  && go run ./cmd/server          # :8080
cd frontend && npm install && npm run dev   # :5173 (mode mock)
```

**Penting:** `npm run dev` default `VITE_API_MODE=mock` (data contoh di
browser). Untuk backend sungguhan:

```bash
VITE_API_MODE=http VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

Backend **tidak membaca `.env`** (tanpa `godotenv`); ekspor variabel di shell.

**Versi:** `go.mod` dipatok **Go 1.24** dan CI memakai 1.24. Jangan `go mod tidy`
dengan toolchain lebih baru tanpa memeriksa direktif `go`.

---

## 3. Pemeriksaan (jalankan sebelum bilang "selesai")

```bash
cd backend  && go vet ./... && go test -race ./...
cd frontend && npm run typecheck && npm run lint && npm run test
cd frontend && npm run build && npm run size && npm run contrast
cd frontend && npm run test:e2e
bash tools/check-demo-diff.sh
```

Hasil terakhir (5 Nov 2026):

| Pemeriksaan | Hasil |
|---|---|
| `go vet` + `go test -race` | 264 test lulus, 14 paket |
| `svelte-check` | 0 error, 0 warning |
| `eslint` | bersih |
| `vitest` | 56 test lulus di 9 file |
| `vite build` | sukses, adapter-static |
| bundle | 120,2 KB gz / batas 250 KB |
| contrast | 82 pasang, dua tema |
| Playwright e2e | 14/14 (termasuk axe 6 permukaan) |
| demo diff | sinkron |

---

## 4. Fitur: apa yang benar-benar ada

### Backend

| Bagian | Status | Bukti |
|---|---|---|
| `internal/scoring` | ADA | bobot ELA .30 / AI .15 / Fact .35 / Link .20; band + 5 floor |
| `internal/security` | ADA | SSRF `Dialer.Control`, upload, rate limit, header, CORS + `Expose-Headers` |
| `internal/token` | ADA | HMAC share token, tamper, rotasi |
| `internal/forensics` | ADA | ELA q90, grid 16×16, reliabilitas per format, `Hotspots` |
| `internal/linkcheck` | ADA | heuristik + unroll redirect via safe client |
| `internal/llm` | ADA | schema-constrained findings + mock deterministik |
| `internal/grounding` | ADA | allowlist + `configs/trusted_sources.yaml` (test paritas) |
| `internal/pipeline` | ADA | errgroup, sub-deadline, degradasi, cache; hotspot + anomaly + injection flag tersambung ke skor |
| `internal/tactics` | ADA | NFKC + verifikasi substring |
| `internal/reply` | **TIDAK ADA** | `/reply` masih 4×3 template deterministik di `httpapi/verify.go`, sekarang sadar status |
| `internal/demo` | ADA | 3 kasus, `is_demo: true` |

Endpoint: `GET /healthz` · `POST /api/v1/audit/{url,file,text}` · `POST /api/v1/verify` ·
`POST /api/v1/reply` · `GET /api/v1/demo/{id}`. SSE: `stage → partial → verdict → done`.

### Frontend

| Halaman / fitur | Status | Catatan |
|---|---|---|
| `/` beranda | ADA | hero + **bar tautan** (F1), scanner contoh, marquee, sekolah scam, CTA |
| `/deteksi` | ADA | tab, dropzone, preset demo, radar dari SSE asli, state error |
| `/hasil` | ADA | verdict, batas jujur, slider ELA, taktik, sumber, bagikan, balas, kartu bukti |
| `/edukasi` | ADA | 3 modus + 3 soal |
| `/cek-cepat` | ADA | 6 pertanyaan + sandi keluarga, offline |
| `/tentang` | ADA | tim, privasi, batas, disclaimer |
| Slider ELA (`SplitSliderViewer`) | ADA | tersambung lewat `ElaViewer`, kotak anomali di luar lapisan clip |
| Worker ELA | ADA | off-thread, cap 4096 px, object URL dibersihkan |
| Demo offline | ADA | `demo-cases.json` dibaca `/deteksi`, nol panggilan API |
| Mode Jelajah/Esensial | ADA | kontrol di header + picker kunjungan pertama; tipe ikut membesar |
| Ukuran teks 100/125/150 | ADA | kontrol di header, tersimpan, skala root |
| Read-aloud id-ID | ADA | feature-detected, berhenti saat unmount |
| Kartu bukti 1200×630 | ADA | QR + disclaimer, palet gelap |
| PWA + share target | ADA | manifest + `service-worker.ts`, fallback tempel |
| axe + Playwright | ADA | 6 permukaan, 0 serious/critical; 14 skenario |

---

## 5. Bug & risiko yang masih terbuka

1. **`/reply` tidak memakai model.** Template deterministik, sadar status. PRD F12
   mengizinkan fallback ini; draf LLM belum disambungkan.
2. **Audit tautan saja masih bisa `INSUFFICIENT`** pada skor 85 (bobot Link 0,20
   < ambang 0,30). Perilaku spec apa adanya; keputusan produk belum diambil.
3. **Mock tidak bisa mencapai TAMPERED lewat pilar fakta** (sumber `example.org`
   ditolak grounding). Disengaja.
4. **`GOOGLE_SAFE_BROWSING_KEY` dan `METRICS_ADDR`** dibaca config, belum dipakai.
5. **ID demo berbeda dari PRD F8** (`contoh-*` vs `demo-tiktok` dll.). Isi dan
   perilakunya sesuai; dicatat di `docs/DECISIONS.md`.
6. **`instruction/CLAUDE.md` dan `.claude/` di-gitignore.** Rekan yang clone tidak
   mendapat playbook. Kalau disengaja, sebutkan di README.

---

## 6. Keputusan desain yang sudah diambil

Semua detail di `docs/DECISIONS.md`. Yang paling sering ditanya:

1. **Skor hanya dihitung di Go.** Model tidak pernah mengembalikan skor/status.
2. **Status fakta dari sumber terverifikasi**, bukan klaim model.
3. **ELA sendirian tidak bisa TAMPERED** (cap 60), dan cap hanya bila ELA tersedia.
4. **ELA buta pada gambar yang sudah q90** — jangan tampilkan skor 0 sebagai "bersih".
5. **Token di fragment URL**, dikirim lewat body `POST /verify`.
6. **Tailwind 3**, token di `tokens.css`.
7. **CSP dihasilkan SvelteKit** (`csp.mode: 'hash'`).
8. **Warna memakai peran semantik** (`--on-fill`, `--on-ink`, `--on-brand`,
   `--on-panel-dark`), bukan pasangan `ink`/`white` yang tetap. `--ink` adalah
   warna teks dan berubah antar-tema, jadi tidak boleh dipakai sebagai latar.

---

## 7. Alur kerja tim

- Branch kerja `main`; commit/push milik pemilik repo.
- Prefix: `feat(scoring):`, `fix(ssrf):`, `test(link):`, `docs:`.
- **Kontrak lebih dulu:** perubahan kontrak mengubah `api/openapi.yaml` →
  `api/dossier.schema.json` → `backend/internal/models/dossier.go` →
  `frontend/src/lib/api/schemas.ts` dalam satu commit.

### Aturan yang tidak boleh dilanggar

1. Tanpa akun, tanpa database, tanpa menyimpan input pengguna.
2. Jangan pernah mencatat IP, URL, teks, judul, atau nama file.
3. Vonis dihitung di Go.
4. Data demo tidak pernah dipakai untuk permintaan nyata.
5. Jangan mengarang URL, penerbit, atau sumber.
6. Semua teks antarmuka berbahasa Indonesia dan mudah dipahami.
7. Jangan mengklaim kepastian.

---

## 8. Rambu teknis

- **`$lib` di SvelteKit 3 dihapus**; pakai `#lib/...`.
- **CSP `script-src 'self'` memblokir bootstrap** → halaman kosong. Jangan hardcode.
- **`tokens.css` wajib diimpor sebelum `app.css`.**
- **`go mod tidy` dengan toolchain lebih baru** menulis ulang direktif `go`.
- **Kontras** diperiksa `tools/check-contrast.mjs` (82 pasang, dua tema).
- **Budget bundle** 250 KB gz; sekarang 120,2 KB.
- **Jangan memakai emoji Unicode**; ikon dari `@lucide/svelte`.

---

## 9. Peta file

```
api/                 openapi.yaml · dossier.schema.json
backend/
  cmd/server/        entrypoint, graceful shutdown
  internal/
    scoring/         satu-satunya tempat skor dihitung
    security/        ssrf · upload · ratelimit · headers
    forensics/       ELA server + hotspot grid
    linkcheck/       heuristik URL + unroll
    llm/             klien 9router + mock
    grounding/       allowlist (cermin configs/trusted_sources.yaml)
    pipeline/        errgroup, sub-deadline, degradasi
    token/           HMAC share token
    tactics/         verifikasi frasa
    demo/            demo-cases.json (embed)
    httpapi/         router, handler, SSE, error envelope
    models/          Dossier (cermin kontrak)
frontend/
  src/routes/        +page · deteksi · hasil · edukasi · cek-cepat · tentang
  src/lib/
    audit.svelte.ts  sesi audit + SSE + AbortController
    sse.ts           parser SSE berbasis fetch
    api/schemas.ts   cermin Zod dari kontrak
    result.svelte.ts hasil in-memory
    mode.svelte.ts   Jelajah/Esensial
    a11y.svelte.ts   ukuran teks + read-aloud
    components/      ui · deteksi · hasil · mode · a11y · share
    forensics/       ela.worker.ts
    data/            demo-cases · tactics · reply-templates
configs/             reply_templates.json · trusted_sources.yaml
docs/                PROGRESS · DECISIONS · DESIGN-SYSTEM · DOCKER
tools/               check-demo-diff.sh
```

---

## 10. Langkah berikutnya

1. Sambungkan draf LLM ke `/reply` (validasi URL + panjang sudah ada tempatnya).
2. Putuskan nasib audit tautan-saja (turunkan ambang floor atau terima INSUFFICIENT).
3. Isi `GOOGLE_SAFE_BROWSING_KEY` atau hapus dari config.
4. Tambah metrik `METRICS_ADDR` atau hapus.
