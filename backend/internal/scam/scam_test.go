package scam

import (
	"strings"
	"testing"

	"github.com/amba/amba/backend/internal/tactics"
)

// scams are messages of the modus documented by Indonesian banks, operators,
// the police and news outlets (2023-2026): the quoted OCBC prize SMS, the
// APK "undangan pernikahan", "foto paket" and "surat tilang" files, the
// "kode pembatalan transaksi" call script described by Bank Mandiri, "mama
// minta pulsa", wrong-transfer refunds, paid like-and-subscribe jobs,
// guaranteed-profit investments and loans with an up-front fee.
var scams = []struct{ text, want string }{
	{"SELAMAT!!! No. Whatsapp Anda terpilih memenangkan undian berhadiah Rp250 Juta dari PT xxx. Call 081122223333 atau klik link.ini hadiahuang45.blogspot.com.", PrizeFee},
	{"Assalamualaikum, mohon maaf mengganggu. Ini undangan pernikahan kami, dibuka ya. Surat Undangan Pernikahan.apk", APKLure},
	{"Permisi kak, saya kurir J&T, ini foto paketnya ya kak, dicek dulu alamatnya. Lihat Foto Paket.apk", APKLure},
	{"Surat Tilang.apk Pemberitahuan tilang elektronik, silakan buka untuk melihat bukti pelanggaran.", APKLure},
	{"Kak tolong install aplikasi tagihan ini dulu ya, buat cek resi pengiriman.", APKLure},
	{"Selamat siang Bapak, kami dari pihak bank Mandiri. Ada transaksi mencurigakan di rekening Bapak. Untuk pembatalan, kode pembatalan transaksi yang masuk ke HP Bapak tolong disebutkan.", CredentialRequest},
	{"Halo kak, tadi ada kode OTP masuk ke nomor kakak, salah kirim. Tolong kirimkan kode OTP itu ke saya ya kak, penting.", CredentialRequest},
	{"Mohon kirimkan kode verifikasi 6 digit yang baru saja Anda terima untuk aktivasi akun WhatsApp Anda.", CredentialRequest},
	{"Pelanggan Yth, untuk upgrade kartu 4G Anda, balas SMS ini dengan PIN yang Anda terima.", CredentialRequest},
	{"Kartu kredit Anda terblokir. Segera isi nomor kartu, masa berlaku, dan CVV di link berikut untuk aktivasi kembali.", CredentialRequest},
	{"Halo, maaf kode 6 digit yang dikirim ke nomor kamu tadi salah kirim, bisa tolong diteruskan ke aku?", CredentialRequest},
	{"Jika Bapak ingin membatalkan transaksi tersebut, sebutkan kode OTP yang barusan masuk.", CredentialRequest},
	{"Ma, ini nomor baru aku. Tolong kirimin pulsa 100rb dulu ya, nanti aku ganti. Jangan bilang papa dulu.", MoneyPressure},
	{"Pa ini nomor baru mama, hp mama hilang. Tolong transfer 2 juta ke rekening ini sekarang ya, penting.", MoneyPressure},
	{"Kak, aku lagi di kantor polisi, butuh uang tebusan 5 juta segera, transfer sekarang, jangan bilang mama.", MoneyPressure},
	{"Bapak/Ibu, kami dari BPJS Kesehatan. Iuran Anda menunggak, segera bayar denda Rp 1.200.000 agar kartu tidak dinonaktifkan.", MoneyPressure},
	{"Paket Anda tertahan di gudang karena alamat kurang lengkap. Bayar biaya pengiriman ulang Rp 15.000, klik jnt-ekspres-id.top", MoneyPressure},
	{"Maaf pak, saya salah transfer Rp 1.500.000 ke rekening bapak. Tolong dikembalikan ya pak, itu uang untuk berobat.", WrongTransfer},
	{"Halo! Kami menawarkan kerja paruh waktu dari rumah. Cukup like dan subscribe video YouTube, dibayar Rp 30.000 per tugas. Penghasilan bisa Rp 500rb per hari!", JobScam},
	{"Lowongan kerja online freelance, gaji Rp 150.000 - Rp 1.000.000 per hari, hanya perlu follow akun TikTok. Minat? Balas YA.", JobScam},
	{"Investasi emas digital, profit 5% per hari, modal kecil untung besar!", InvestmentScam},
	{"Robot trading pasti untung, join sekarang sebelum kuota penuh.", InvestmentScam},
	{"Join grup sinyal saham, profit 10% per minggu dijamin!", InvestmentScam},
	{"Pinjaman dana tunai tanpa jaminan, langsung cair 50 juta. Syarat cukup KTP, biaya admin Rp 350.000 dibayar di awal.", LoanFee},
	{"Selamat! Nomor Anda terpilih sebagai pemenang undian Shopee Rp 100 juta. Untuk pencairan, transfer biaya administrasi Rp 2.500.000 ke rekening berikut.", PrizeFee},
	{"BLT Rp 600.000 sudah cair untuk Anda! Klaim bantuan sekarang di bansos-kemensos.info sebelum hangus hari ini.", PrizeFee},
	{"Dana kaget Rp 250.000 untuk kamu! Ambil sekarang di link.dana.id-kaget.my.id", PrizeFee},
	{"Selamat kamu menang iPhone 15 dari giveaway kami! Untuk klaim hadiah, transfer ongkir Rp 250.000 dulu ya.", PrizeFee},
	{"Nomor WhatsApp kamu terpilih dapat saldo DANA gratis Rp 1.000.000! Klik link ini untuk klaim: dana-gratis.my.id", PrizeFee},
}

