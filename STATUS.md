# STATUS.md — AMBA

> Dokumen ini untuk **anggota tim**, bukan untuk juri. Isinya: apa yang sudah
> jadi, apa yang belum, apa yang rusak, dan bagaimana cara kerja repo ini.
> Ditulis 1 November 2026. Sumber kebenaran tetap
> `instruction/TECH_SPEC_AMBA.md`; kalau dokumen itu dan STATUS.md berbeda,
> TECH_SPEC yang menang.

---

## 1. Ringkasan satu menit

| Hal | Kondisi |
|---|---|
| Backend Go (scoring, security, pipeline, token) | **Selesai & hijau** — 261 test lulus |
| Frontend (5 halaman, komponen hasil) | **Sebagian** — halaman ada, tapi alur utama **belum bisa dipakai** |
| Kontrak API (OpenAPI ↔ Go ↔ Zod) | **Tidak sinkron saat runtime** — lihat §5 nomor 1 |
| CI GitHub Actions | **Hijau** (run terakhir sukses) |
| Bisa dipakai untuk demo juri? | **Belum.** Audit nyata berakhir di pesan error, tidak pernah sampai `/hasil` |

Kalau kamu baru bergabung: **baca §5 (Bug & risiko) lebih dulu.** Empat jam kerja
bisa terbuang kalau kamu menganggap halaman `/hasil` sudah jalan.

---

## 2. Cara menjalankan

```bash
# 1. backend  -> http://localhost:8080
cd backend && go run ./cmd/server
curl localhost:8080/healthz    # {"status":"ok"}

# 2. frontend -> http://localhost:5173
cd frontend && npm install && npm run dev
```

**Penting soal `.env`:** backend **tidak membaca file `.env`**. Tidak ada
`godotenv` di `go.mod`; config hanya memanggil `os.Getenv`. Jadi
`cp .env.example .env` tidak berpengaruh apa-apa — default sudah aman
(`LLM_MOCK=true`), dan kalau perlu mengubah variabel, ekspor di shell:

```bash
SERVER_PORT=18080 go run ./cmd/server
```

**Catatan port:** di mesin dev yang dipakai untuk audit ini, port 5173 sudah
dipakai proyek lain. Jalankan `npx vite dev --port 5180 --strictPort` bila perlu.
Backend menerima origin itu lewat `PUBLIC_APP_URL`, jadi CORS tetap jalan.

**Catatan versi:** `go.mod` dipatok ke **Go 1.24** dan CI memakai 1.24. Toolchain
lokal 1.26/1.27 akan menulis ulang direktif `go` kalau kamu menjalankan
`go mod tidy` — jangan commit hasilnya tanpa memeriksa baris itu.

---

## 3. Pemeriksaan (jalankan sebelum bilang "selesai")

```bash
cd backend  && go vet ./... && go test -race ./...
cd frontend && npm run typecheck && npm run lint && npm run test
cd frontend && npm run build && npm run size && npm run contrast
bash tools/check-demo-diff.sh          # demo JSON backend == frontend
```

Hasil terakhir di mesin lokal (1 Nov 2026):

| Pemeriksaan | Hasil |
|---|---|
| `go vet` + `go test -race` | 261 test lulus, 14 paket, 125 fungsi test |
| `svelte-check` | 0 error, 0 warning |
| `eslint` | bersih |
| `vitest` | 30 test lulus di 5 file |
| `vite build` | sukses, adapter-static |
| bundle | 102,2 KB gz / batas 250 KB |
| contrast | 22 pasang lulus |
| demo diff | sinkron |

Yang **tidak** ada padahal tertulis di `package.json`: `npm run test:e2e`.
Tidak ada `playwright.config.*` dan folder `frontend/tests/` kosong. Jangan
percaya klaim e2e/axe di `docs/PROGRESS.md` — belum pernah ada yang menjalankannya.

---

## 4. Fitur: apa yang benar-benar ada

Legenda: **ADA** = jalan · **SEBAGIAN** = kode ada tapi tidak tersambung/terbatas ·
**TIDAK ADA** = belum dibuat.

