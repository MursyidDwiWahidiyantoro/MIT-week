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
| `NINEROUTER_BASE_URL`, `NINEROUTER_API_KEY`, `NINEROUTER_MODEL` | ya | Gateway model apa pun yang kompatibel OpenAI. Lihat 1.1. |
| `GOOGLE_API_KEY` | sangat disarankan | Satu kunci Google Cloud untuk Fact Check Tools dan Safe Browsing. Lihat 1.2. |
| `TELEGRAM_BOT_TOKEN`, `TELEGRAM_WEBHOOK_SECRET` | opsional | Lihat bagian 5. |

Simpan semua nilai ini sebagai *secret* di hosting, jangan di repo.

### 1.1 Model (LLM gateway)

Model harus menerima **gambar** (`image_url` berisi data URI) dan **keluaran
terstruktur** (`response_format: json_schema`). Dua pilihan:

- **NineRouter milik tim**: isi `NINEROUTER_BASE_URL` dengan alamat publiknya
  (bukan `localhost`, karena server di cloud tidak bisa menjangkau laptop).
- **Gemini langsung**: buat API key di Google AI Studio, lalu
  `NINEROUTER_BASE_URL=https://generativelanguage.googleapis.com/v1beta/openai`
  dan `NINEROUTER_API_KEY=<kunci Gemini>`.

Nama model berubah dari waktu ke waktu, jadi ambil dari daftar gateway sendiri:

```bash
curl -s "$NINEROUTER_BASE_URL/models" -H "Authorization: Bearer $NINEROUTER_API_KEY" | jq -r '.data[].id'
```

Saat boot, log `readiness` menulis `llm_reachable` dan `llm_model_listed`.
Keduanya harus `true`. Kalau salah satu `false`, audit tetap berjalan tetapi
pilar AI dan cek fakta ditandai tidak tersedia (degraded).

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
fly secrets set \
  REPORT_SIGNING_KEY="$(openssl rand -base64 48)" \
  NINEROUTER_BASE_URL="https://..." \
  NINEROUTER_API_KEY="..." \
  NINEROUTER_MODEL="..." \
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
   NINEROUTER_BASE_URL=...
   NINEROUTER_API_KEY=...
   NINEROUTER_MODEL=...
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
5. Coba di Telegram: `/start`, `/cek https://contoh.tautan`, kirim teks yang
   diteruskan, atau kirim foto tangkapan layar.

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
curl -s -o /dev/null -w "%{http_code}\n" -X OPTIONS $API/api/v1/audit/text \
  -H "Origin: $APP" -H "Access-Control-Request-Method: POST"

# Audit teks, stream SSE: stage → partial → verdict → done
curl -N -X POST $API/api/v1/audit/text \
  -H "Content-Type: application/json" -H "Origin: $APP" -H "Accept: text/event-stream" \
  -d '{"text":"Bantuan Rp 500.000 cair hari ini, klik tautan dan isi data KTP"}'

# Header keamanan frontend
curl -sI $APP | grep -i -E "strict-transport|x-content-type|referrer-policy|content-security"

# Lighthouse terhadap produksi (dari folder frontend)
cd frontend && LH_URL=$APP npm run lighthouse
```

Lalu coba manual di ponsel: periksa satu tautan, unggah satu tangkapan layar,
buka tautan bagikan di perangkat lain, dan bagikan dari menu Share Android ke AMBA.

---

## 7. Operasional

- **Log** berformat JSON dan tidak pernah memuat IP, URL, teks, judul, nama
  file, maupun kunci. Kegagalan layanan luar hanya mencatat alasan dan kode HTTP,
  misalnya `fact check search call failed reason=http_status status=403`.
- **Metrik** (penghitung anonim, opsional): set `METRICS_ADDR=0.0.0.0:9091` dan
  aktifkan blok `[metrics]` yang sudah disiapkan di `fly.toml`. Port metrik tidak publik.
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
| Log `status=429` | Kuota habis. Naikkan kuota atau tunggu. |
| Hasil selalu "Belum cukup data" | Periksa log `readiness`: kemungkinan model atau `GOOGLE_API_KEY` belum siap. |