// legit are ordinary messages and genuine warnings, several of them quoted
// from Telkomsel's and Bank Mandiri's own notices, that must never trigger
// the floor.
var legit = []string{
	"Telkomsel tidak pernah meminta password atau kode verifikasi kepada pelanggan untuk alasan apapun.",
	"Jangan Pernah Berikan Kode OTP dan Data Rahasia Anda kepada Siapa Pun.",
	"Jika ada pihak yang tidak dikenal menghubungi Anda dan meminta kode tersebut, sudah pasti itu adalah penipuan.",
	"Kode OTP Anda 482913. JANGAN BERIKAN kode ini kepada siapa pun, termasuk pihak bank. Berlaku 5 menit.",
	"Cukup kamu yang tahu, jangan bagikan kode OTP-mu ke siapa pun.",
	"Waspada! Beredar file Surat Undangan Pernikahan.apk di WhatsApp yang bisa menguras rekening. Jangan instal aplikasi dari pengirim tidak dikenal.",
	"Polisi mengimbau masyarakat waspada modus penipuan kurir paket yang mengirim file APK.",
	"Pelaku mengirim pesan \"Kirimkan kode OTP Anda untuk pembatalan transaksi\" kepada korban.",
	"Selamat! Kamu dapat voucher gratis ongkir untuk belanja hari ini. Cek di aplikasi Shopee.",
	"Nikmati cashback 20% untuk transaksi QRIS pakai BRImo. Syarat dan ketentuan berlaku.",
	"Ma, nanti aku transfer uang sekolah adik ya, sore ini.",
	"Pa, aku udah sampai kos. Besok aku kabari lagi.",
	"Tolong kirim laporan bulanan ke email saya sebelum jam 5 sore ya.",
	"Tagihan listrik bulan ini Rp 350.000, jatuh tempo tanggal 20. Bayar lewat aplikasi PLN Mobile.",
	"Lowongan kerja staf administrasi, penempatan Samarinda. Kirim CV ke hrd@perusahaan.co.id paling lambat 30 November.",
	"Bunga deposito 3,5% per tahun, dijamin LPS sesuai ketentuan.",
	"KTA bunga mulai 0,99% per bulan. Ajukan di aplikasi resmi bank.",
	"Kemensos memastikan bansos PKH tahap 4 cair mulai pekan depan melalui Kantor Pos dan bank Himbara.",
	"BLT Rp 600 ribu cair bulan ini, cek status penerima di cekbansos.kemensos.go.id.",
	"Bro, transfer dulu ya buat patungan futsal, 50rb ke rekening aku.",
	"Jangan bilang siapa-siapa ya, aku mau kasih kejutan ulang tahun buat Ibu. Transfer patungan kadonya nanti malam aja.",
	"Kode verifikasi WhatsApp Anda: 123-456. Jangan bagikan kode ini.",
	"Masukkan kode OTP yang dikirim ke nomor Anda untuk melanjutkan.",
	"Selamat pagi Bapak/Ibu, mohon segera membayar iuran kas RT bulan ini ke bendahara.",
	"Pesanan Anda sedang dikirim. Lacak paket di aplikasi J&T Express.",
	"Undangan pernikahan kami bisa dilihat di https://undangan-digital.com/rina-budi",
	"Pemenang giveaway minggu ini sudah diumumkan di akun Instagram resmi kami.",
	"Halo kak, ini nomor baru aku ya, yang lama hilang. Simpan ya.",
	"Saya dari kurir JNE, paket kakak sudah di depan rumah ya.",
	"Kalau ada yang minta kode OTP, langsung blokir saja nomornya ya Ma.",
	"Ingat ya, OTP itu rahasia, jangan dikasih ke siapa pun.",
	"Harga beras di pasar naik Rp 2.000 per kilogram minggu ini.",
	"Pemerintah menaikkan gaji ASN mulai Januari, kata Menteri Keuangan dalam konferensi pers.",
	"Rapat RT diundur ke hari Minggu jam 7 malam, mohon kehadirannya.",
	"Kerja dari rumah hari ini ya, nanti meeting jam 2 lewat Zoom.",
}

func TestDocumentedScamsAreCaught(t *testing.T) {
	caught := 0
	for _, c := range scams {
		got := Analyze(c.text).Pattern
		if got != c.want {
			t.Errorf("%q\n  want %q, got %q", c.text, c.want, got)
		}
		if got != "" {
			caught++
		}
	}
	t.Logf("scam patterns: %d of %d documented scam messages", caught, len(scams))
}

func TestLegitimateMessagesAreNotFlagged(t *testing.T) {
	for _, text := range legit {
		if r := Analyze(text); r.Strong() {
			t.Errorf("false alarm %s on %q (hits %+v)", r.Pattern, text, r.Hits)
		}
	}
}

// Every phrase handed to the highlighter must survive tactics.Verify, which
// only keeps text that really occurs in the message.
func TestTacticsArePhrasesFromTheText(t *testing.T) {
	for _, c := range scams {
		r := Analyze(c.text)
		cands := r.Tactics("text")
		kept := tactics.Verify(cands, c.text)
		if len(kept) != min(len(cands), tactics.MaxTactics) && len(cands) <= tactics.MaxTactics {
			t.Errorf("%q: %d candidates, %d verified (%+v)", c.text, len(cands), len(kept), cands)
		}
		for _, k := range kept {
			if !strings.Contains(strings.ToLower(strings.Join(strings.Fields(c.text), " ")), strings.ToLower(k.Phrase)) {
				t.Errorf("phrase %q is not in %q", k.Phrase, c.text)
			}
		}
	}
}

func TestEmptyText(t *testing.T) {
	if r := Analyze("   "); r.Strong() || len(r.Hits) != 0 {
		t.Fatalf("empty text: %+v", r)
	}
}

