# AMBA — Aplikasi Mengenali & Bertindak Aman (IN DEVELOPMENT)

Alat bantu literasi digital untuk memeriksa tautan, tangkapan layar, dan pesan
yang mencurigakan. Hasil analisis bersifat **indikatif** — bukan bukti pasti,
bukan alat bukti hukum.

## Arsitektur

```
Browser (SvelteKit SPA)
   │  POST /api/v1/audit/{url,file,text}   (JSON or SSE)
   ▼
Go API (chi)
   ├── internal/security     SSRF-safe client, upload validation, rate limit, CORS
   ├── internal/pipeline     errgroup, real per-pillar events, degradation, TTL cache
   │      ├── forensics     server-side ELA (original bytes)
   │      ├── pagefetch     reads the submitted page through the SSRF client
   │      ├── llm           Maru via 9router, vision + structured findings (or mock)
   │      ├── retrieval     fixed news endpoint, filtered by the allowlist
   │      ├── grounding     allowlist source verification
   │      └── imaging       downscale for the vision model
   ├── internal/scoring      the ONLY place a score is computed
   ├── internal/reply        Balas Sopan: validated model draft + templates
   ├── internal/safebrowsing optional Safe Browsing v4 (placeholder without a key)
   ├── internal/token        HMAC-signed share tokens
   └── internal/demo         bundled example dossiers
```

The model never returns a score or status. `internal/scoring` combines the
pillar signals and applies the floors from TECH_SPEC §3.3.

## Menjalankan

Cukup **Docker**. Tidak perlu memasang Go, Node, atau `make`.

### Paling mudah — satu klik

| Sistem | Caranya |
|---|---|
| **Windows** | Klik dua kali **`start.bat`** |
| **Linux / macOS** | Jalankan **`./start.sh`** (atau klik dua kali di file manager) |

Skrip mengecek Docker, menyalakan semuanya, menunggu sampai siap, lalu
menampilkan alamatnya. Buka alamat itu di browser. Matikan dengan
`stop.bat` / `./stop.sh`.

- Aplikasi: **http://localhost**
- API: **http://localhost:8080**
- Kalau port 80 sedang dipakai, skrip otomatis pindah ke **http://localhost:3000**

Ini stack **sungguhan**: nginx menyajikan SPA hasil build dan mem-proxy `/api/`
ke Go. Tidak ada mock.

