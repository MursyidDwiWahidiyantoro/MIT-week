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
   ├── internal/security   SSRF-safe client, upload validation, rate limit, CORS
   ├── internal/pipeline   errgroup, sub-deadlines, degradation, TTL cache
   │      ├── forensics    server-side ELA
   │      ├── linkcheck    URL heuristics
   │      ├── llm          OpenAI-compatible gateway (or mock)
   │      └── grounding    allowlist source verification
   ├── internal/scoring    the ONLY place a score is computed
   ├── internal/token      HMAC-signed share tokens
   └── internal/demo       bundled example dossiers
```

The model never returns a score or status. `internal/scoring` combines the
pillar signals and applies the floors from TECH_SPEC §3.3.

## Menjalankan

```bash
# backend
cd backend
cp ../.env.example ../.env
go run ./cmd/server            # http://localhost:8080
curl localhost:8080/healthz

# frontend
cd frontend
npm install
npm run dev                    # http://localhost:5173
```

`LLM_MOCK=true` is the default, so no model credentials are needed.

## Pemeriksaan

```bash
cd backend  && go test -race ./... && go vet ./...
cd frontend && npm run typecheck && npm run lint && npm run test
cd frontend && npm run build && npm run size && npm run contrast
```

## Environment

| Variable | Default | Notes |
|---|---|---|
| `SERVER_PORT` | `8080` | |
| `ENVIRONMENT` | `development` | production requires `REPORT_SIGNING_KEY` |
| `PUBLIC_APP_URL` | `http://localhost:5173` | CORS origin in production |
| `LLM_MOCK` | `true` | deterministic fake findings |
| `NINEROUTER_BASE_URL` | `http://localhost:20128/v1` | OpenAI-compatible gateway |
| `NINEROUTER_MODEL` | — | required when `LLM_MOCK=false` |
| `REPORT_SIGNING_KEY` | — | ≥ 32 bytes, required in production |
| `MAX_UPLOAD_SIZE_MB` | `10` | |
| `HARD_TIMEOUT_MS` | `12000` | past this, a degraded verdict or a 504 |
| `CACHE_TTL_SECONDS` | `600` | in-memory verdict cache only |

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
- Pemeriksaan fakta hanya memakai sumber yang ada di allowlist.
- Menu bagikan ke AMBA hanya tersedia di peramban yang mendukung Web Share
  Target (Chrome Android). Di iOS gunakan tombol tempel.
- Laporan yang dibagikan hanya memuat ringkasan; gambar dan teks aslinya tidak
  disimpan.

## Struktur

```
api/          OpenAPI contract + Dossier JSON Schema
backend/      Go 1.24 + chi
frontend/     SvelteKit 3 (Svelte 5, TS, Tailwind 3) SPA
configs/      reply templates, trusted-source allowlist
docs/         PROGRESS.md, DECISIONS.md
instruction/  PRD, tech spec, design reference
```