// holdoutScams and holdoutLegit were written after the rules above, to
// measure how the rules generalise. The first pass caught 13 of 20 scams; the
// misses led to general rules (a registration fee in a job offer, "ini
// perawat", "pinjemin", card number with CVV), after which a second, fresh
// set caught 13 of 15 with no false alarm. Both sets now guard regressions.
var holdoutScams = []string{
	"Bpk/Ibu yth, kami CS BRI. Rekening anda terdeteksi transaksi janggal. Utk membatalkan, sebutkan 6 digit kode yg kami kirim via SMS.",
	"Halo sayang, ini mama pakai nomor teman. Hp mama rusak. Bisa tolong tf 500rb ke rek ini? Penting banget.",
	"SELAMAT! Akun Shopee Anda mendapatkan hadiah undian 2025 berupa uang tunai Rp 175.000.000. Info lengkap hub: 0852-1234-5678",
	"Pekerjaan sampingan mudah! Cukup kasih like di postingan Instagram, dapat komisi 15rb/like. Gabung grup Telegram kami.",
	"Kartu ATM anda diblokir sementara. Hubungi petugas kami di 0812-3456-7890 untuk membuka blokir & siapkan nomor PIN anda.",
	"Mas, aku salah kirim saldo DANA 200rb ke nomor mas, tolong balikin ya ke 085712345678",
	"Dapatkan profit 30% dalam 7 hari! Investasi crypto aman terpercaya, minimal deposit 100rb.",
	"Kak ini undangan nikahan aku ya, Undangan Pernikahan.pdf.apk dibuka aja kak",
	"Kakak dapat paket bantuan sembako gratis dari pemerintah, isi data diri di bantuan-sembako.site ya, kuota terbatas.",
	"Yth Nasabah, poin reward BNI Anda akan hangus hari ini. Tukarkan sekarang dgn memasukkan nomor kartu & CVV di bni-poin.xyz",
	"Om, ini Dika keponakan om, nomor lama hilang. Om bisa pinjemin uang 3 juta dulu? Aku lagi kena musibah, jangan bilang tante ya.",
	"Info loker: admin input data via HP, kerja 2 jam sehari, gaji 3jt/minggu. Tidak dipungut biaya kecuali biaya pendaftaran 150rb.",
	"Anda terpilih menerima subsidi listrik PLN Rp 1.500.000. Segera konfirmasi dengan mengisi formulir di pln-subsidi.online",
	"Kami dari kantor pajak. Ada tunggakan pajak kendaraan Anda, bayar denda Rp 2.300.000 hari ini juga agar tidak diproses hukum.",
	"Halo kak, aku admin olshop. Mau kerja tambahan? Tugasnya cuma rating produk bintang 5, dibayar 20rb per produk.",
	"Bu, ini perawat RS Harapan. Anak ibu kecelakaan dan harus segera dioperasi, mohon transfer biaya operasi 15 juta sekarang.",
	"Kode OTP sudah kami kirim ke HP bapak, mohon dibacakan agar proses pemblokiran kartu bisa kami batalkan.",
	"Paket COD kakak gagal dikirim, silakan unduh aplikasi J&T Tracking berikut untuk jadwal ulang: jnt-tracking.apk",
	"Gabung arisan online, setor 100rb dapat 1 juta dalam seminggu! Dijamin aman tanpa risiko.",
	"WA kamu akan diblokir permanen karena melanggar aturan. Kirim kode verifikasi yang masuk sekarang supaya akun aman.",
	"Pemberitahuan: nomor Anda mendapat hadiah mobil dari undian Telkomsel. Hubungi 0811-2233-4455 untuk proses klaim.",
	"Sist, aku dari tim HRD Lazada, kami buka kerja part time di rumah. Komisi 50rb tiap selesai tugas. Minat chat aku ya.",
	"Kak, sy salah transfer 750rb ke rek kakak, mohon bantuannya dikembalikan ke rekening saya ya kak, uangnya buat bayar sekolah anak.",
	"Pak ini nomor baru saya Budi atasan Bapak. Tolong transferkan dulu 5 juta ke rekening vendor ini, urgent, nanti saya ganti.",
	"Akun Livin' Anda terindikasi diretas. Segera kirimkan kode OTP yang Anda terima kepada petugas kami agar akun tidak dibekukan.",
	"Resi pengiriman paket Anda: Cek_Resi_JNE.apk silakan diinstal untuk melacak paket.",
	"Selamat, Anda berhak mendapat BLT UMKM Rp 2.400.000. Pencairan dikenakan biaya materai Rp 300.000, transfer ke rekening petugas.",
	"Raih keuntungan 2% per hari dengan deposit minimal 500rb, withdraw kapan saja!",
	"Pinjol legal OJK, langsung cair tanpa BI checking. Cukup bayar biaya asuransi 250rb di awal.",
	"Dek, ini kakak pakai nomor temen, hp kakak dicuri. Tolong isiin pulsa 50rb ke nomor ini ya.",
	"Halo, kami dari Shopee. Akun Anda terpilih mendapatkan saldo ShopeePay gratis 1 juta. Klaim sekarang di shopeepay-gift.click",
	"Mohon foto kartu ATM bagian depan dan belakang serta sebutkan PIN untuk proses verifikasi data nasabah.",
	"Tugas harian: tonton video 30 detik, setiap video dibayar Rp 5.000, penghasilan hingga 400rb per hari.",
	"Ibu, saya polisi dari Polres Jakarta Selatan. Anak Ibu ditahan karena narkoba, siapkan uang damai 20 juta dan transfer secepatnya.",
	"Pesan dari PLN: tagihan listrik Anda menunggak, listrik akan diputus hari ini. Bayar via link pln-tagihan.top",
}

var holdoutLegit = []string{
	"Bank BRI tidak pernah meminta kode OTP, PIN, maupun password melalui telepon, SMS, atau WhatsApp.",
	"Mohon maaf, rapat koordinasi besok dimundurkan ke jam 10 pagi di ruang rapat lantai 3.",
	"Kak, tolong transfer uang kos bulan ini ke rekening Bu Siti ya, sudah tanggal 5.",
	"Selamat ulang tahun ya Nak, semoga panjang umur dan sehat selalu. Hadiahnya Mama kirim lewat paket.",
	"Promo Gajian! Diskon hingga 50% untuk semua produk di Tokopedia, berlaku sampai 27 November.",
	"Pengumuman: pembagian bansos beras 10 kg di kantor kelurahan hari Sabtu, bawa KTP dan KK asli.",
	"Saya sudah transfer Rp 250.000 untuk iuran arisan bulan ini ya Bu, mohon dicek.",
	"Untuk keamanan, kami tidak akan pernah meminta Anda menyebutkan kode OTP. Abaikan pesan yang mengatasnamakan kami.",
	"Hai, paketmu dengan resi JX123456 sudah tiba di gudang Samarinda dan akan dikirim besok.",
	"Investasi reksa dana memiliki risiko, kinerja masa lalu tidak menjamin kinerja masa depan.",
	"Nanti kalau sudah sampai rumah kabari ya, jangan lupa makan.",
	"Gaji karyawan bulan November akan ditransfer paling lambat tanggal 25.",
	"Tolong bacakan kode booking tiket pesawatnya ya, mau aku cek di aplikasi.",
	"Waspada penipuan! Jangan klik tautan mencurigakan yang mengatasnamakan bank, dan jangan instal file APK dari orang tidak dikenal.",
	"Buat yang mau ikut kerja bakti hari Minggu, kumpul di balai desa jam 7 pagi.",
	"Pak, saya izin tidak masuk kerja hari ini karena sakit, surat dokter menyusul.",
	"Dapatkan cashback hingga Rp 100.000 dengan pembayaran QRIS di merchant pilihan. S&K berlaku.",
	"Harga BBM nonsubsidi turun mulai 1 Desember, kata Pertamina dalam keterangan resminya.",
	"Mohon kirim foto KTP untuk keperluan pendaftaran BPJS anak di kantor desa ya Pak.",
	"Selamat! Anda lolos seleksi administrasi. Tes wawancara dilaksanakan Senin pukul 09.00 di kantor kami.",
	"Mandiri tidak pernah meminta nasabah menyebutkan PIN, OTP, atau password melalui saluran apa pun.",
	"Bapak/Ibu, iuran sampah bulan ini Rp 20.000, bisa transfer ke rekening bendahara RT atau bayar tunai.",
	"Lowongan Bidan RSUD, berkas dikirim ke rsud@kaltimprov.go.id, seleksi tidak dipungut biaya apa pun.",
	"Selamat ya atas wisudanya! Hadiahnya nanti aku kirim lewat ojol.",
	"Info dari sekolah: pembayaran SPP bisa melalui virtual account BNI paling lambat tanggal 10.",
	"Gojek: Pakai GoPay dapat diskon 30% untuk GoFood hari ini. S&K berlaku.",
	"Pak, besok saya cuti ya, pekerjaan sudah saya serahkan ke Rina.",
	"Ingat, petugas bank tidak akan pernah meminta kode OTP Anda. Kalau ada yang minta, itu penipuan.",
	"Kakak, transfer uang buat beli kado ayah ya, nanti aku yang belanja. Rahasia ya biar surprise!",
	"Paket kamu sudah dikirim, nomor resi JNE 123456789, bisa dicek di jne.co.id.",
	"Bunga KPR fixed 4,5% per tahun selama 3 tahun pertama, ajukan di kantor cabang terdekat.",
	"Mama, aku butuh uang buat bayar buku 150rb, besok transfer ya.",
	"Pemerintah resmi memperpanjang program diskon listrik 50% hingga Februari, kata Kementerian ESDM.",
	"Undangan: syukuran rumah baru kami Sabtu ini jam 4 sore, alamat terlampir.",
	"Klik tautan berikut untuk mengunduh laporan keuangan resmi perusahaan: https://www.bankmandiri.co.id/laporan-tahunan",
}

