# Deploy AMBA ke produksi

Arsitektur yang dipakai: **frontend SPA statis di Vercel**, **API Go di Fly.io**
(direkomendasikan, region Singapura) **atau Railway**. Browser memanggil API
lintas origin; API hanya menerima origin yang persis sama dengan
`PUBLIC_APP_URL` (ditambah `CORS_EXTRA_ORIGINS`).

```
Pengguna ──► Vercel (SPA, CSP, cache)  ──fetch/SSE──►  Fly.io / Railway (Go API)
                                                         ├─► LLM gateway (OpenAI-compatible)
Telegram ──webhook──────────────────────────────────────►├─► Google Fact Check Tools
                                                         └─► Google Safe Browsing
```

Urutannya: siapkan kunci → deploy API → deploy frontend → sambungkan keduanya →
uji. Semua perintah di bawah aman diulang.

---

## 1. Siapkan kunci

| Variabel | Wajib | Cara mendapatkan |
|---|---|---|
| `REPORT_SIGNING_KEY` | ya | `openssl rand -base64 48`. Mengganti kunci membatalkan semua tautan bagikan yang sudah beredar. |
| `NINEROUTER_API_KEY` | ya | Kunci OpenRouter (`sk-or-...`). Alamat gateway dan nama model sudah diisi di `fly.toml`. Lihat 1.1. |
| `GOOGLE_API_KEY` | sangat disarankan | Satu kunci Google Cloud untuk Fact Check Tools dan Safe Browsing. Lihat 1.2. |
| `SIGHTENGINE_API_USER`, `SIGHTENGINE_API_SECRET` | opsional | Pemeriksa khusus gambar AI untuk Jejak AI. Lihat 1.3. |
| `DOMAIN_AGE_LOOKUP` | tidak (bawaan `true`) | Umur domain tautan lewat RDAP, tanpa kunci. Isi `false` untuk mematikan. |
| `TELEGRAM_BOT_TOKEN`, `TELEGRAM_WEBHOOK_SECRET` | opsional | Lihat bagian 5. |

Simpan semua nilai ini sebagai *secret* di hosting, jangan di repo.

### 1.1 Model (LLM gateway)

Model harus menerima **gambar** (`image_url` berisi data URI) dan **keluaran
terstruktur** (`response_format: json_schema`).

**Bawaan repo: OpenRouter, model gratis `nvidia/nemotron-3.5-super-vl:free`.**
Model ini menerima gambar, mendukung `structured_outputs`, konteks 1M token,
dan tercatat sekitar 1,6 detik latensi di OpenRouter (November 2026).
`fly.toml` sudah mengisi:

```
NINEROUTER_BASE_URL = "https://openrouter.ai/api/v1"
NINEROUTER_MODEL    = "nvidia/nemotron-3.5-super-vl:free"
LLM_EXTRA_BODY      = '{"models":["nvidia/nemotron-3.5-super-vl:free","openrouter/free"],"reasoning":{"enabled":false}}'
```

- `models` adalah daftar cadangan OpenRouter. Kalau model utama sibuk atau
  kena batas, OpenRouter mencoba `openrouter/free`, router yang memilih model
  gratis lain yang mendukung gambar dan keluaran terstruktur. Isi daftar ini
  hanya dengan model gratis supaya tidak pernah tertagih.
- `reasoning.enabled=false` mematikan "berpikir panjang" model ini supaya
  jawabannya muat di anggaran 9 detik.
- Satu-satunya rahasia adalah `NINEROUTER_API_KEY` (kunci OpenRouter). Jangan
  pernah menaruhnya di repo: repo ini publik dan kunci yang bocor akan
  disalahgunakan. Kalau kunci pernah ditempel di chat atau dokumen, buat kunci
  baru di dashboard OpenRouter dan hapus yang lama.