### Backend (kuat — ini aset terbesar tim)

| Bagian | Status | Bukti |
|---|---|---|
| `internal/scoring` — mesin skor deterministik | ADA | bobot ELA .30 / AI .15 / Fact .35 / Link .20; band: ≤20 SAFE, ≤60 SUSPICIOUS, >60 TAMPERED, tanpa bobot cukup → INSUFFICIENT |
| 5 floor: `FACT_CONTRADICTED`, `LINK_MALICIOUS`, `ELA_CORROBORATED`, `ELA_SOLO_CAP`, `INJECTION_GUARD` | SEBAGIAN | 3 jalan; `LINK_MALICIOUS` hanya lewat skor ≥90 (Safe Browsing tidak tersambung); `INJECTION_GUARD` tidak pernah terpicu |
| `internal/security` — SSRF, upload, rate limit, header, CORS | ADA | `Dialer.Control` memvalidasi IP hasil resolve; `Proxy: nil`; batas 4096 px; magic bytes; animasi ditolak |
| `internal/token` — HMAC share token | ADA | round-trip, tamper, rotasi, ukuran |
| `internal/forensics` — ELA di server | ADA | re-encode q90, grid 16×16, reliabilitas per format |
| `internal/linkcheck` — heuristik URL | ADA | lookalike brand, punycode, TLD, unroll redirect |
| `internal/llm` + mock | ADA | mock deterministik; **tanpa tool loop** |
| `internal/grounding` — verifikasi sumber | SEBAGIAN | allowlist **hardcoded di Go**, `configs/trusted_sources.yaml` tidak ada |
| `internal/pipeline` — errgroup, sub-deadline, degradasi | ADA | sub-deadline 4 s; `degraded:true` saat ELA gagal pada input file |
| `internal/tactics` | ADA | NFKC + normalisasi spasi, substring terverifikasi |
| `internal/reply` | TIDAK ADA | direktori kosong; `/reply` hanya 3 string hardcoded di `httpapi/verify.go` |
| `internal/demo` + `GET /demo/{id}` | ADA | 3 kasus, tervalidasi skema |

**Endpoint yang terdaftar (7 path, 8 operasi):**
`GET /healthz` · `POST /api/v1/audit/url` · `POST /api/v1/audit/file` ·
`POST /api/v1/audit/text` · `POST /api/v1/verify` · `POST /api/v1/reply` ·
`GET /api/v1/demo/{id}`

SSE mengirim `stage → partial → verdict → done` (plus `error`). Urutan ini sudah
diverifikasi dengan `curl` langsung ke server.

### Frontend

| Halaman / fitur | Status | Catatan |
|---|---|---|
| `/` beranda | ADA | hero + input tautan, 3 jenis konten, 3 langkah |
| `/deteksi` | SEBAGIAN | tab, dropzone, preset, radar dari event SSE asli (tanpa timer palsu), state 429/offline/504/415 |
| `/hasil` | SEBAGIAN | 5 cabang: terverifikasi / tidak sah / tak terjangkau / tidak tersimpan / hasil hidup |
| `/edukasi` | ADA | 3 pertanyaan + penjelasan, tanpa gambar |
| `/tentang` | ADA | manifesto privasi + batas |
| Taktik manipulasi (`HighlightedText`) | ADA | 8 taktik di `tactics.json` |
| Balas Sopan (`ReplyCard`) | SEBAGIAN | hanya template, tanpa panggilan LLM |
| Kartu bukti 1200×630 (`ProofCardExporter`) | ADA | QR + baris disclaimer |
| Bagikan (`ShareButton`) | ADA | token di fragment, POST ke `/verify` |
| Read-aloud id-ID | ADA | feature-detected, berhenti saat unmount |
| Ukuran teks 100/125/150 | ADA | mengubah `font-size` root, tersimpan di localStorage |
| Mode Jelajah/Esensial | SEBAGIAN | hanya mengubah gerak; **tipografi tidak berubah** |
| **Slider ELA (`SplitSliderViewer`)** | TIDAK TERSAMBUNG | kode lengkap, tidak diimpor halaman mana pun |
| **Worker ELA (`ela.worker.ts`)** | TIDAK TERSAMBUNG | kode lengkap, tidak diimpor |
| **Contoh kasus demo** | TIDAK TERSAMBUNG | `demo-cases.json` ada & sinkron, tidak dibaca UI |
| PWA + share target | TIDAK ADA | tidak ada manifest, service worker, `/share-target` — padahal `/tentang` dan README menyebutnya |
| `/cek-cepat` + Sandi Keluarga | TIDAK ADA | fitur P1 di PRD §8.5 |
| Playwright e2e + axe | TIDAK ADA | dependensi terpasang, tidak dipakai |