func TestHoldoutScams(t *testing.T) {
	for _, text := range holdoutScams {
		if !Analyze(text).Strong() {
			t.Errorf("missed scam: %q", text)
		}
	}
	for _, text := range holdoutLegit {
		if r := Analyze(text); r.Strong() {
			t.Errorf("false alarm %s on %q", r.Pattern, text)
		}
	}
}

// The sets below were each written before the rule changes they led to and
// measured once on first sight (November 2026):
//
//	set 3 (account threats, new fees, customs; fee and notice messages):
//	      0 of 15 scams, 5 false alarms in 25 ordinary messages. The rules
//	      had no account-threat pattern, and a payment with a threat or a
//	      registration or admin fee was enough for MONEY_PRESSURE.
//	set 4: 11 of 20 scams, no false alarm in 20.
//	set 5: 10 of 20 scams, no false alarm in 20.
//	set 6: 14 of 25 scams, 1 false alarm in 25 (a per-day stand job).
//
// So on messages the rules had not seen they caught about half of the scams
// (35 of 65) with 1 false alarm in 65 ordinary messages; the rest is left to
// the model and the link check. Everyday probes also found code requests
// that were not secrets (a location pin, a Wi-Fi or Zoom password) and a
// card-data warning read as a request. Every set now guards regressions;
// leftToTheModel keeps the messages still missed in sight.

var holdout3Scams = []string{
	"Nasabah Yth, akun BRImo Anda akan diblokir dalam 24 jam. Segera verifikasi data di https://brimo-verifikasi.my.id",
	"Rekening Anda terdeteksi aktivitas mencurigakan dan akan dibekukan. Hubungi CS kami di 0812-3456-7890",
	"Pelanggan BRI Yth, mulai hari ini tarif transfer berubah menjadi Rp 150.000/bulan. Jika tidak setuju, isi formulir di bit.ly/tarif-bri",
	"Akun WhatsApp kamu akan dinonaktifkan permanen. Login ulang di wa-verifikasi.com untuk mempertahankan akun.",
	"Kartu kredit Anda telah diblokir sementara. Untuk membuka blokir, klik s.id/buka-blokir",
	"Paket Anda tertahan di bea cukai. Segera transfer bea masuk Rp 1.750.000 agar paket bisa dikirim.",
	"Akun DANA Anda dibatasi karena data belum lengkap. Lengkapi data di dana-premium.site sebelum jam 12 malam.",
	"Yth. Pelanggan, nomor Anda akan diputus karena belum registrasi ulang. Registrasi di regis-kartu.xyz",
	"Hai kak, kamu dapat hadiah 5 juta dari undian Lazada! Hubungi admin di 0857-1111-2222 buat pencairan.",
	"Bapak terpilih program pemutihan pajak kendaraan. Cukup bayar 30% dari tunggakan via transfer ke rekening petugas hari ini juga.",
	"Selamat siang, saya dari Bea Cukai Soekarno-Hatta. Paket kiriman luar negeri atas nama Anda tertahan, wajib bayar pajak Rp 3.200.000 hari ini.",
	"Akun Instagram kamu dilaporkan melanggar hak cipta dan akan dihapus dalam 48 jam. Ajukan banding di ig-copyright-help.com",
	"Info dari BPJS: kartu Anda akan dinonaktifkan. Aktifkan kembali di bpjs-aktivasi.online sebelum besok.",
	"Ini dengan petugas PLN. Meteran Anda akan diputus karena pelanggaran, hubungi 0813-9999-0000 untuk penyelesaian.",
	"Mulai besok biaya admin rekening naik jadi Rp 145.000 per bulan. Jika tidak bersedia, konfirmasi lewat WA 0821-4444-5555",
}

var holdout3Legit = []string{
	"Kartu ATM Anda akan diblokir otomatis jika salah memasukkan PIN 3 kali. Info hubungi 14017.",
	"Pendaftaran lomba akan ditutup Jumat ini. Daftar di www.lombafotokaltim.com",
	"Akun Anda akan dinonaktifkan jika tidak aktif selama 90 hari. Masuk di https://ibank.klikbca.com untuk tetap aktif.",
	"Nomor Anda akan diblokir jika belum registrasi ulang. Ketik REG#NIK#NoKK# kirim ke 4444.",
	"Polisi mengimbau warga mewaspadai SMS yang menyebut rekening akan diblokir dan meminta klik tautan bri-cek.com",
	"Mulai 1 Januari, tarif transfer antarbank menjadi Rp 2.500. Jika tidak setuju, nasabah dapat menutup rekening di kantor cabang.",
	"Akun email kantor kamu akan dihapus setelah resign. Backup file penting ke drive pribadi ya.",
	"Grup WA alumni akan dihapus akhir bulan, gabung grup baru di sini: https://chat.whatsapp.com/AbCdEf",
	"Paket kiriman luar negeri dikenai bea masuk sesuai PMK 4/2025. Cek rincian di www.beacukai.go.id",
	"Saldo e-money kamu tinggal Rp 5.000, isi ulang di minimarket terdekat ya.",
	"Kartu kredit lama Anda akan dinonaktifkan setelah kartu pengganti aktif. Aktivasi di aplikasi myBCA.",
	"Langganan internet IndiHome akan diputus sementara jika tagihan belum dibayar. Bayar di indihome.co.id",
	"Mas, akun Netflix kita mau diblokir karena sharing, aku ganti password ya nanti.",
	"Rekening dormant akan ditutup otomatis setelah 12 bulan tanpa transaksi, sesuai ketentuan bank.",
	"Untuk mengaktifkan kembali kartu, silakan datang ke kantor cabang terdekat dengan membawa KTP.",
	"Biaya pendaftaran seminar Rp 50.000, transfer ke rekening panitia. Daftar di www.eventku.com",
	"Mohon segera bayar biaya pendaftaran ulang paling lambat 30 Juni ya Bapak/Ibu.",
	"Ongkir ke Samarinda Rp 25.000 ya kak, transfer ke rekening toko lalu kirim bukti ke 0812-3456-7890.",
	"Segera bayar biaya administrasi kartu pelajar di koperasi sekolah sebelum hari Jumat.",
	"Tagihan air bulan ini Rp 120.000. Bayar sebelum tanggal 20 agar tidak kena denda, bisa via aplikasi atau loket PDAM.",
	"Pak, biaya servis motor totalnya Rp 350.000, bisa transfer ke rekening bengkel ya.",
	"Kak, untuk pesanan custom ada biaya pengiriman Rp 30.000, chat admin di 0857-1234-5678 ya.",
	"Mohon segera transfer iuran kas kelas, kalau belum bayar nanti tidak bisa ikut study tour.",
	"Pinjaman tanpa jaminan, langsung cair. Biaya admin 3% dipotong saat pencairan, ajukan lewat aplikasi.",
	"Layanan internet diputus sementara karena pemeliharaan jaringan, info hubungi teknisi kami 0812-7777-8888.",
}

