# AMBA — Aplikasi Mengenali & Bertindak Aman (IN DEVELOPMENT)

Alat bantu literasi digital untuk memeriksa tautan, tangkapan layar, dan pesan
yang mencurigakan. Hasil analisis bersifat **indikatif** — bukan bukti pasti,
bukan alat bukti hukum.

> Status: Phase 0 (scaffold & contract). UI masih placeholder.

## Struktur

```
AMBA/
├── api/openapi.yaml        # kontrak API (sumber kebenaran)
├── backend/                # Go 1.24 + chi
├── frontend/               # SvelteKit 3 (Svelte 5, TS, Tailwind 3) SPA
├── configs/                # allowlist sumber terpercaya
├── docs/                   # PROGRESS.md, DECISIONS.md
└── instruction/            # PRD, tech spec, design reference
```

## Menjalankan

### Backend

```bash
cd backend
cp ../.env.example ../.env      # sesuaikan bila perlu
go run ./cmd/server             # http://localhost:8080
curl localhost:8080/healthz     # {"status":"ok"}
```

Default `LLM_MOCK=true`, jadi tidak perlu kredensial model untuk pengembangan.

### Frontend

```bash
cd frontend
npm install
npm run dev                     # http://localhost:5173
```

### Pemeriksaan

```bash
cd backend  && go test -race ./...
cd frontend && npm run typecheck && npm run lint && npm run test
cd frontend && npm run build && npm run size
cd frontend && npm run contrast
```

## Prinsip yang tidak boleh dilanggar

1. Tidak ada akun, basis data, atau penyimpanan input pengguna.
2. Tidak pernah mencatat URL, teks, gambar, judul, atau IP.
3. Skor dihitung di Go (`internal/scoring`); model bahasa tidak pernah memberi skor.
4. Tidak pernah menyajikan data demo untuk permintaan nyata.
5. Tidak mengarang URL, penerbit, atau sumber.
6. Semua teks antarmuka berbahasa Indonesia dan mudah dipahami.
7. Tidak pernah mengklaim kepastian.
8. Lingkup mengikuti PRD; tanpa deployment/docker untuk saat ini.

Detail lengkap: `instruction/CLAUDE.md`.

## Lisensi font

Geist dan Geist Mono (OFL) disertakan lewat paket `@fontsource-variable`.