---

## 5. Bug & risiko (URUT PRIORITAS — baca ini)

### P0-1 · Audit nyata gagal di browser, tidak pernah sampai `/hasil`

**Gejala:** pilih preset "Pesan berantai", klik "Mulai periksa" → radar berjalan,
lalu muncul *"Pemeriksaan terhenti sebelum selesai. Coba lagi."* URL tetap
`/deteksi`, tidak ada hasil.

**Sebab:** backend mengirim `null` untuk beberapa field yang di kontrak
seharusnya array, dan string kosong untuk `fact.status`:

```json
"ela":  {"anomalies": null}
"fact": {"status": "", "sources": null}
"link": {"redirect_chain": null, "reasons": null}
```

`frontend/src/lib/api/schemas.ts` menuntut array dan enum. `parseDossier()`
melempar ZodError, `AuditSession` menangkapnya sebagai kegagalan jaringan
(`catch` generik), jadi UI menampilkan pesan "coba lagi" yang menyesatkan.

**Bukti:** payload asli dari `POST /audit/text` diuji dengan `dossierSchema` →
`SUCCESS=false`, tiga issue: `pillars.ela.anomalies`, `pillars.link.redirect_chain`,
`pillars.link.reasons`. Diuji juga di Chromium: `verdict block count: 0`,
`H1: "Periksa konten"`.

**Perbaikan yang disarankan (pilih satu, jangan dua-duanya):**
- Backend: jangan kirim `null` untuk slice — inisialisasi ke `[]`; `fact.status`
  default `UNVERIFIED`. Ini yang paling benar karena kontrak sudah bilang array.
- Frontend: longgarkan Zod jadi `.nullish().transform(v => v ?? [])` — hanya
  sebagai pertahanan, bukan pengganti perbaikan backend.
- Tambahan wajib: `catch` di `AuditSession` harus membedakan `ZodError` dari
  kegagalan jaringan, supaya pesan errornya jujur.

### P0-2 · `/deteksi` → `/hasil` memakai reload penuh

`frontend/src/routes/deteksi/+page.svelte` memakai
`window.location.href = '/hasil'`. Itu memuat ulang aplikasi, dan karena hasil
disimpan **in-memory** (`result.svelte.ts`, sesuai aturan "tanpa penyimpanan"),
hasilnya hilang → `/hasil` menampilkan "Hasil tidak tersimpan".

Bahkan setelah P0-1 diperbaiki, cabang "hasil hidup" tetap tidak tercapai lewat
alur normal. Gunakan `goto('/hasil')` dari `$app/navigation`.

### P0-3 · Progressive reveal tidak jalan

`audit.svelte.ts` menerima event `partial`, lalu:

```ts
case 'partial':
  break;          // dibuang
```

Akibatnya skeleton per pilar tidak pernah dipakai, dan janji "hasil muncul
bertahap" di PRD/beranda tidak benar. Ini juga membuat komponen `Skeleton` dan
prop `pending` di `EvidenceAccordion` jadi kode mati.

### P1-4 · Mode Esensial hanya mengubah gerak

`mode.svelte.ts` menulis `document.documentElement.dataset.mode`, tetapi tidak
ada satu pun aturan CSS yang membaca `[data-mode]`. Terverifikasi: ukuran `h1`
tetap **52px** di kedua mode. `INFO_KEYS` didefinisikan tapi tidak dipakai, jadi
paritas informasi antar-mode tidak ditegakkan.

