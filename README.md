# AMBA — Aplikasi Mengenali & Bertindak Aman

Alat bantu literasi digital untuk memeriksa **gambar** dan **tautan** yang
mencurigakan: apakah gambar dibuat dengan AI atau asli, apakah teks di gambar
ditimpa, dan apakah tautan itu asli atau jebakan/hoaks. Hasil analisis bersifat
**indikatif** — bukan bukti pasti, bukan alat bukti hukum.

## Arsitektur

```
Browser (SvelteKit SPA, ID/EN)        Telegram
   │  POST /api/v1/audit/{url,file}        │  POST /telegram/webhook
   │  (SSE: stage → partial → verdict)     │
   ▼                                       ▼
Go API (chi)
   ├── internal/security   SSRF-safe client, upload validation, rate limit, CORS
   ├── internal/pipeline   per-step budgets, progressive SSE, degradation, LRU cache
   │      ├── forensics    server-side ELA (original bytes)
   │      ├── linkcheck    URL heuristics, redirect unrolling, Safe Browsing, domain age
   │      ├── metadata     oEmbed / OpenGraph preview
   │      ├── pagefetch    visible text of an ordinary web page, through the SSRF client
   │      ├── video        yt-dlp + ffmpeg: keyframes and soundtrack of social videos
   │      ├── llm          OpenAI-compatible gateway with fallback (or mock): analyze, classify, video
   │      ├── grounding    fact-check search (Google, Tavily) and allowlist verification
   │      └── retrieval    keyless news search fallback, filtered by the allowlist
   ├── internal/scoring    the ONLY place a score is computed
   ├── internal/reply      Balas Sopan drafts, validated, template fallback
   ├── internal/token      HMAC-signed share tokens
   ├── internal/telegram   bot entry point (PRD P2)
   ├── internal/metrics    anonymous counters on an internal listener
   └── internal/demo       generated example dossiers (cmd/gendemo)
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
Deploy publik (Vercel + Fly.io atau Railway): ikuti [`docs/DEPLOY.md`](docs/DEPLOY.md).

---

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
- Di DevTools → Network ada `POST http://localhost:8080/api/v1/audit/url`
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

- Semua halaman terbuka: `/`, `/deteksi`, `/hasil`, `/edukasi`, `/tentang`,
  dan satu URL ngawur untuk melihat 404.
- Tidak ada scroll horizontal di 360px, 390px, 768px, 1280px.
- Toggle **A / A+ / A++** memperbesar seluruh halaman dan tersimpan setelah reload.
- Toggle **Jelajah / Esensial**: Esensial menjadi satu kolom dan tanpa animasi.
- Tab / Shift+Tab: skip link **“Langsung ke isi utama”** muncul pertama, setiap
  perhentian punya ring fokus hitam 4px.
- Tab **Gambar / Tautan** di kartu hero mengganti contoh, meter risiko, dan
  tiga tanda bahaya.
- Kuis “Asli atau jebakan?” memberi umpan balik lalu bisa diulang.

---

## Tes & pemeriksaan

### Pemeriksaan statis (wajib lulus sebelum commit)

```bash
cd backend  && gofmt -l . && go vet ./... && go test -race ./...
cd backend  && go run ./cmd/gendemo && git status --porcelain   # harus kosong
cd frontend && npm run typecheck && npm run lint && npm run test
cd frontend && npm run build && npm run size && npm run contrast && npm run headers
```

| Perintah | Menjaga apa |
|---|---|
| `go test -race ./...` | 385 test di 16 paket, termasuk vektor skor 1–12 |
| `go run ./cmd/gendemo` | data demo dan gambar kuis dibuat ulang tanpa diff |
| `npm run typecheck` | Svelte + TypeScript, 0 error / 0 warning |
| `npm run lint` | ESLint, termasuk aturan a11y Svelte |
| `npm run test` | 60 unit test (Vitest + jsdom), termasuk paritas kamus ID/EN |
| `npm run contrast` | 82 pasang warna ≥ 4.5:1 di tema gelap dan terang, dibaca langsung dari `tokens.css` |
| `npm run size` | total JS ≤ 250 KB gzip (sekarang 142,0 KB) |
| `npm run headers` | header di `vercel.json` sama dengan `security-headers.json` |
| `npm run lighthouse` | dari folder `frontend` setelah build: performa/best practices/SEO ≥ 90, a11y ≥ 95, CLS ≤ 0,1 di setiap rute |

### End-to-end (membuktikan backend sungguhan)

```bash
cd frontend
npm run test:e2e            # sama dengan: npx playwright test
```

Tes ini **menyalakan sendiri** Go (`LLM_MOCK=true`, port 8080) dan preview build
(port 4173), lalu menjalankan semua skenario: audit tautan sampai `/hasil`, keempat
pilar, demo offline tanpa panggilan API, unggah gambar sampai slider ELA,
tautan yang tidak bisa dibaca, token yang diubah, rate limit, error server,
paritas Mode Jelajah/Esensial, read-aloud, pindah bahasa, dan axe di enam
permukaan (tema gelap dan terang) serta lima rute dalam bahasa Inggris.
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
| Halaman kosong setelah build | CSP memblokir bootstrap | jangan ubah `csp.mode` di `vite.config.ts` |
| `/hasil` kosong setelah reload | memang tidak disimpan | ulangi pemeriksaan; hasil hanya ada di memori |
| `make up`: cannot connect to the Docker daemon | daemon Docker belum jalan | `sudo systemctl start docker` (Linux) atau buka Docker Desktop |
| Port 80 terpakai (Docker) | ada web server lain | `FRONTEND_PORT=3000 make up` |


## Environment