Batas pemakaian model gratis OpenRouter (dokumentasi OpenRouter, 2026):
**20 permintaan per menit** dan **50 permintaan per hari**; naik menjadi
**1.000 per hari** setelah akun pernah membeli kredit minimal $10 (kreditnya
tidak terpakai oleh model gratis). Satu audit memakai sampai dua panggilan
model (analisis dan klasifikasi sumber), Balas Sopan satu lagi. Jadi tanpa
top-up, kapasitasnya kira-kira 20–25 audit per hari: cukup untuk mencoba,
kurang untuk dinilai juri atau dipakai umum.

Privasi: OpenRouter punya setelan privasi terpisah untuk model gratis. Kalau
akun melarang penyedia yang mencatat atau melatih model dengan prompt, permintaan
ke model gratis tertentu bisa ditolak (log `status=404`). Periksa kebijakan
data penyedia di halaman model sebelum memakainya untuk konten pengguna, dan
sesuaikan teks privasi di `/tentang` kalau penyedia menyimpan prompt.

Model gratis bisa ditarik kapan saja. Kalau log menunjukkan
`llm_model_listed=false`, pilih model gratis lain di openrouter.ai yang
mendukung input gambar dan structured outputs, lalu ganti `NINEROUTER_MODEL`
dan isi `models` di `LLM_EXTRA_BODY`.

**Alternatif:**

- **NineRouter milik tim**: isi `NINEROUTER_BASE_URL` dengan alamat publiknya
  (bukan `localhost`) dan kosongkan `LLM_EXTRA_BODY`.
- **Gemini langsung**: `NINEROUTER_BASE_URL=https://generativelanguage.googleapis.com/v1beta/openai`,
  kunci dari Google AI Studio, `LLM_EXTRA_BODY` kosong.

Nama model berubah dari waktu ke waktu, jadi cocokkan dengan daftar gateway:

```bash
curl -s "$NINEROUTER_BASE_URL/models" -H "Authorization: Bearer $NINEROUTER_API_KEY" | jq -r '.data[].id' | grep ':free'
```

Saat boot, log `readiness` menulis `llm_reachable`, `llm_model_listed`, dan
`llm_extra_fields`. Dua yang pertama harus `true`. Kalau salah satu `false`,
audit tetap berjalan tetapi pilar AI dan cek fakta ditandai tidak tersedia
(degraded).

### 1.2 Google Cloud (cek fakta dan reputasi tautan)

1. Buat atau pilih project di Google Cloud Console.
2. Aktifkan **Fact Check Tools API** dan **Safe Browsing API**.
3. Buat API key, lalu batasi key itu hanya untuk dua API tersebut.
4. Isi `GOOGLE_API_KEY`. Jika ingin kunci terpisah, `GOOGLE_FACTCHECK_KEY` dan
   `GOOGLE_SAFE_BROWSING_KEY` menimpa `GOOGLE_API_KEY` untuk layanannya masing-masing.

Catatan lisensi: Safe Browsing API v4 hanya untuk penggunaan **non-komersial**.
Jika AMBA nanti dipakai untuk tujuan komersial, Google mengarahkan ke Web Risk
API (belum diimplementasikan di AMBA).

Tanpa `GOOGLE_API_KEY`, pilar cek fakta tidak pernah bisa memverifikasi sumber
(grounding gagal tertutup) dan pemeriksaan tautan hanya memakai heuristik lokal.

### 1.3 Pemeriksa khusus gambar AI (opsional)

Jejak AI selalu membaca data asal-usul di dalam file (Content Credentials/C2PA,
IPTC/XMP, pengaturan Stable Diffusion/ComfyUI, EXIF) dan ciri yang terlihat
oleh model. Banyak gambar yang beredar lewat WhatsApp sudah kehilangan data
itu, jadi penilaian visual model menjadi satu-satunya petunjuk. Untuk penilaian
yang lebih kuat, aktifkan detektor khusus Sightengine (model `genai`):

1. Daftar di sightengine.com. Paket gratis saat ini 2.000 operasi per bulan,
   maksimal 500 per hari (halaman harga Sightengine, November 2026).