### P1-5 · Slider ELA dan contoh demo tidak tersambung

`SplitSliderViewer.svelte`, `ela.worker.ts`, dan `demo-cases.json` lengkap tapi
nol impor. Konsekuensi: unggah gambar tidak menampilkan heatmap (padahal ini
fitur utama yang dijual), dan tombol demo tidak ada. Selain itu semua demo
punya `share_token: ""`, jadi panel bagikan/balas rusak untuk kasus demo.

### P1-6 · Fitur yang diklaim ada padahal tidak

| Klaim | Lokasi klaim | Kenyataan |
|---|---|---|
| Menu bagikan PWA | `README.md`, `/tentang` | tidak ada manifest/service worker |
| e2e Playwright + axe 0 violation | `package.json`, `docs/PROGRESS.md` | tidak ada config/spec |
| `internal/reply` | `TECH_SPEC` §1 | direktori kosong |
| `configs/trusted_sources.yaml` | `TECH_SPEC` §1, §6 | hardcoded di `grounding.go` |

Untuk juri, klaim yang tidak bisa ditunjukkan lebih berbahaya daripada fitur yang
tidak ada. Perbaiki kode atau hapus kalimatnya.

### P1-7 · `/hasil` salah memberi label pada kegagalan `/verify`

`verifyToken()` mengembalikan `invalid` untuk semua respons non-2xx. Akibatnya
429 (rate limit) dan 5xx ditampilkan sebagai **"Laporan tidak sah"** — tuduhan
bahwa bukti pengguna dipalsukan, padahal servernya yang bermasalah. Juga tidak
ada state loading: selama verifikasi berjalan, yang tampil "Hasil tidak tersimpan".

### P2-8 · Sisa hal kecil

- `INJECTION_GUARD` tidak mungkin terpicu (`injectionSuspected` hardcoded `false`).
- `GOOGLE_SAFE_BROWSING_KEY` dan `METRICS_ADDR` dibaca config, tidak dipakai.
- `ocr_text` diterima lalu dibuang.
- 504 didokumentasikan, tidak pernah dikirim; `errUpstreamTimeout` tidak dipakai.
- `docs/PROGRESS.md` saling bertentangan: satu bagian bilang "semua fase selesai",
  bagian lain "berhenti di Phase 2, Phase 4 belum mulai". Angka lama (6 test,
  42,1 KB) juga tidak diperbarui.
- **Skor risiko tampil dua kali** di `/hasil`: di dalam `StatusBadge` dan lagi di
  `<p class="meta">` paling bawah.
- **Tidak ada `+error.svelte`** — error tak tertangani jatuh ke halaman default
  SvelteKit, bukan ke UI AMBA.
- **`instruction/CLAUDE.md` bilang Go 1.23**, padahal `go.mod`, CI, dan README
  memakai 1.24.
- **CI tidak memeriksa sinkronisasi kontrak.** Job `contract` hanya memastikan
  string `openapi` diawali `3.`, lalu menjalankan demo-diff. Justru rantai
  OpenAPI ↔ Go ↔ Zod yang putus (P0-1) tidak pernah diperiksa.
- **`docs/design/` tidak ada** — artefak `frontend-prd-fidelity`
  (`00-contract`, `01-brief`, `02-system`, `03-motion`) tidak pernah dibuat.
- `instruction/CLAUDE.md` dan `.claude/` **di-gitignore**. Rekan yang clone tidak
  mendapat playbook maupun 4 skill proyek. Kalau ini disengaja, sebutkan di README.

---

## 6. Keputusan desain yang sudah diambil (jangan diubah tanpa alasan)

Semua detail ada di `docs/DECISIONS.md`. Yang paling sering ditanya:

1. **Skor hanya dihitung di Go.** Model bahasa tidak pernah mengembalikan skor
   atau status. Jangan "disederhanakan" di frontend.