var holdout4Scams = []string{
	"Selamat pagi, kami dari tim keamanan BCA. Ada login mencurigakan di akun m-BCA Anda. Mohon sebutkan kode m-BCA yang tampil di layar untuk pengamanan.",
	"Hai kak, mau dapat penghasilan tambahan 300-500rb sehari? Cukup screenshot dan follow akun toko kami. Daftar via WA 0896-1234-0000",
	"PEMBERITAHUAN: Paket Anda tidak dapat dikirim karena alamat salah. Perbarui alamat dan bayar Rp 9.500 di posindo-id.com/track",
	"Ibu ini anaknya kecelakaan motor, sekarang di IGD. Segera kirim uang 7 juta untuk biaya rumah sakit ke rekening dokter jaga.",
	"Dengan hormat, nomor rekening Anda menang program Mandiri Untung Beruntung Rp 100 jt. Konfirmasi ke petugas: wa.me/6281299998888",
	"Akun TikTok Shop kamu ditangguhkan karena pelanggaran. Ajukan pemulihan sebelum 24 jam di tiktokshop-seller-center.online",
	"Mas maaf saya tadi kirim pulsa 100rb ke nomor mas, salah ketik. Mohon dikembalikan ke nomor saya ya mas.",
	"Promo investasi properti syariah, return 15% per bulan, BEP 6 bulan. Slot terbatas!",
	"Kepada Yth Wajib Pajak, NPWP Anda akan dinonaktifkan karena belum pemadanan NIK. Lakukan pemadanan di djp-pemadanan.web.id",
	"Lihat undangan acara 17an RT kita ya Pak, filenya: Undangan_17an.apk",
	"Bu, saya dari sekolah, anak ibu pingsan di kelas dan dibawa ke klinik. Tolong transfer 2,5 juta ke rek wali kelas sekarang ya.",
	"Pinjaman online cepat cair 1 jam, plafon hingga 20 juta tanpa survei. Transfer biaya verifikasi Rp 200.000 dulu ke admin.",
	"Rekening BRI Anda telah terblokir karena 3 kali salah PIN. Silakan hubungi Bpk. Andi CS BRI di 0852-9876-5432.",
	"Kamu terpilih jadi member VIP Shopee, berhak dapat voucher 2 juta. Klaim di shopee-vip-member.xyz sebelum hangus.",
	"Halo, saya Rina dari HRD PT Maju Jaya, Anda lolos seleksi. Untuk medical check up harap transfer biaya medical Rp 750.000 ke rekening kami.",
	"Ini Ayah pakai nomor teman, hp Ayah jatuh. Tolong belikan pulsa 200rb ke nomor ini dulu, nanti Ayah ganti.",
	"Dear Customer, kartu kredit BNI Anda akan dibekukan hari ini. Klik bni-cardcenter.top untuk pembaruan data.",
	"Arisan online modal 50rb dapat 750rb, 100% amanah, tanpa risiko. Join grup WA sekarang.",
	"Kak saya kurir JNE, mau konfirmasi paket. Tolong sebutkan kode yang masuk ke SMS kakak ya.",
	"Telkomsel: Selamat! Nomor Anda mendapatkan 1 unit motor dari Poin Telkomsel. Info pengambilan hadiah: 0813-5555-1212",
}

var holdout4Legit = []string{
	"Kode OTP BCA Anda 554210, berlaku 3 menit. Jangan berikan kepada siapa pun termasuk petugas BCA.",
	"Pengumuman: pendaftaran beasiswa dibuka sampai 15 Desember. Info lengkap di www.beasiswa.kemdikbud.go.id",
	"Mas, nanti tolong transfer uang arisan 200rb ke rekening aku ya, nomornya 1234567890 BCA.",
	"Kak, ongkir ke Bandung 18rb ya, total jadi 168rb. Transfer ke BCA 0987654321 a.n. Toko Rina.",
	"Selamat atas kelahiran putranya! Semoga sehat selalu. Kado sudah kami kirim lewat Gojek.",
	"Mohon maaf, akun Anda terkunci sementara karena salah kata sandi berulang. Coba lagi dalam 30 menit.",
	"Bapak/Ibu, biaya study tour kelas 6 Rp 450.000, transfer ke rekening bendahara komite paling lambat Jumat.",
	"Info: kantor pajak membuka layanan pemadanan NIK-NPWP, bisa dilakukan sendiri di pajak.go.id.",
	"Hi kak, pesanan kamu sudah aku kirim ya, resinya JP1234567890. Cek di aplikasi J&T.",
	"Mama, nanti pulang kerja aku mampir beli obat ya, Mama mau titip apa?",
	"Diskon 20% semua menu khusus hari ini! Pesan lewat GoFood atau datang langsung ke outlet.",
	"Rapat wali murid diadakan Sabtu, 9 Desember, jam 09.00 di aula sekolah. Mohon hadir tepat waktu.",
	"Investasi reksa dana pasar uang, imbal hasil sekitar 4-5% per tahun. Bukan jaminan, baca prospektus.",
	"Kartu e-toll kamu saldonya tinggal Rp 10.000, isi ulang di minimarket atau m-banking ya.",
	"Bro, sorry aku salah kirim chat tadi, abaikan aja ya.",
	"Lowongan kasir minimarket, lulusan SMA/SMK, kirim lamaran ke alamat toko. Tidak dipungut biaya.",
	"Pak RT, sampah di depan rumah belum diangkut 3 hari, mohon dibantu hubungi petugas kebersihan.",
	"Pelanggan Yth, tagihan Indihome Anda bulan ini Rp 385.000 jatuh tempo 20 Desember. Bayar di aplikasi myIndiHome.",
	"Nanti malam ada pengajian di masjid, setelah isya. Jangan lupa ajak bapak-bapak lain.",
	"WhatsApp: akun Anda dikeluarkan dari perangkat lain. Jika ini bukan Anda, buka Pengaturan > Perangkat tertaut.",
}