2. Isi `SIGHTENGINE_API_USER` dan `SIGHTENGINE_API_SECRET` sebagai *secret*.
   Isi keduanya atau tidak sama sekali; server menolak boot jika hanya satu.
3. Log `readiness` menampilkan `ai_detector=true`.

Setiap gambar yang diunggah, dan gambar pratinjau tautan, lalu dikirim juga ke
Sightengine. Halaman Tentang sudah menyebut ini. Jika kuota habis atau layanan
gagal, kartu Jejak AI tidak menampilkan baris detektor dan hasil tetap
dihitung dari bukti lain.

---

## 2. API di Fly.io (direkomendasikan)

`backend/fly.toml` sudah berisi semua setelan non-rahasia: `ENVIRONMENT=production`,
`LLM_MOCK=false`, `TRUSTED_PROXY_CIDRS=*`, `CLIENT_IP_HEADER=Fly-Client-IP`,
satu mesin selalu hidup (cold start akan memakan anggaran audit 12 detik), dan
health check `/healthz`.

```bash
cd backend
# 1. Ganti `app = "amba-api"` di fly.toml dengan nama aplikasi Anda, lalu:
fly launch --no-deploy --copy-config

# 2. Rahasia. PUBLIC_APP_URL = origin frontend di Vercel (https, tanpa garis miring di akhir).
#    Alamat gateway dan model OpenRouter sudah ada di fly.toml; cukup kuncinya.
fly secrets set \
  REPORT_SIGNING_KEY="$(openssl rand -base64 48)" \
  NINEROUTER_API_KEY="sk-or-..." \
  GOOGLE_API_KEY="..." \
  PUBLIC_APP_URL="https://amba.vercel.app"

# 3. Deploy dan periksa.
fly deploy
curl -s https://<nama-app>.fly.dev/healthz
fly logs | grep readiness
```

Origin tambahan (domain kustom di samping `*.vercel.app`):
`fly secrets set CORS_EXTRA_ORIGINS="https://amba.example.id"`.

---

## 3. API di Railway (alternatif)

1. New Project → Deploy from GitHub repo → pilih repo ini.
2. Settings → **Root Directory**: `/backend`.
3. Settings → **Config file path**: `/backend/railway.json`. Railway tidak
   mencari file konfigurasi di Root Directory, jadi path ini harus ditulis
   lengkap dari akar repo.
4. Variables:

   ```
   ENVIRONMENT=production
   LLM_MOCK=false
   TRUSTED_PROXY_CIDRS=*
   CLIENT_IP_HEADER=X-Real-IP
   PUBLIC_APP_URL=https://amba.vercel.app
   REPORT_SIGNING_KEY=...
   NINEROUTER_BASE_URL=https://openrouter.ai/api/v1
   NINEROUTER_API_KEY=sk-or-...
   NINEROUTER_MODEL=nvidia/nemotron-3.5-super-vl:free
   LLM_EXTRA_BODY={"models":["nvidia/nemotron-3.5-super-vl:free","openrouter/free"],"reasoning":{"enabled":false}}
   GOOGLE_API_KEY=...
   ```

   Jangan isi `SERVER_PORT`: Railway menyuntikkan `PORT` dan server memakainya.
   Edge Railway menimpa `X-Real-IP` dengan alamat pengguna, sehingga rate limit
   berlaku per pengguna, bukan satu ember untuk semua orang.
5. Settings → Networking → **Generate Domain**. Alamatnya
   (`https://<nama>.up.railway.app`) dipakai di langkah 4.

---

## 4. Frontend di Vercel

1. Add New Project → import repo → **Root Directory**: `frontend`.
   `frontend/vercel.json` sudah mengatur build (`npm run build`), output
   `build/`, fallback SPA, header keamanan (HSTS, COOP, `frame-ancestors`), dan cache.