> **Satu-satunya prasyarat: Docker.** Kalau belum ada, pasang
> [Docker Desktop](https://www.docker.com/products/docker-desktop/), buka, tunggu
> sampai berstatus running, lalu jalankan skrip lagi. Skrip akan memberi tahu
> sendiri kalau Docker belum terpasang atau belum berjalan.

### Untuk yang terbiasa terminal

```bash
docker compose up --build     # sama dengan start.sh / start.bat
docker compose down           # sama dengan stop.sh / stop.bat
```

`make up` dan `make dev` tetap tersedia untuk pengembangan (hot-reload).

### Pengembangan — hot-reload

Hanya untuk yang mengubah kode; tidak perlu untuk mencoba aplikasi.

```bash
make dev      # API :8080, web :5173 (Go Air + Vite HMR)
```

Tanpa `make` (Windows): `start-dev.bat`.

---

### Manual — dua terminal

**Terminal 1, backend:**

```bash
cd backend
go run ./cmd/server             # http://localhost:8080
curl localhost:8080/healthz     # {"status":"ok"}
```

`LLM_MOCK=true` adalah default, jadi **tidak perlu kredensial model** untuk
mencoba semuanya.

> **Backend tidak membaca file `.env`.** Tidak ada `godotenv` di `go.mod`;
> config hanya memakai `os.Getenv`, jadi `cp .env.example .env` **tidak
> berpengaruh**. Default sudah aman. Untuk mengubah variabel, ekspor di shell:
>
> ```bash
> SERVER_PORT=18080 go run ./cmd/server
> ```

**Terminal 2, frontend:**

```bash
cd frontend
npm install
npm run dev                     # http://localhost:5173
```

> **Kalau port 5173 sudah dipakai proyek lain** (umum di mesin dev), pilih port
> lain. CORS backend menerima origin `localhost` mana pun, jadi tidak perlu
> setting tambahan:
>
> ```bash
> npm run dev -- --port 5199 --strictPort
> ```

#### Penting: `npm run dev` memakai **mode mock**

Tanpa variabel tambahan, `vite dev` menjalankan `VITE_API_MODE=mock`
(lihat `src/lib/api/config.ts`). Artinya seluruh UI berjalan dari data contoh
di dalam bundle, dan muncul banner kuning **“Mode pratinjau”**.

Ini disengaja: kerja frontend tidak boleh tergantung backend hidup. Tapi
artinya **`npm run dev` biasa tidak memanggil Go sama sekali** — jangan pakai
ini untuk demo atau untuk menguji kontrak API.

Untuk memakai backend sungguhan, set dua variabel sekaligus:

```bash
VITE_API_MODE=http VITE_API_BASE_URL=http://localhost:8080/api/v1 \
  npm run dev -- --port 5199
```

Cara memastikan sudah benar:

- Banner kuning “Mode pratinjau” **hilang**.
- Di DevTools → Network ada `POST http://localhost:8080/api/v1/audit/text`
  saat menekan **Mulai periksa**.
- Halaman `/hasil` **tidak** menampilkan pita merah “CONTOH / SIMULASI”.

Kalau ketiganya terpenuhi, hasilnya datang dari Go, bukan dari mock.

---

### Preview build produksi

Untuk melihat hasil akhir persis seperti yang akan dinilai (bundle, CSP, dan
aspek statis lainnya):

```bash
cd frontend
npm run build
npm run preview -- --port 4400
```

Buka **http://localhost:4400**.

Build produksi default-nya `VITE_API_MODE=http`, jadi ia memang mencari backend
di `:8080`. Untuk preview **tanpa** backend:

```bash
VITE_API_MODE=mock npm run build && npm run preview -- --port 4400
```

> **Jalankan `npm run build` lebih dulu, lalu preview.** `vite preview`
> menyajikan aset ber-hash dari `.svelte-kit/output/client`. Kalau kamu
> menjalankan build **saat preview masih hidup**, proses preview mati dengan
> `ENOENT` pada nama CSS lama. Urutannya selalu: hentikan preview → build →
> preview lagi.

Yang perlu dicek di preview:

- Semua halaman terbuka: `/`, `/deteksi`, `/hasil`, `/edukasi`, `/cek-cepat`,
  `/tentang`, dan satu URL ngawur untuk melihat 404.
- Tidak ada scroll horizontal di 360px, 390px, 768px, 1280px.
- Toggle **A / A+ / A++** memperbesar seluruh halaman dan tersimpan setelah reload.
- Toggle **Jelajah / Esensial**: Esensial menjadi satu kolom dan tanpa animasi.
- Tab / Shift+Tab: skip link **“Langsung ke isi utama”** muncul pertama, setiap
  perhentian punya ring fokus hitam 4px.
- Tab **Pesan / Tautan** di kartu hero mengganti teks contoh, meter risiko,
  dan tiga tanda bahaya.
- Kuis “Asli atau scam?” memberi umpan balik lalu bisa diulang.

---

## Tes & pemeriksaan

### Pemeriksaan statis (wajib lulus sebelum commit)

```bash
cd backend  && go vet ./... && go test -race ./...
cd frontend && npm run typecheck && npm run lint && npm run test
cd frontend && npm run build && npm run size && npm run contrast
```

| Perintah | Menjaga apa |
|---|---|
| `npm run typecheck` | Svelte + TypeScript, 0 error / 0 warning |
| `npm run lint` | ESLint, termasuk aturan a11y Svelte |
| `npm run test` | 56 unit test (Vitest + jsdom) |
| `npm run contrast` | 82 pasang warna ≥ 4.5:1 di tema gelap dan terang, dibaca langsung dari `tokens.css` |
| `npm run size` | JS awal ≤ 250 KB gzip |

### End-to-end (membuktikan backend sungguhan)

```bash
cd frontend
npm run test:e2e            # sama dengan: npx playwright test
```

Tes ini **menyalakan sendiri** Go (`LLM_MOCK=true`, port 8080) dan preview build
(port 4173), lalu menjalankan 14 skenario: audit teks sampai `/hasil`, keempat
pilar, demo offline tanpa panggilan API, unggah gambar sampai slider ELA,
token yang diubah, rate limit, error server, tipografi Esensial, dan axe di
enam permukaan (lima rute + `/hasil`), masing-masing di tema gelap dan terang.
Inilah perintah yang dipakai job `e2e` di CI.

Butuh Chromium sekali saja:

```bash
npx playwright install chromium
```

### Audit tampilan (aksesibilitas, keyboard, performa)

Hasil akhir desain diverifikasi di browser sungguhan, bukan dari kode:

| Yang diukur | Target | Hasil terakhir |
|---|---|---|
| axe WCAG 2.1 AA | 0 serious/critical | 0 di 6 permukaan (dijalankan e2e) |
| Ring fokus | semua perhentian terlihat | 14/14 |
| Target sentuh | ≥ 44px | 0 di bawah 44px |
| Hover target utama | bergerak, bukan cuma ganti warna | semua bergerak |
| LCP / CLS (mobile) | < 2.5 s / < 0.05 | 464 ms / 0 |
| Scroll horizontal | 0 di 360–1280px | 0 |

Catatan desain lengkap ada di [`docs/DESIGN-SYSTEM.md`](docs/DESIGN-SYSTEM.md),
riwayat pengerjaan dan daftar bug yang ditemukan ada di
[`docs/PROGRESS.md`](docs/PROGRESS.md).

---

## Kalau ada masalah

| Gejala | Sebab | Solusi |
|---|---|---|
| Banner “Mode pratinjau” tidak hilang | masih mode mock | set `VITE_API_MODE=http` |
| “Pemeriksaan terhenti” / gagal konek | backend belum jalan | `curl localhost:8080/healthz` |
| `EADDRINUSE :5173` | port dipakai proyek lain | `npm run dev -- --port 5199 --strictPort` |
| Halaman kosong setelah build | CSP memblokir bootstrap | jangan ubah `csp.mode` di `svelte.config.js` |
| `/hasil` kosong setelah reload | memang tidak disimpan | ulangi pemeriksaan; hasil hanya ada di memori |
| `make up`: cannot connect to the Docker daemon | daemon Docker belum jalan | `sudo systemctl start docker` (Linux) atau buka Docker Desktop |
| Port 80 terpakai (Docker) | ada web server lain | `FRONTEND_PORT=3000 make up` |


## Environment

| Variable | Default | Notes |
|---|---|---|
| `SERVER_PORT` | `8080` | |
| `ENVIRONMENT` | `development` | production requires `REPORT_SIGNING_KEY` |
| `PUBLIC_APP_URL` | `http://localhost:5173` | CORS origin in production |
| `LLM_MOCK` | `true` | deterministic fake findings; no network |
| `NINEROUTER_BASE_URL` | `http://localhost:20128/v1` | OpenAI-compatible gateway |
| `NINEROUTER_MODEL` | — | combo name, e.g. `Maru`; required when `LLM_MOCK=false` |
| `NINEROUTER_REASONING_EFFORT` | `low` | `minimal`/`low`/`medium`/`high`; low keeps audits inside the budget |
| `FACT_RETRIEVAL` | `true` | news search for the fact pillar (skipped in mock mode) |
| `FACT_SEARCH_URL` | `https://www.bing.com/news/search` | fixed endpoint; only allowlisted publishers survive |
| `REPORT_SIGNING_KEY` | — | ≥ 32 bytes, required in production |
| `MAX_UPLOAD_SIZE_MB` | `10` | |
| `HARD_TIMEOUT_MS` | `20000` | Maru reasoning needs headroom for image audits |
| `METRICS_ADDR` | — | optional separate listener, `/metrics` JSON |
| `CACHE_TTL_SECONDS` | `600` | in-memory verdict cache only |
| `VITE_API_MODE` | `mock` saat `vite dev`, `http` saat build | `http` untuk memakai Go sungguhan |
| `VITE_API_BASE_URL` | `http://localhost:8080/api/v1` saat dev, `/api/v1` saat build | alamat API yang dipanggil browser; juga menjadi `connect-src` di CSP |

## Prinsip

1. Tidak ada akun, basis data, atau penyimpanan input pengguna.
2. Tidak pernah mencatat URL, teks, gambar, judul, atau IP.
3. Skor dihitung di Go; model bahasa tidak pernah memberi skor.
4. Data demo tidak pernah dipakai untuk permintaan nyata.
5. Tidak mengarang URL, penerbit, atau sumber.
6. Semua teks antarmuka berbahasa Indonesia dan mudah dipahami.
7. Tidak pernah mengklaim kepastian.

## Batasan yang diketahui

- Analisis ELA melemah pada PNG dan tangkapan layar, karena gambar sudah
  disimpan ulang sebelum diperiksa.
- Pemeriksaan fakta hanya memakai sumber yang ada di allowlist. Tanpa
  `NINEROUTER_API_KEY` (mode mock) tidak ada pencarian sama sekali; pilar fakta
  tampil "belum terverifikasi", bukan hasil karangan.
- Google Safe Browsing adalah **placeholder** sampai `GOOGLE_SAFE_BROWSING_KEY`
  diisi; `GET /api/v1/ops/status` menyatakannya secara terbuka.
- Model dapat mengabaikan `response_format`; hasilnya tetap dipakai selama
  memuat skor, tetapi bentuknya dinormalisasi di Go.
- Menu bagikan ke AMBA hanya tersedia di peramban yang mendukung Web Share
  Target (Chrome Android). Di iOS gunakan tombol tempel.
- Laporan yang dibagikan hanya memuat ringkasan; gambar dan teks aslinya tidak
  disimpan.

## Struktur

```
api/          OpenAPI contract + Dossier JSON Schema
backend/      Go 1.24 + chi
frontend/     SvelteKit 3 (Svelte 5, TS, Tailwind 3) SPA
  src/lib/styles/tokens.css     satu-satunya sumber warna, tipe, bayangan
  src/lib/components/home/      bagian-bagian beranda
  src/lib/components/ui/Icon.svelte   set ikon stroke inline
configs/      reply templates, trusted-source allowlist
docs/         PROGRESS.md, DECISIONS.md, DESIGN-SYSTEM.md
instruction/  PRD, tech spec, design reference
```

Dokumen desain:

- `docs/waspada-design-spec.md` — spesifikasi visual sumber (neo-brutalist).
- `docs/DESIGN-SYSTEM.md` — bagaimana spesifikasi itu diterapkan di app ini,
  termasuk penyimpangan yang disengaja beserta alasannya.