var holdout5Scams = []string{
	"BPJS KESEHATAN: Kartu peserta anda NONAKTIF. Aktifkan kembali dengan bayar iuran tunggakan ke VA 8888 0812 3456 789 sebelum 17.00 WIB.",
	"Kak, ada pesanan masuk atas nama kakak tapi gagal bayar. Cek detailnya di aplikasi ini ya: Detail_Pesanan.apk",
	"Saya Kompol Hendra dari Polda Metro. Nama Anda tercatat dalam kasus pencucian uang. Untuk pembersihan nama, transfer dana verifikasi ke rekening negara yang kami berikan.",
	"Pemberitahuan dari Shopee: pesanan Anda dibatalkan, dana Rp 1.250.000 akan dikembalikan setelah Anda konfirmasi lewat link shopee-refund-center.com",
	"Dapatkan penghasilan 2 juta per minggu hanya dengan nonton video YouTube! Gabung sekarang, gratis tanpa modal.",
	"Hai, aku Jessica dari Singapura, mau kirim paket hadiah untukmu tapi tertahan di bandara. Tolong bayar biaya pengiriman ulang Rp 4.500.000 ke petugas.",
	"Yth Pelanggan PLN, Anda berhak atas diskon token listrik 50%. Klaim di pln-diskon-token.com sebelum kuota habis.",
	"Nak, ini Bapak. HP Bapak rusak jadi pakai nomor ini. Tolong kirim uang 1 juta ke rekening 1234567890 sekarang, buat bayar rumah sakit.",
	"Investasi bodong? Bukan! Platform kami resmi, profit harian 3%, sudah 10 ribu member. Deposit minimal 250rb.",
	"WA Anda telah login di perangkat lain. Jika bukan Anda, kirim kode 6 digit yang Anda terima ke nomor ini untuk mengamankan akun.",
	"Halo kak, aku dari tim promo Tokopedia, kamu menang lucky draw Rp 15.000.000! Hubungi admin Tokopedia di WA 0858-2222-3333 untuk klaim.",
	"Gaji 8-15 juta/bulan, kerja di Jepang, tanpa pengalaman. Biaya proses visa dan pelatihan 25 juta, bisa dicicil.",
	"Akun Gojek Anda terdeteksi digunakan di perangkat lain dan akan diblokir. Verifikasi identitas: gojek-verify-id.com",
	"Cuma butuh KTP! Pinjol tanpa BI checking, cair 30 menit. Transfer biaya admin 150rb ke rekening CS kami terlebih dahulu.",
	"Selamat, nomor HP Anda memenangkan cek tunai Rp 175 juta dari undian Bank Rakyat Indonesia. Kode PIN pemenang: X7Y9. Hubungi 0812-0000-1111.",
	"Pak, ini dari kantor kelurahan. Bantuan sosial Bapak sudah cair 2,4 juta, tapi harus bayar biaya pencairan 300rb dulu ke petugas.",
	"Maaf mengganggu, saya petugas Bank BRI. Nomor kartu ATM Anda akan diganti ke chip baru. Mohon sebutkan 16 digit nomor kartu dan PIN Anda.",
	"URGENT!!! Akun Instagram anda akan dihapus permanen dalam 24 jam karena laporan pelanggaran. Banding di: instagram-help-center.net",
	"Ini Tante Rini, nomor baru ya. Tante lagi di rumah sakit, uang tante kurang, bisa pinjem 3 juta dulu? Nanti tante transfer balik minggu depan.",
	"Bisnis jual beli kripto dengan robot AI, untung 1-2% setiap hari otomatis. Daftar sekarang, kuota terbatas!",
}

var holdout5Legit = []string{
	"Pelanggan Yth, saldo pulsa Anda Rp 2.300. Isi ulang pulsa di aplikasi MyTelkomsel atau gerai terdekat.",
	"Selamat datang di grup alumni SMA 1! Silakan perkenalkan diri dan jangan lupa baca aturan grup.",
	"Ayah, aku mau bayar uang kuliah semester ini 4,5 juta, tolong transfer ke rekening aku sebelum tanggal 15 ya.",
	"Dari panitia: kupon doorprize jalan sehat bisa diambil di sekretariat RW mulai besok pagi.",
	"Info dari bank: layanan mobile banking sedang pemeliharaan pukul 00.00-03.00 WIB. Mohon maaf atas ketidaknyamanannya.",
	"Akun Instagram kami sempat diretas, sekarang sudah pulih. Abaikan DM aneh dari akun kami kemarin ya.",
	"Bunda, nanti tolong transfer uang les bulan ini ke rekening Bu Guru ya, Rp 300.000.",
	"Kak, paket COD kamu sudah sampai, totalnya Rp 189.000, siapkan uang pas ya.",
	"Nasabah Yth, kartu debit Anda akan kedaluwarsa bulan depan. Silakan ganti di kantor cabang terdekat tanpa biaya.",
	"Lomba 17-an: pemenang lomba balap karung mendapat hadiah dari Pak RT, diumumkan nanti malam.",
	"Mohon doanya, Bapak masuk rumah sakit tadi pagi karena darah tinggi. Sekarang sudah ditangani dokter.",
	"Gaji bulan ini sudah masuk ya, tolong dicek. Kalau ada selisih, lapor ke bagian keuangan.",
	"Sist, mau order hijab 3 pcs, total 225rb ya. Nanti aku transfer, kirim nomor rekening kakak.",
	"Workshop fotografi gratis Sabtu ini di perpustakaan kota, daftar di www.perpustakaan.samarindakota.go.id",
	"Pengingat: cicilan motor jatuh tempo tanggal 25. Bayar via aplikasi Adira atau Indomaret.",
	"Tolong kirim foto struk belanja kemarin ya, mau aku masukin laporan pengeluaran.",
	"Ma, aku lolos seleksi kerja di Balikpapan! Mulai kerja tanggal 1 bulan depan.",
	"Info grup: arisan bulan ini di rumah Bu Tini, iuran 100rb transfer ke rekening Bu Tini atau bawa tunai.",
	"PT KAI: Tiket Anda KA Argo Bromo Anggrek, kode booking ABC123, berangkat 12 Des 08.00 WIB.",
	"Pak, saya dari bengkel, motornya sudah selesai diservis. Total 275 ribu, bisa diambil sore ini.",
}