Daftar lengkap dengan komentar ada di [`.env.example`](.env.example).

| Variable | Default | Notes |
|---|---|---|
| `SERVER_PORT` | `8080` | falls back to a PaaS-injected `PORT` |
| `ENVIRONMENT` | `development` | production requires `REPORT_SIGNING_KEY`, an https `PUBLIC_APP_URL` and `LLM_MOCK=false` |
| `PUBLIC_APP_URL` | `http://localhost:5173` | CORS origin and the link in Telegram replies |
| `CORS_EXTRA_ORIGINS` | — | more exact origins, comma-separated |
| `TRUSTED_PROXY_CIDRS` / `CLIENT_IP_HEADER` | — | `*` + `Fly-Client-IP` (Fly.io) or `X-Real-IP` (Railway) |
| `LLM_MOCK` | `true` | deterministic fake findings; refused in production |
| `NINEROUTER_BASE_URL` | `http://localhost:20128/v1` | OpenAI-compatible gateway; deployments use OpenRouter (`https://openrouter.ai/api/v1`) |
| `NINEROUTER_API_KEY` / `NINEROUTER_MODEL` | — | required when `LLM_MOCK=false`; default model `nvidia/nemotron-3.5-super-vl:free` (free, vision) |
| `LLM_EXTRA_BODY` | — | JSON object merged into every model request, e.g. OpenRouter fallbacks and `reasoning` off |
| `REPORT_SIGNING_KEY` | — | ≥ 32 bytes, required in production |
| `GOOGLE_API_KEY` | — | Fact Check Tools + Safe Browsing (`GOOGLE_FACTCHECK_KEY` / `GOOGLE_SAFE_BROWSING_KEY` override) |
| `SIGHTENGINE_API_USER` / `SIGHTENGINE_API_SECRET` | — | optional AI-image detector for Jejak AI; both or neither |
| `DOMAIN_AGE_LOOKUP` | `true` | RDAP domain age for links (`NEW_DOMAIN`, `YOUNG_DOMAIN`); only the domain name is sent |
| `TELEGRAM_BOT_TOKEN` | — | enables the bot; `TELEGRAM_MODE` = `webhook` / `polling` / `off` |
| `METRICS_ADDR` | — | internal Prometheus listener, off when empty |
| `MAX_UPLOAD_SIZE_MB` | `10` | |
| `HARD_TIMEOUT_MS` | `20000` | past this, a degraded verdict or a 504; the model gets this minus 3 s (a video link has its own `VIDEO_TIMEOUT_MS`) |
| `NINEROUTER_REASONING_EFFORT` | `low` | `minimal`/`low`/`medium`/`high`, sent as `reasoning_effort` unless `LLM_EXTRA_BODY` sets reasoning |
| `FACT_RETRIEVAL` / `FACT_SEARCH_URL` | `true` / Bing News RSS | keyless fact search used when neither `GOOGLE_API_KEY` nor `TAVILY_API_KEY` is set (skipped in mock mode); only allowlisted publishers survive |
| `TAVILY_API_KEY` | — | free fact-check search restricted to the fact-checkers sites |
| `LLM_FALLBACK_*`, `LLM_AUDIO_INPUT`, `VIDEO_TIMEOUT_MS` | — | second gateway, audio input, budget for social video links (see `.env.example`) |
| `CACHE_TTL_SECONDS` | `600` | in-memory verdict cache only |
| `VITE_API_MODE` | `mock` saat `vite dev`, `http` saat build | `http` untuk memakai Go sungguhan |
| `VITE_API_BASE_URL` | `http://localhost:8080/api/v1` saat dev, `/api/v1` saat build | alamat API yang dipanggil browser; juga menjadi `connect-src` di CSP |

## Prinsip

1. Tidak ada akun, basis data, atau penyimpanan input pengguna.
2. Tidak pernah mencatat URL, teks, gambar, judul, atau IP.
3. Skor dihitung di Go; model bahasa tidak pernah memberi skor.
4. Data demo tidak pernah dipakai untuk permintaan nyata.
5. Tidak mengarang URL, penerbit, atau sumber.
6. Semua teks antarmuka berbahasa Indonesia secara default dan mudah dipahami;
   bahasa Inggris tersedia sebagai pilihan.
7. Tidak pernah mengklaim kepastian.

## Batasan yang diketahui

- Analisis ELA melemah pada PNG dan tangkapan layar, karena gambar sudah
  disimpan ulang sebelum diperiksa.
- Pemeriksaan fakta hanya memakai sumber yang ada di allowlist.
- Google Safe Browsing adalah **placeholder** sampai `GOOGLE_SAFE_BROWSING_KEY`
  diisi; `GET /api/v1/ops/status` menyatakannya secara terbuka.
- Skor ELA adalah rata-rata seluruh gambar: suntingan kecil terlihat di peta
  panas tetapi skornya tetap rendah (lihat `docs/DECISIONS.md`).
- Isi tautan media sosial yang dikunci login tidak bisa dibaca; hasil
  mengatakannya dan menawarkan unggah tangkapan layar.
- Menu bagikan ke AMBA hanya tersedia di peramban yang mendukung Web Share
  Target (Chrome Android). Di iOS gunakan tombol tempel atau bot Telegram.
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
docs/         DEPLOY.md, DOCKER.md, PROGRESS.md, DECISIONS.md, DESIGN-SYSTEM.md
instruction/  PRD, tech spec, design reference
```

Dokumen desain:

- `docs/waspada-design-spec.md` — spesifikasi visual sumber (neo-brutalist).
- `docs/DESIGN-SYSTEM.md` — bagaimana spesifikasi itu diterapkan di app ini,
  termasuk penyimpangan yang disengaja beserta alasannya.