2. **Status fakta berasal dari sumber terverifikasi**, bukan klaim model. Klaim
   `CONTRADICTS` tanpa sumber lolos allowlist diturunkan ke `UNVERIFIED`.
3. **ELA tidak bisa sendirian mencapai TAMPERED** (cap 60), dan cap hanya berlaku
   kalau ELA benar-benar tersedia.
4. **ELA buta pada gambar yang sudah tersimpan di kualitas q90** — perbedaan
   re-encode jadi nol. Karena itu skor ELA 0 tidak boleh ditampilkan sebagai
   "bersih"; selalu tampilkan catatan keandalan.
5. **Token di fragment URL**, dikirim lewat body `POST /verify`, supaya tidak
   pernah masuk log server.
6. **Tailwind 3, bukan 4**; token warna di `tokens.css`, `tailwind.config.ts`
   hanya memetakan `var(--…)`.
7. **CSP dihasilkan SvelteKit** (`csp.mode: 'hash'`), bukan header statis.
   `security-headers.json` hanya menyimpan `frame-ancestors`.
8. **`tokens.css` wajib diimpor sebelum `app.css`.**

---

## 7. Pertanyaan produk yang belum diputuskan

1. **Audit tautan saja selalu INSUFFICIENT.** Skor 85 pun tidak menghasilkan
   vonis, karena bobot Link 0,20 < ambang 0,30. Sesuai spec apa adanya. Pilihan:
   turunkan ambang floor ke 85, beri bobot minimum untuk Link, atau terima
   INSUFFICIENT. Perubahan harus mengubah TECH_SPEC §3.3 + tabel test bersamaan.
2. **Copy untuk INSUFFICIENT karena satu pilar.** "Belum cukup data" terbaca
   seperti kegagalan, bukan "sudah diperiksa, tidak ada temuan".
3. **Mock tidak bisa mencapai TAMPERED lewat pilar fakta** karena mock menyitir
   `example.org`. Ini disengaja (aturan "jangan mengarang sumber"). Jangan
   "diperbaiki" dengan meng-allowlist example.org.

---

## 8. Alur kerja tim

### Commit & branch
- Branch kerja: `main`. Remote: `github.com/fujimaruC/MIT-week`.
- **Commit dan push milik pemilik repo.** Jangan jalankan `git add/commit/push`
  tanpa diminta. Jangan menambahkan `Co-Authored-By`.
- Prefix yang dipakai: `feat(scoring):`, `fix(ssrf):`, `test(link):`, `docs:`.

### Aturan yang tidak boleh dilanggar (dari `instruction/CLAUDE.md`)
1. Tanpa akun, tanpa database, tanpa menyimpan input pengguna.
2. **Jangan pernah mencatat** IP, URL, teks, judul, atau nama file.
3. Vonis dihitung di Go; model tidak memberi skor.
4. Data demo tidak pernah dipakai untuk permintaan nyata.
5. Jangan mengarang URL, penerbit, atau sumber.
6. Semua teks antarmuka berbahasa Indonesia dan mudah dipahami.
7. Jangan mengklaim kepastian.
8. Jangan menambah lingkup di luar PRD (tanpa auth, PDF, video, Docker/deploy).

### Kontrak lebih dulu
Setiap perubahan kontrak harus mengubah **empat** tempat dalam satu commit:
`api/openapi.yaml` → `api/dossier.schema.json` → `backend/internal/models/dossier.go`
→ `frontend/src/lib/api/schemas.ts`. P0-1 di atas adalah contoh persis apa yang
terjadi kalau rantai ini putus.

### Skill proyek (lokal, tidak ter-commit)
Empat skill di `.claude/skills/` menegakkan aturan copy dan UX. Pakai saat
menyentuh area terkait:

| Skill | Untuk |
|---|---|
| `amba-voice` | semua string Indonesia, istilah, larangan overclaim |
| `amba-trust-ux` | skor, band status, panel batas, degraded, demo |
| `amba-sse-ux` | radar, progres, aria-live, skeleton, kegagalan stream |
| `amba-forensics-explain` | copy ELA, slider, geometri kotak, demo/edukasi |