var holdout6Scams = []string{
	"PROMO SPESIAL ULANG TAHUN BRI!!! Dapatkan saldo 2.500.000 gratis. Isi formulir di bri-ultah.online paling lambat malam ini.",
	"Ass. Bu, sy dr Puskesmas. Anak ibu tadi jatuh dr motor & kepalanya berdarah. Mohon kirim dana 3jt utk tindakan, rek dokter kami 0123456789 BRI.",
	"Kak, kurir J&T nih, paketnya gak bisa dikirim krn alamat kurang jelas. Download aplikasi cek alamat ya: JNT_Cek_Alamat.apk",
	"Selamat! Anda terpilih sebagai penerima Bantuan UMKM 2026 senilai Rp 3.500.000. Daftar segera di bantuan-umkm-kemenkop.site",
	"Akun e-wallet OVO kamu akan dinonaktifkan dlm 1x24 jam krn belum upgrade ke premium. Upgrade di ovo-premium-upgrade.com",
	"Mau kerja santai dari rumah? Tugas cuma kasih bintang 5 ke hotel di Google Maps, bayaran 50rb per tugas. Langsung chat 0877-1234-5678",
	"Halo, kami dari Bank Mandiri. Kartu kredit Anda kena biaya tahunan Rp 1.200.000. Jika ingin dibatalkan, sebutkan kode OTP yang kami kirim.",
	"WOW! Investasi emas digital return 20% per bulan, dijamin halal dan aman. Minimal setor 1 juta, WA admin 0896-5555-4444.",
	"Kak, ini aku Dinda, ganti nomor. Bisa minta tolong beliin token listrik 200rb ke nomor meter ini dulu? Besok aku transfer.",
	"Info penting: Saldo BPJS Ketenagakerjaan Anda Rp 12.450.000 bisa dicairkan. Klik bpjs-jht-cair.site untuk proses pencairan, biaya admin Rp 150.000.",
	"Dek, ini Om Budi, lagi di bandara, dompet hilang. Tolong transfer 1,5 juta ke rekening teman Om ini 9876543210, nanti Om ganti.",
	"Selamat pagi Bapak/Ibu, kami dari Kantor Pajak Pratama. Ada SPT yang belum dilaporkan, dikenakan denda Rp 5.000.000. Bayar via link pajak-online-id.com hari ini.",
	"Hai kak! Kamu pemenang giveaway kami! Hadiah iPhone 16 Pro siap dikirim, tinggal bayar ongkir + asuransi Rp 350rb ya.",
	"Ini dari kantor polisi. Anak bapak kami amankan karena tawuran. Kalau mau diselesaikan secara kekeluargaan, siapkan 15 juta.",
	"Dear nasabah BCA, ada transaksi Rp 4.750.000 dari rekening Anda. Jika bukan Anda, segera hubungi Halo BCA di 0821-1500-888.",
	"Cuan dari rumah! Join tim marketing kami, cukup share link produk, komisi 100rb/hari. Daftar gratis via t.me/cuanrumahan",
	"Pemberitahuan Tilang Elektronik: kendaraan Anda melanggar lalu lintas. Lihat bukti foto di aplikasi E-Tilang.apk",
	"Nomor Anda menang kuis Ramadhan Indosat Rp 50 juta. Untuk pencairan hubungi CS di 0857-2020-3030 dan siapkan biaya pajak 2,5%.",
	"Halo Pak, kami dari tim IT kantor. Akun email kantor Bapak akan dinonaktifkan. Mohon balas dengan password email Bapak untuk verifikasi.",
	"Ma, aku di kantor polisi, nabrak orang. Polisinya minta uang damai 8 juta biar gak diperpanjang, tolong transfer sekarang ya, jangan bilang Papa.",
	"Pinjaman dana KUR tanpa agunan sampai 100 juta, proses 1 hari. Syarat: transfer biaya survei Rp 500.000 ke rekening petugas.",
	"Akun Facebook kamu dilaporkan karena pelanggaran hak cipta dan akan diblokir. Ajukan keberatan di fb-appeal-center.com dalam 24 jam.",
}

var holdout6Legit = []string{
	"Pak, rapat evaluasi diundur ke Kamis jam 2 siang ya, di ruang meeting lantai 2.",
	"Selamat ya, kamu juara 1 lomba cerdas cermat! Hadiahnya diserahkan Senin saat upacara.",
	"Promo akhir tahun: diskon 30% semua produk elektronik, kunjungi toko kami di Jl. Ahmad Yani No. 12.",
	"Bu, uang SPP bulan ini Rp 350.000 sudah saya transfer ke rekening sekolah ya, bukti terlampir.",
	"Pengingat: perpanjang SIM kamu sebelum 10 Januari. Bisa lewat aplikasi SINAR atau Satpas terdekat.",
	"Dik, nanti kalau ke pasar tolong belikan cabai 1/4 kg ya, uangnya Mama ganti.",
	"Kami dari Puskesmas Sungai Kunjang menginformasikan imunisasi balita gratis hari Rabu di posyandu.",
	"Kartu kredit Anda telah berhasil diaktifkan. Nikmati cicilan 0% di merchant pilihan.",
	"Halo, ini Andi teman kuliahmu dulu, ganti nomor ya. Apa kabar? Kapan-kapan reuni yuk.",
	"Info dari BMKG: potensi hujan lebat disertai petir di Samarinda sore ini. Hati-hati di jalan.",
	"Transfer ke rekening yang tertera di invoice ya Pak, setelah itu barang kami kirim besok.",
	"Mohon maaf, layanan aplikasi sedang gangguan. Tim kami sedang memperbaiki. Terima kasih atas kesabarannya.",
	"Bro, tiket konser udah aku beliin, nanti ganti 750rb aja ke rekening aku.",
	"Pemberitahuan: pemadaman listrik bergilir di wilayah Sempaja pukul 09.00-12.00 untuk pemeliharaan jaringan.",
	"Teman-teman, lomba menulis esai nasional dibuka, hadiah total 10 juta. Info di www.lombaesai.or.id",
	"Kak, aku udah follow akun IG tokomu ya, ditunggu promo berikutnya.",
	"Selamat siang, kami dari JNE ingin mengonfirmasi alamat pengiriman paket Anda ke Jl. Pramuka 5. Benar ya?",
	"Pak, saya izin pinjam uang 500rb dulu buat bayar kontrakan, gajian nanti saya kembalikan.",
	"Untuk keamanan akun, aktifkan verifikasi dua langkah di pengaturan WhatsApp.",
	"Ayo ikut arisan keluarga bulan depan, iuran 200rb, kocokan di rumah Nenek.",
	"Laporan: PIN ATM saya terblokir karena salah input 3 kali, besok saya ke bank untuk reset.",
	"Gaji karyawan honorer dibayar per minggu setiap Sabtu, mohon absensi diisi lengkap.",
	"Bapak masuk IGD tadi malam, sekarang sudah stabil. Doakan ya.",
	"Mau ikut kerja part time jaga stand pameran? Bayaran 150rb per hari, makan siang ditanggung. Daftar ke panitia.",
	"Promo menginap: diskon 40% khusus hari ini, pesan lewat Traveloka atau tiket.com.",
}

var leftToTheModel = []string{
	"Bpk/Ibu yth, KTP anda terdaftar sbg peminjam di 5 aplikasi pinjol. Utk pembatalan, hubungi petugas OJK 0813-2468-1357 hari ini.",
	"Maaf, saya dari bagian keuangan kantor bapak. Bapak diminta transfer dana talangan 10 jt ke rekening ini sekarang, nanti diganti kantor.",
	"Your Shopee account has been suspended. Verify your identity at shopee-verify-account.top to restore access.",
}

