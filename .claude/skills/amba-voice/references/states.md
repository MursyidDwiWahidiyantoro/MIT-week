# State copy — rules and one example each

Load when writing any non-verdict state: loading, offline, error, empty, degraded.

| State | Rule | Example |
|---|---|---|
| Offline | Say what still works; offer a local action; never blame the network. | "Kamu sedang offline. Contoh kasus masih bisa dibuka. Periksa lagi saat koneksi kembali." |
| 429 (rate limited) | Show the countdown from the response; state the wait, not an error code. | "Terlalu banyak pemeriksaan. Coba lagi dalam 30 detik." |
| Degraded verdict (`degraded: true`) | Say which parts could not be checked; keep the verdict and its score visible; never present as complete. | "Sebagian pemeriksaan tidak selesai. Hasil ini masih indikatif dan bisa kurang lengkap." |
| Unsupported file | Name the accepted types and the limit; no blame. | "Format file ini belum didukung. Gunakan JPG, PNG, atau WebP maksimal 10 MB." |
| No checkable claim (`INSUFFICIENT`) | Say there is nothing to verify; suggest adding text or a link; never imply SAFE. | "Belum ada klaim yang bisa diperiksa dari isi ini. Tambahkan teks atau tautan yang ingin dicek." |
| Platform-blocked link | Exact spec string + CTA. | "Platform tidak mengizinkan pembacaan otomatis" + "Unggah tangkapan layar" |
| Invalid share token | Exact spec string; no partial result. | "Laporan tidak sah" |
| Valid share token | Exact spec string. | "Terverifikasi AMBA" |
| Reloaded live result (nothing persisted) | State the loss and give one-tap recovery. | "Hasil tidak tersimpan — periksa lagi" |
| Unknown `schema_version` | Ask for an update, no technical detail. | "Perbarui aplikasi" |
| ELA on PNG/screenshot | State reduced reliability; do not hide the result. | "ELA kurang andal pada PNG/tangkapan layar" |
| Loading / stage progress | Name the step, no fake percentage. | "Memeriksa tautan…" |
| Demo content | Exact watermark + banner. | "CONTOH / SIMULASI" · "Ini data contoh, bukan hasil audit sungguhan" |
| Generic error envelope | Use the server `message`; add one next step; never show codes. | "Tautan tidak valid." |
| Honest limits | Exact prefixes from `verdict.basis` / `not_checked`. | "Berdasarkan: …" / "Tidak dapat diperiksa: …" |
| AI pillar caveat | Exact string. | "Indikatif, bukan bukti pasti" |

## Hard rules for all states

- No error codes, stack traces, or internal rule names (BLOCKED_TARGET stays generic).
- Every state names one next action or says why none is possible.
- Never say "palsu", "aman", "pasti", "terbukti", "100%".
- No English strings.
- Loading copy must not claim progress that is not real.