2. Environment Variables (scope **Production**):

   ```
   VITE_API_BASE_URL=https://<nama-app>.fly.dev/api/v1
   VITE_API_MODE=http
   ```

   Variabel `VITE_*` dibaca saat **build**. CSP `connect-src` juga dibuat dari
   `VITE_API_BASE_URL` saat build, jadi setiap kali alamat API berubah, jalankan
   **Redeploy**.
3. Untuk scope **Preview**, isi `VITE_API_MODE=mock`. URL preview Vercel berubah
   setiap deploy dan tidak ada di daftar CORS API, jadi preview memakai data
   contoh yang berlabel jelas.
4. Setelah deploy pertama, pastikan `PUBLIC_APP_URL` di API sama persis dengan
   origin produksi Vercel (atau domain kustom Anda), lalu deploy ulang API bila diubah.

---

## 5. Bot Telegram (opsional)

1. Buat bot lewat @BotFather, simpan token-nya.
2. Buat secret webhook (hanya `A-Z a-z 0-9 _ -`, minimal 16 karakter):
   `openssl rand -hex 24`.
3. Set di API:

   ```bash
   fly secrets set \
     TELEGRAM_BOT_TOKEN="..." \
     TELEGRAM_WEBHOOK_SECRET="..." \
     TELEGRAM_WEBHOOK_URL="https://<nama-app>.fly.dev/telegram/webhook"
   ```

   `TELEGRAM_MODE` sudah `webhook` di `fly.toml`. Saat boot server memanggil
   `setWebhook` dengan secret tersebut; Telegram lalu mengirim header
   `X-Telegram-Bot-Api-Secret-Token` dan permintaan tanpa header yang cocok ditolak.
4. Periksa: `curl -s "https://api.telegram.org/bot<TOKEN>/getWebhookInfo"`.
5. Coba di Telegram: `/start`, `/cek https://contoh.tautan`, kirim tautan
   langsung, atau kirim foto/tangkapan layar. Teks tanpa tautan dijawab bahwa
   AMBA hanya memeriksa tautan dan gambar.

Untuk menjalankan bot di laptop (tanpa URL https publik), pakai
`TELEGRAM_MODE=polling`. Mode polling menghapus webhook yang terdaftar, jadi
jangan menjalankan polling lokal dengan token bot produksi.

---

## 6. Uji setelah deploy

```bash
API=https://<nama-app>.fly.dev
APP=https://amba.vercel.app

# API hidup dan CORS menerima origin aplikasi
curl -s $API/healthz
curl -s -o /dev/null -w "%{http_code}\n" -X OPTIONS $API/api/v1/audit/url \
  -H "Origin: $APP" -H "Access-Control-Request-Method: POST"

# Audit tautan, stream SSE: stage → partial → verdict → done
curl -N -X POST $API/api/v1/audit/url \
  -H "Content-Type: application/json" -H "Origin: $APP" -H "Accept: text/event-stream" \
  -d '{"url":"https://bri-verifikasi.my.id/login"}'

# Audit gambar (JPG/PNG/WebP, maksimal 10 MB)
curl -N -X POST $API/api/v1/audit/file \
  -H "Origin: $APP" -H "Accept: text/event-stream" -F "file=@foto.jpg"

# Header keamanan frontend
curl -sI $APP | grep -i -E "strict-transport|x-content-type|referrer-policy|content-security"

# Lighthouse terhadap produksi (dari folder frontend)
cd frontend && LH_URL=$APP npm run lighthouse
```

Lalu coba manual di ponsel: periksa satu tautan, unggah satu tangkapan layar,
buka tautan bagikan di perangkat lain, dan bagikan dari menu Share Android ke AMBA.

### 6.1 Kalibrasi deteksi gambar dengan model sungguhan

Ambang Jejak AI (`AI_LIKELY` mulai skor 80) ditetapkan tanpa kunci model. Setelah
kunci terpasang, jalankan pipeline yang sama pada gambar berlabel, satu folder
per label. Pakai gambar **tanpa** metadata (misalnya yang sudah lewat WhatsApp),
karena gambar dengan metadata sudah terbaca tanpa model:

```bash
cd backend
LLM_MOCK=false NINEROUTER_API_KEY=... NINEROUTER_MODEL=nvidia/nemotron-3.5-super-vl:free \
  NINEROUTER_BASE_URL=https://openrouter.ai/api/v1 \
  go run ./cmd/evalimages ./eval/ai ./eval/asli ./eval/teks-ditempel > hasil.tsv
```

Setiap baris berisi skor AI, asal-usul, skor detektor, bukti teks ditempel,
status, dan floor. Ringkasan per folder muncul di stderr. Untuk membandingkan
beberapa model gratis sekaligus dan menghitung AUC serta alarm palsu per
ambang, pakai `tools/eval/compare.py` (lihat `tools/eval/README.md`). Jika foto asli sering
mencapai `AI_LIKELY`, naikkan `scoring.AILikelyMin`; jika gambar AI jarang
mencapainya, turunkan, lalu catat angkanya di `docs/DECISIONS.md`. Benchmark
teks ditempel bisa diulang dengan `tools/textbench/gen.py` (lihat isinya).

---

## 7. Operasional

- **Log** berformat JSON dan tidak pernah memuat IP, URL, teks, judul, nama
  file, maupun kunci. Kegagalan layanan luar hanya mencatat alasan dan kode HTTP,
  misalnya `fact check search call failed reason=http_status status=403`.
- **Metrik** (penghitung anonim, opsional): set `METRICS_ADDR=0.0.0.0:9091` dan
  aktifkan blok `[metrics]` yang sudah disiapkan di `fly.toml`. Port metrik tidak publik.
  Selain jumlah audit per status, ada hitungan per kode aturan:
  `amba_floors_total{floor}` dan `amba_link_reasons_total{reason}`. Isinya hanya kode, tidak pernah teks atau
  tautan. Pakai hitungan ini untuk melihat aturan mana yang sering memicu alarm
  palsu (laporan pengguna) dan aturan mana yang tidak pernah terpakai.
- **Rate limit**: 15 permintaan per menit per pengguna (burst 5) dan 32 audit
  bersamaan per server (TECH_SPEC §7.3).
- **Skala**: satu mesin `shared-cpu-1x` 512 MB cukup untuk demo. Untuk beban
  lebih besar, naikkan `min_machines_running`. Cache hasil dan rate limit
  disimpan per mesin, sesuai aturan tanpa database.

## 8. Masalah umum

| Gejala | Penyebab dan solusi |
|---|---|
| Boot gagal: `LLM_MOCK: must be false in production` | Model tiruan mengarang temuan, jadi produksi menolaknya. Set `LLM_MOCK=false` dan variabel `NINEROUTER_*`. |
| Boot gagal: `PUBLIC_APP_URL: must be an https:// origin` | Isi origin frontend dengan `https://`. |
| Browser: error CORS | `PUBLIC_APP_URL` atau `CORS_EXTRA_ORIGINS` harus sama persis dengan origin (skema + host, tanpa `/` di akhir). |
| Browser: `Refused to connect` (CSP) | `VITE_API_BASE_URL` salah saat build. Perbaiki di Vercel lalu Redeploy. |
| Semua pengguna kena rate limit bersamaan | `TRUSTED_PROXY_CIDRS` atau `CLIENT_IP_HEADER` belum diisi. |
| Log `llm gateway call failed ... status=400` | Model atau gateway tidak mendukung `json_schema` atau gambar. Pilih model lain. |
| Log `status=401` atau `status=403` | Kunci salah, atau API belum diaktifkan untuk kunci tersebut. |
| Log `llm gateway ... status=404` | OpenRouter: model gratis ditarik, atau setelan privasi akun untuk model gratis menolak penyedianya. |
| Log `status=429` | Kuota habis. OpenRouter gratis: 50 permintaan per hari tanpa top-up, 1.000 per hari setelah top-up $10. |
| Hasil selalu "Belum cukup data" | Periksa log `readiness`: kemungkinan model atau `GOOGLE_API_KEY` belum siap. |