// Everyday messages that look like requests for a code but are not.
var notSecrets = []string{
	"Kirimkan pin lokasi rumahmu ya, aku jemput jam 7.",
	"Ma, kasih tau password wifi rumah dong, aku lupa.",
	"Nomor kartu, masa berlaku, dan CVV jangan pernah dibagikan ke siapa pun.",
	"Nomor kartu dan PIN Anda bersifat rahasia.",
	"Share pin lokasi kantor dong, aku mau antar dokumen.",
	"Kirim password zoom-nya ya, aku telat masuk.",
	"Tolong kirimkan kode booking hotelnya ke email aku.",
	"Tolong sebutkan kode pos rumah kamu buat alamat pengiriman.",
}

func TestLaterSets(t *testing.T) {
	sets := []struct {
		name  string
		scams []string
		legit []string
	}{
		{"set 3", holdout3Scams, holdout3Legit},
		{"set 4", holdout4Scams, holdout4Legit},
		{"set 5", holdout5Scams, holdout5Legit},
		{"set 6", holdout6Scams, holdout6Legit},
		{"not secrets", nil, notSecrets},
	}
	for _, set := range sets {
		for _, text := range set.scams {
			if !Analyze(text).Strong() {
				t.Errorf("%s: missed scam %q", set.name, text)
			}
		}
		for _, text := range set.legit {
			if r := Analyze(text); r.Strong() {
				t.Errorf("%s: false alarm %s on %q (hits %+v)", set.name, r.Pattern, text, r.Hits)
			}
		}
	}
	missed := 0
	for _, text := range leftToTheModel {
		if !Analyze(text).Strong() {
			missed++
		}
	}
	t.Logf("still left to the model: %d of %d", missed, len(leftToTheModel))
}

// A bank's service line is checked against the bank's own WhatsApp number.
func TestUnofficialBankContact(t *testing.T) {
	cases := map[string]string{
		"Dear nasabah BCA, ada transaksi Rp 4.750.000 dari rekening Anda. Jika bukan Anda, segera hubungi Halo BCA di 0821-1500-888.": UnofficialContact,
		"Hubungi CS BRI di 0852-9876-5432 untuk pembatalan.":                                                                          UnofficialContact,
		"Untuk bantuan, chat WhatsApp Halo BCA 0811-1500-998 atau telepon 1500888.":                                                   "",
		"Sabrina, layanan nasabah BRI di WhatsApp 0812-1214-017.":                                                                     "",
		"Call center Mandiri 14000 atau WA MITA +62 811-8414-000.":                                                                    "",
		"Saya mantri BRI unit Sempaja, untuk pengajuan KUR hubungi saya di 0812-3333-4444.":                                           "",
		"Transfer ke BCA 8730123456 a.n. Rina ya.":                                                                                    "",
		"Jangan percaya CS BCA di nomor 0813-1111-2222, itu penipuan.":                                                                "",
	}
	for text, want := range cases {
		if got := Analyze(text).Pattern; got != want {
			t.Errorf("%q: pattern %q, want %q", text, got, want)
		}
	}
}

// SMS-style messages written for these tests (placeholder numbers) in the
// wording of the public Indonesian SMS spam corpus (dataset_sms_spam_v1.csv,
// 2013-2016 messages; the corpus itself is not copied here):
// abbreviations ("M-dptkan", "Slmt", "No.anda"), winner PINs, top-ups for a
// "new number", payments redirected to someone else and dial codes.
var smsStyleScams = []struct{ text, want string }{
	{"Selamat No.SIM Card Anda M-dptkan HADIAH 1 UNIT MOBIL PIN:8XX1234 U/INFO Hub: 0812-0000-0001 klik www.info-hadiah-contoh.blogspot.com", PrizeFee},
	{"Slmt!! No anda Mendpt hadiah Rp. 75 juta dr undian poin. Info Hub : 0812 0000 0002 IR. CONTOH", PrizeFee},
	{"PESAN RESMI: PIN_PEMENANG ANDA HH25H77 Mendapatkan Hadiah 1 Unit Mobil, U/Info Klik: www.contoh-undian.webnode.com", PrizeFee},
	{"Nasabah Yth. Anda dapat cek. 27.juta dari undian bank. Kode cek anda (0259) info klik www.contoh-pusat2016.blogspot.com", PrizeFee},
	{"Plg yth no. Anda resmi meraih hadiah ke-2 pin M248K81 cek di www.contoh-hadiah29.blogspot.com", PrizeFee},
	{"Belikan dulu mama pulsa 20 ribu ini nomor barunya mama. kalau bisa kirim sekarang penting", MoneyPressure},
	{"ma, isi dl pulsa 20 rb di nmr ini 0812-0000-0003", MoneyPressure},
	{"ayah isikan pulsa 50rb di no 0812-0000-0004 sekarang penting", MoneyPressure},
	{"Yank tlong isi’in aku pulsa 10rb, di no ini, 0812-0000-0005 nanti aq yg tlpn", MoneyPressure},
	{"Mengenai uang kontrakan harap ke Bpk Contoh, Nomor Hp: 0812-0000-0006, soalnya sdh saya kuasakan penuh ke dia.", MoneyPressure},
	{"Anda mempunyai 3 pesan suara dari 0812-0000-0007; untuk mendengarkan silahkan tekan *886*10000*081200000007# tekan OK/YA.", CallForward},
	{"Paket kakak tertahan, untuk konfirmasi kurir silakan tekan **21*081200000008# lalu panggil.", CallForward},
}

// Operator promos and ordinary SMS in the same corpus style.
var smsStyleLegit = []string{
	"[PROMO] Beli paket Flash mulai 1GB di aplikasi dpt EXTRA kuota 2GB. Buruan, cek di tsel.me/mytsel S&K",
	"Plg Yth, sisa kuota Anda 478KB. Download aplikasi di http://tsel.me/tsel utk cek kuota atau hub *363#",
	"2.5 GB/30 hari hanya Rp 35 Ribu Spesial buat Anda yang terpilih. Aktifkan sekarang juga di *550*905#. S&K",
	"Isi ulang Pulsa Rp.5000 Regular pada 11/02 19:57. Ref:78330280. Info cek *888#",
	"Ma, nanti aku isi pulsa sendiri aja di konter dekat kampus.",
	"Pemenang lomba 17-an diumumkan nanti malam di balai RW, hadiahnya dibagikan Pak RT.",
	"Cek kuota tekan *363# lalu pilih menu 1.",
}

func TestSMSStyle(t *testing.T) {
	for _, c := range smsStyleScams {
		if got := Analyze(c.text).Pattern; got != c.want {
			t.Errorf("%q: pattern %q, want %q", c.text, got, c.want)
		}
	}
	for _, text := range smsStyleLegit {
		if r := Analyze(text); r.Strong() {
			t.Errorf("false alarm %s on %q (hits %+v)", r.Pattern, text, r.Hits)
		}
	}
}