---

## 9. Rambu teknis (jebakan yang sudah pernah kena)

- **`$lib` di SvelteKit 3 dihapus.** Pakai `#lib/...` atau path relatif. Alias
  `$lib` saat ini ditambahkan manual di `vite.config.ts` dan memunculkan warning
  deprecated — sebaiknya migrasi ke subpath import.
- **CSP `script-src 'self'` memblokir bootstrap inline SvelteKit** → halaman
  kosong tanpa error yang jelas. Jangan hardcode CSP di header.
- **Import `tokens.css` terlambat** → semua `var(--…)` jatuh ke default browser
  (serif, h1 16px). Selalu impor pertama di `+layout.svelte`.
- **`go mod tidy` dengan toolchain lebih baru** menulis ulang `go 1.24` menjadi
  1.26/1.27 dan memecahkan CI.
- **Rasio kontras** diperiksa `tools/check-contrast.mjs` (22 pasang). Kalau
  mengubah `tokens.css`, jalankan `npm run contrast`.
- **Budget bundle** 250 KB gz; saat ini 102,2 KB. `html-to-image` dan `qrcode`
  diimpor malas di `ProofCardExporter` — jangan pindahkan ke import statis.
- **Jangan memakai emoji Unicode** di antarmuka. Ikon memakai `@lucide/svelte`
  dengan `aria-hidden="true"` dan label teks di sebelahnya.

---

## 10. Peta file

```
api/                 openapi.yaml (7 path, 11 skema) · dossier.schema.json
backend/
  cmd/server/        entrypoint, graceful shutdown
  internal/
    scoring/         satu-satunya tempat skor dihitung
    security/        ssrf · upload · ratelimit · headers
    forensics/       ELA server
    linkcheck/       heuristik URL
    llm/             klien 9router + mock
    grounding/       allowlist (hardcoded)
    pipeline/        errgroup, sub-deadline, degradasi
    token/           HMAC share token
    tactics/         verifikasi frasa
    demo/            demo-cases.json (embed)
    httpapi/         router, handler, SSE, error envelope
    models/          Dossier (cermin kontrak)
frontend/
  src/routes/        +page (beranda) · deteksi · hasil · edukasi · tentang
  src/lib/
    audit.svelte.ts  sesi audit + SSE + AbortController
    sse.ts           parser SSE berbasis fetch (bukan EventSource)
    api/schemas.ts   cermin Zod dari kontrak  ← sumber P0-1
    result.svelte.ts penyimpanan hasil in-memory
    mode.svelte.ts   Jelajah/Esensial
    a11y.svelte.ts   ukuran teks + read-aloud
    components/      ui · deteksi · hasil · mode · a11y · share
    forensics/       ela.worker.ts  ← belum tersambung
    data/            demo-cases · tactics · reply-templates
configs/             reply_templates.json
docs/                PROGRESS.md · DECISIONS.md
instruction/         PRD · TECH_SPEC · DESIGN (CLAUDE.md di-gitignore)
tools/               check-demo-diff.sh
```

---

## 11. Langkah berikutnya (urutan yang disarankan)

1. **Perbaiki P0-1** (null vs array) — tanpa ini tidak ada yang bisa didemokan.
2. **Perbaiki P0-2** (`goto` menggantikan reload) — hasil akhirnya baru muncul.
3. **Sambungkan `partial` → skeleton** (P0-3) — janji "bertahap" jadi benar.
4. **Sambungkan slider ELA + contoh demo** (P1-5) — ini fitur yang paling
   membedakan AMBA di depan juri.
5. **Mode Esensial: tambahkan aturan CSS `[data-mode]`** (P1-4).
6. **Hapus atau penuhi klaim PWA / e2e** (P1-6).
7. **Tulis ulang `docs/PROGRESS.md`** supaya tidak bertentangan dengan STATUS.md.

Urutan 1–2 kira-kira setengah hari. Tanpa keduanya, demo hanya bisa memakai
`/demo/{id}` lewat `curl`, bukan lewat antarmuka.
