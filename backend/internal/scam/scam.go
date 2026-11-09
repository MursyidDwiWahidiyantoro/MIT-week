// Package scam recognises the message patterns of the scams sent most often
// to Indonesian phones (TECH_SPEC §6.3): requests for OTP or PIN codes, APK
// files disguised as invitations or parcels, prizes that need a fee, "wrong
// transfer" refunds, paid "like and subscribe" jobs, guaranteed-profit
// investments, loans with an up-front fee, someone posing as family or a bank
// who asks for money, and notices that an account will be blocked (or a new
// fee charged) unless you follow a link or call a number that is not an
// official channel.
//
// It is deterministic and runs without the model, so a typical scam message
// still gets a verdict when the model is unavailable or out of quota. Only
// combinations that ordinary messages do not use make a message Strong (the
// SCAM_PATTERN floor); single persuasion cues only become tactics for the
// highlighter. Warnings about these scams ("jangan berikan kode OTP kepada
// siapa pun", "waspada file APK undangan") are recognised as warnings.
package scam

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/amba/amba/backend/internal/linkcheck"
	"github.com/amba/amba/backend/internal/models"
)

// Signals found in a message. Patterns are the Strong combinations.
const (
	CredentialRequest = "CREDENTIAL_REQUEST"
	APKLure           = "APK_LURE"
	WrongTransfer     = "WRONG_TRANSFER"
	JobScam           = "JOB_SCAM"
	InvestmentScam    = "INVESTMENT_SCAM"
	PrizeFee          = "PRIZE_FEE"
	LoanFee           = "LOAN_FEE"
	MoneyPressure     = "MONEY_PRESSURE"
	AccountThreat     = "ACCOUNT_THREAT"
	UnofficialContact = "UNOFFICIAL_CONTACT"
	CallForward       = "CALL_FORWARD"

	prize       = "prize"
	fee         = "fee"
	money       = "money"
	urgency     = "urgency"
	secrecy     = "secrecy"
	threat      = "threat"
	impersonate = "impersonation"
	callToAct   = "call_to_action"
	loan        = "loan"
	acctThreat  = "account_threat"
	emergency   = "emergency"
)

// Patterns lists every pattern in the order the rules try them. The API
// contract (scam_pattern) and the explanation copy must cover all of them.
var Patterns = []string{
	CredentialRequest, CallForward, APKLure, WrongTransfer, JobScam, InvestmentScam,
	PrizeFee, LoanFee, MoneyPressure, AccountThreat, UnofficialContact,
}

// Hit is one matched cue, with the phrase exactly as it appears in the text.
type Hit struct {
	Signal string
	Phrase string
}

// Result lists the cues found and the strongest scam pattern, if any.
type Result struct {
	Hits    []Hit
	Pattern string
}

// Strong reports whether the message carries a scam pattern on its own.
func (r Result) Strong() bool { return r.Pattern != "" }

// Has reports whether a cue was found.
func (r Result) Has(signal string) bool {
	for _, h := range r.Hits {
		if h.Signal == signal {
			return true
		}
	}
	return false
}

func re(pattern string) *regexp.Regexp { return regexp.MustCompile(`(?i)` + pattern) }

var (
	// Asking for a secret code. The verb comes first ("kirimkan kode OTP")
	// or the code does ("kode OTP yang masuk tolong disebutkan").
	secretWord  = `(?:kode\s*(?:otp|verifikasi|aktivasi|pembatalan(?:\s+transaksi)?|rahasia|keamanan|unik|token|sms|wa|whatsapp)|kode\s+m-?[a-z]+|kode\s+\d\s*digit|\d\s*digit\s+kode|kode\s+(?:\S+\s+){0,2}?(?:yang|yg)\s+(?:\S+\s+){0,2}?(?:masuk|dikirim|kirim|terkirim|diterima|terima|tampil|muncul)|otp|pin(?:\s*atm)?|m-?pin|password|pasword|kata\s*sandi|cvv|cvc)`
	credVerbObj = re(`\b(?:kirim(?:kan|in)?|berikan|beri|kasih(?:kan)?|kasi|beritahu(?:kan)?|kasih\s*tau|kasi\s*tau|sebut(?:kan|in)?|bacakan|baca(?:in)?|infokan|info(?:in)?|share|balas|forward|teruskan|fotokan|foto(?:in)?|screenshot|tunjukkan|siapkan|minta)\s+(?:\S+\s+){0,4}?` + secretWord + `\b`)
	credObjVerb = re(`\b` + secretWord + `\b(?:[^.!?\n]|\.\w){0,80}?\b(?:(?:tolong|mohon|harap|silakan|silahkan|bisa|boleh)\s+(?:di)?(?:sebut|kirim|beri|baca|info|forward|teruskan|foto|share)(?:kan|in)?|di(?:sebutkan|bacakan|infokan|teruskan|forward(?:kan)?)|(?:kirim|kasih|kasi|beri)(?:kan|in)?\s+ke\s+(?:saya|kami|sini|nomor\s+ini|no\s+ini))\b`)
	cardEntry   = re(`\b(?:masukkan|memasukkan|input|mengisi)\b(?:[^.!?\n]|\.\w){0,40}?\b(?:\d{2}\s*digit\s+)?(?:nomor|no)\s+kartu\b(?:[^.!?\n]|\.\w){0,60}?\b(?:cvv|cvc|masa\s+berlaku|pin)\b`)
	cardData    = re(`\b(?:isi(?:kan)?|kirim(?:kan|in)?|sebut(?:kan|in)?|berikan|beri|foto(?:kan|in)?|ketik(?:kan)?|lengkapi|balas|minta)\b(?:[^.!?\n]|\.\w){0,40}?\b(?:\d{2}\s*digit\s+)?(?:nomor|no)\s+kartu\b(?:[^.!?\n]|\.\w){0,60}?\b(?:cvv|cvc|masa\s+berlaku|pin)\b`)
	// A location pin or the password of a meeting or a Wi-Fi network is not
	// a bank secret.
	notASecret     = re(`\b(?:lokasi|location|loc|maps?|wifi|wi-fi|hotspot|zoom|meet|gmeet|meeting|rapat|room|grup|group|kelas)\b`)
	notASecretNext = re(`^(?:lokasi|location|loc|maps?|wifi|wi-fi|hotspot|zoom|meet|gmeet|meeting|rapat|room|grup|group|kelas)\b`)

	// APK files sent through chat are almost always malware in Indonesia.
	apkFile = re(`[\p{L}\p{N}][\p{L}\p{N} _\-().]{0,60}\.apk\b`)
	apkUse  = re(`\b(?:instal|install|pasang|unduh|download|buka|klik|cek)\b(?:[^.!?\n]|\.\w){0,40}?\b(?:apk|aplikasi\s+(?:undangan|surat|tilang|paket|resi|tagihan|kurir))\b`)

	// "Saya salah transfer, tolong kembalikan".
	wrongLoose    = re(`\bsalah\s+(?:ketik|input|nomor|no|rek(?:ening)?)\b[\s\S]{0,140}?\b(?:kembalikan|balikin|dikembalikan|refund)\b`)
	wrongTransfer = re(`\b(?:salah|keliru|kesalahan)\s+(?:transfer|kirim|tf|trf|isi)\b[\s\S]{0,140}?\b(?:kembalikan|balikin|dikembalikan|refund|tolong\s+(?:transfer|kirim|tf|balikin|kembalikan))\b`)

	// Paid "like and subscribe" jobs and other task scams.
	jobOffer      = re(`\b(?:kerja|pekerjaan|kerjaan|job|lowongan|loker)\b[\s\S]{0,60}?\b(?:paruh\s*waktu|part\s*-?\s*time|sampingan|freelance|online|dari\s+rumah|rumahan|harian)\b`)
	remoteWork    = re(`\b(?:online|dari\s+rumah|di\s+rumah|rumahan|via\s+hp|lewat\s+hp|modal\s+hp|cuma\s+(?:pakai|pake)\s+hp)\b`)
	jobAny        = re(`\b(?:lowongan|loker|lowker|rekrutmen|pekerjaan|kerja)\b`)
	taskWord      = re(`\b(?:like|follow|subscribe|screenshot|review|rating|ulasan|tonton|nonton|menonton|klik\s+suka|share\s+link|bagikan\s+link)\b`)
	payPerWeek    = re(`\b(?:komisi|gaji|bayaran|penghasilan|upah|dibayar|cuan|honor|income)\b[\s\S]{0,40}?\b(?:per|setiap|tiap)\s+(?:minggu|pekan)\b`)
	recruit       = re(`\b(?:lolos\s+(?:seleksi|tahap|tes|interview|wawancara)|diterima\s+(?:kerja|bekerja|sebagai)|panggilan\s+(?:kerja|interview|wawancara)|calon\s+karyawan|medical\s+check\s*-?\s*up)\b`)
	taskPay       = re(`\b(?:like|follow|subscribe|review|rating|ulasan|tonton|klik\s+suka)\b[\s\S]{0,80}?\b(?:dibayar|komisi|bayaran|reward|upah)\b`)
	payPerDay     = re(`\b(?:komisi|gaji|bayaran|penghasilan|upah|dibayar|cuan|honor|income)\b[\s\S]{0,60}?(?:\b(?:per|setiap|tiap)\s+(?:\S+\s+)?(?:hari|tugas|task|jam|video|produk|like)\b|/\s*(?:hari|tugas|task|jam)\b|\bperhari\b|\bsehari\b|\bharian\b)`)
	investRate    = re(`\b(?:profit|keuntungan|untung|return|imbal\s*hasil|cuan)\b[\s\S]{0,30}?(\d+(?:[.,]\d+)?)\s*%[\s\S]{0,30}?\b(?:per|setiap|tiap|dalam\s+\d+)\s*(hari|minggu|pekan|bulan)\b`)
	investDaily   = re(`\b(?:profit|keuntungan|untung|return|imbal\s*hasil|cuan|bunga)\s+(?:harian|mingguan|per\s*hari|perhari|per\s*minggu)\s+(?:\S+\s+){0,2}?\d+(?:[.,]\d+)?\s*%`)
	investPromise = re(`\b(?:investasi|trading|robot\s+trading|arisan|deposito|saham)\b[\s\S]{0,80}?\b(?:pasti\s+untung|dijamin\s+untung|untung\s+dijamin|tanpa\s+risiko|anti\s+rugi|modal\s+kecil\s+untung\s+besar|untung\s+pasti)\b`)

	// Cues that only count in combination.
	prizeCash = re(`\b(?:(?:anda|kamu|nomor|no|akun)\b[\s\S]{0,40}?\bterpilih\b[\s\S]{0,40}?\b(?:mendapat(?:kan)?|dapat|memenangkan|menerima)|saldo\s+(?:\S+\s+)?gratis|selamat\b[\s\S]{0,80}?\b(?:anda|kamu|nomor|no|akun|rekening|pelanggan|nasabah)\b[\s\S]{0,40}?\b(?:terpilih|memenangkan|menang|berhak\s+(?:mendapat|menerima)|resmi\s+menerima)|(?:anda|kamu|nomor|no|akun)\s+(?:\S+\s+){0,2}?(?:dapat|mendapat(?:kan)?|memenangkan|menang|berhak\s+(?:atas|mendapat(?:kan)?|menerima))\s+(?:\d[\d.,]*\s*(?:unit|buah|jt|juta|rb|ribu)?\s+|satu\s+unit\s+|sebuah\s+)?(?:hadiah|uang\s+tunai|saldo|bonus|mobil|motor)\b|(?:anda|kamu|nomor|no|akun|rekening)\s+(?:\S+\s+){0,2}?(?:menang|memenangkan)\s+(?:\S+\s+){0,4}?(?:hadiah|undian|rp\s*\d)|undian\s+berhadiah|pemenang\s+(?:undian|giveaway|kuis|lucky\s+draw)|(?:kamu|anda)\s+(?:adalah\s+)?pemenang\b|(?:saldo|dana|jht|uang)\s+(?:\S+\s+){0,6}?(?:bisa|dapat|siap|sudah|telah)\s+dicairkan|hadiah\s+(?:\S+\s+){0,3}?siap\s+dikirim|memenangkan\s+(?:undian|hadiah|uang|mobil|motor|cek)|hadiah\s+undian|uang\s+tunai\s+(?:sebesar\s+)?rp|selamat\b[\s\S]{0,60}?\b(?:mendapat(?:kan)?|dapat)\s+(?:hadiah|uang\s+tunai|mobil|motor|cek\s+tunai)|cek\s+tunai|uang\s+tunai\s+sebesar|hadiah\s+(?:uang|mobil|motor|cek|tunai|sebesar)|dana\s+kaget|bantuan\s+(?:\S+\s+){0,3}?gratis|klaim\s+(?:hadiah|saldo|bansos|bantuan|dana)|bansos\s+(?:cair|rp)|blt\s+(?:cair|rp)|(?:bansos|blt|bantuan(?:\s+sosial)?)\s+(?:\S+\s+){0,3}?(?:sudah|telah)\s+cair|(?:anda|kamu|nomor|no|akun)\s+(?:\S+\s+){0,2}?berhak\s+(?:atas|mendapat(?:kan)?|menerima)\s+(?:diskon\s+token|token\s+(?:listrik\s+)?gratis|subsidi|kuota\s+(?:internet\s+)?gratis)|(?:dana|uang|saldo|refund|pengembalian\s+dana)\b(?:[^.!?\n]|\.\w){0,60}?\b(?:dikembalikan|direfund|dicairkan)\b(?:[^.!?\n]|\.\w){0,30}?\b(?:setelah|jika|bila|dengan)\s+(?:anda\s+|kamu\s+)?(?:konfirmasi|verifikasi|mengisi|isi|klik|login|melengkapi|lengkapi|mengklik))`)
	// Fees asked for to release a prize, a loan, a parcel or an account, or
	// paid into an officer's own account. Ordinary services rarely ask for
	// these by chat.
	scamFee = re(`\b(?:biaya\s+(?:pencairan|pajak|materai|meterai|asuransi|aktivasi|verifikasi|pengurusan|balik\s+nama|deposit|jaminan|penarikan|pelunasan|pembukaan(?:\s+blokir)?|proses|pengiriman\s+ulang|pelatihan|training|seragam|medical)|pajak\s+hadiah|tebus(?:an)?|(?:uang|dana)\s+(?:damai|jaminan|tebusan|pelicin|verifikasi|pembersihan|pengamanan|titipan)|bea\s+(?:masuk|cukai|impor)|rekening\s+(?:petugas|pribadi|oknum|negara|penampungan|(?:yang|yg)\s+(?:kami|saya)\s+(?:berikan|kirim(?:kan)?)))\b`)
	// Fees that ordinary messages mention all the time (registration,
	// shipping, admin). They count with a job offer, a prize or a loan only.
	plainFee = re(`\bbiaya\s+(?:admin(?:istrasi)?|pengiriman|ongkir|pendaftaran|registrasi)\b`)
	// A loan fee paid before the money arrives, not deducted from it.
	upfront    = re(`\b(?:di\s*(?:awal|muka|depan)|terlebih\s+dahulu|duluan|sebelum\s+(?:pencairan|dana\s+cair|cair))\b`)
	moneyAsk   = re(`\b(?:transfer(?:kan|in)?|trf|tf|kirim(?:kan|in)?\s+(?:\S+\s+){0,2}?(?:uang|dana|duit|saldo|pulsa)|isi(?:kan|in)?\s+(?:\S+\s+){0,2}?(?:pulsa|saldo)|top\s*up|beli(?:kan|in)?\s+(?:\S+\s+){0,3}?pulsa|minta\s+(?:pulsa|uang|duit)|(?:pinjam|pinjem)(?:in|kan)?\s+(?:\S+\s+){0,2}?(?:uang|duit|dulu|dana|\d+\s*(?:rb|ribu|jt|juta))|bayar(?:kan|in)?|setor(?:kan)?|siapkan\s+(?:uang\s+|dana\s+)?(?:rp\.?\s*)?\d+(?:[.,]\d+)?\s*(?:rb|ribu|jt|juta))(?:\b|\d)`)
	urgent     = re(`\b(?:segera|secepatnya|hari\s+ini\s+(?:juga|saja)|khusus\s+hari\s+ini|sekarang\s+juga|batas\s+waktu|1\s*x\s*24\s*jam|dalam\s+\d+\s*jam|sebelum\s+(?:jam|pukul|dihapus|hangus|kadaluarsa|kedaluwarsa)|kesempatan\s+terakhir|jangan\s+sampai\s+ketinggalan|(?:verifikasi|klaim|konfirmasi|login|perbarui|aktifkan|lengkapi|transfer|tf|kirim|bayar)\s+sekarang|buruan|darurat|mendesak)\b`)
	secret     = re(`\b(?:jangan\s+(?:bilang|kasih\s*(?:tau|tahu)|beritahu|ceritakan|kasi\s*tau|kabari)\s+(?:\S+\s+){0,3}?(?:siapa\s*(?:pun)?|siapa-siapa|orang|mama|papa|ayah|ibu|bapak|keluarga|polisi)|rahasiakan|cukup\s+(?:kita|kamu|anda)\s+(?:berdua|yang\s+tahu))\b`)
	threatCue  = re(`\b(?:akan\s+(?:di)?(?:blokir|nonaktif(?:kan)?|tutup|hapus|tangguhkan|bekukan|suspend|putus|cabut|matikan)|(?:di)?(?:blokir|nonaktifkan|bekukan|tangguhkan|suspend)\s+(?:dalam|jika|bila|apabila|permanen|otomatis)|denda\s+(?:sebesar\s+)?rp|dipolisikan|diproses\s+(?:secara\s+)?hukum|surat\s+tilang)\b`)
	imperson   = re(`\b(?:ini\s+(?:nomor|no|nmr)\s+(?:baru|mama|papa|ibu|bapak|ayah|kakak|adik|anak|saya)|ini\s+(?:\S+\s+){1,2}?(?:nomor|no|nmr)\s+baru|(?:nomor|no|nmr)\s+baru\s+(?:aku|saya|ku|mama|papa|ibu|bapak|ayah|kakak|adik|om|tante|nenek|kakek)|(?:pakai|pake|pinjam|minjem)\s+(?:nomor|no|nmr|hp)\s+(?:ini|teman|temen|kawan|orang)|ini\s+(?:om|tante|pakde|bude|paman|bibi|kakak|kak|adik|dek|mama|papa|ibu|bapak|ayah|nenek|kakek)\s+\S+(?:[^.!?\n]|\.\w){0,60}?\b(?:dompet|hp|ponsel|tas)\s+(?:\S+\s+)?(?:hilang|dicuri|kecopetan|jatuh|rusak|ketinggalan)|(?:saya|kami|ini)\s+(?:adalah\s+)?(?:kompol|akp|ipda|iptu|aiptu|aipda|bripka|brigadir|briptu|bripda|kombes|akbp|jaksa|hakim|penyidik)\b|ini\s+(?:mama|papa|ibu|bapak|ayah|kakak|adik|om|tante|aku|saya)\s+(?:pakai|pake|pinjam|minjem)\s+(?:nomor|no|nmr|hp)|hp\s+(?:mama|papa|ibu|ayah|bapak|kakak|adik|om|tante|nenek|kakek|aku|saya|ku)\s+(?:rusak|hilang|mati|dicuri|kecebur|jatuh)|ganti\s+(?:nomor|no|nmr)|(?:nomor|hp)\s+(?:lama|yang\s+lama)\s+(?:hilang|rusak|mati|dicuri)|(?:saya|kami|ini)\s+(?:adalah\s+)?(?:dokter|perawat|bidan|polisi)\b|(?:saya|sy|kami|ini)\s+(?:adalah\s+|dengan\s+)?(?:dari|dr|petugas|pihak|cs|customer\s+service|kurir|staf|staff)\s+(?:\S+\s+){0,2}?(?:puskesmas|rumah\s+sakit|rs|klinik|bank|bri|bca|bni|btn|mandiri|bsi|j&t|jnt|jne|pos\s+indonesia|kantor\s+pos|bea\s+cukai|imigrasi|telkomsel|pln|bpjs|pajak|polisi|polda|polres|kejaksaan|pengadilan|ojk|kemensos|shopee|tokopedia|dana|ovo|gopay))\b`)
	callAction = re(`\b(?:klik|kunjungi|buka|hubungi|hub|telepon|telp|call|chat|wa|whatsapp|sms|klaim|daftar|registrasi|regis|ambil|akses|isi|mengisi|konfirmasi|verifikasi|validasi|login|log\s+in|masuk|update|perbarui|aktifkan|aktivasi|lengkapi|ajukan|bayar|transfer|tf|kirim|cek)\b(?:[^.!?\n]|\.\w){0,40}?(?:` + destination + `)`)
	// SMS-era prize wording: "M-dptkan hadiah", "PIN pemenang", "hadiah
	// ke-2", "cek Rp.27jt", "Slmt ... hadiah".
	prizeSlang = re(`\b(?:(?:pin|kode)[\s_:.-]*(?:pemenang|undian|hadiah|cek)|pemenang\s+(?:hadiah|ke|promo|poin|program)|hadiah\s*(?:ke|ke-)\s*-?\s*[({]?\d|(?:anda|no\.?|nomor|resmi)\s+(?:\S+\s+){0,2}?(?:meraih|m-?e?n?['’-]?dpt(?:kan|kn)?|mendpt|mndpt|mndptkn|men-?dapat(?:kan)?|mendpatkn|mdpt|mendapat(?:kan)?)\s+(?:\S+\s+){0,3}?(?:hadiah|cek|uang\s+tunai|mobil|motor|bonus\s+poin|honda|toyota|nissan|suzuki|yamaha|daihatsu|mitsubishi|iphone|samsung)|cek\s+(?:tu?anai\s*)?(?:rp|senilai)|cek\s*\.?\s*(?:rp\.?\s*)?\d+(?:[.,]\d+)?\s*\.?\s*(?:jt|juta)\b|(?:trpilih|terpilh)\b|(?:selamat|slmt|selamt|slamat)\b(?:[^.!?\n]|\.\w){0,80}?\b(?:hadiah|pemenang|cek\s+tunai|mobil|motor))`)
	// "Isi pulsa di nomor baru ini": a top-up for a "new number", or for a
	// mobile number with "penting / sekarang".
	pulsaAsk  = re(`\b(?:isi(?:kan|['’]?in)?|beli(?:kan|['’]?in)?|kirim(?:kan|['’]?in)?|tf|transfer)\s+(?:\S+\s+){0,3}?pulsa`)
	newNumber = re(`\b(?:nomor|no|nmr)\s+baru(?:nya|ku)?\b|\b(?:di|ke)\s+(?:nomor|no|nmr)\s+(?:ini|baru(?:nya|ku)?)\b`)
	pressing  = re(`\b(?:sekarang|penting|segera|cepat|secepatnya|darurat|urgent)\b`)
	// Payment redirected to someone else: "sudah saya kuasakan ke Bpk X".
	redirected = re(`\b(?:sudah|sdh|shd|telah)\s+(?:saya\s+|sy\s+|sya\s+)?(?:di)?kuasakan\b|\bada\s+perubahan\s+pembayaran\b|\bpembayaran\s+(?:harus\s+)?(?:melalui|lewat)\s+(?:ke\s+)?(?:bpk|bapak|pak|ibu|bu)\b`)
	// A dial code that carries a phone number, or a call-forwarding code
	// (**21*, *21*, **61*, **62*, **67*): it diverts calls and SMS, OTP
	// codes included.
	forwardCode = re(`(?:\*{1,2}(?:21|61|62|67)\*|\*\d{2,4}\*(?:\d+\*)*(?:\+?62|0)8\d{7,12}#?)`)
	// Someone reports that your relative had an accident or was arrested.
	relativeEmergency = re(`\b(?:anak|suami|istri|ayah|ibu|bapak|adik|kakak|cucu|keluarga)(?:nya|\s+(?:ibu|bapak|bpk|anda|kamu))\b(?:[^.!?\n]|\.\w){0,60}?\b(?:kecelakaan|tertabrak|jatuh|berdarah|terluka|luka\s+parah|masuk\s+(?:rs|rumah\s+sakit|igd|icu)|di\s+(?:igd|icu)|dirawat|(?:harus\s+(?:segera\s+)?)?dioperasi|ditahan|ditangkap|(?:kami\s+)?amankan|diamankan|pingsan|kritis|koma)\b`)
	// An account, card, number or service that "will be blocked": blocking
	// verbs in any form, or "akan" with cutting off, closing or deleting
	// ("layanan diputus sementara karena pemeliharaan" is not a threat).
	accountThreat = re(`\b(?:akun|rekening|rek|kartu|nomor|no|nmr|simcard|sim\s+card|saldo|atm|m-?banking|mobile\s+banking|internet\s+banking|brimo|livin|wondr|whatsapp|wa|facebook|fb|instagram|ig|tiktok|bpjs|npwp|nik|e-?ktp|ktp|listrik|meteran|kwh|layanan|langganan|paylater|limit)\b(?:[^.!?\n]|\.\w){0,60}?\b(?:(?:di|ter)(?:blokir|nonaktif(?:kan)?|tangguhkan|bekukan|suspend|kunci|batasi|cabut)|nonaktif|akan\s+(?:di)?(?:blokir|nonaktif(?:kan)?|tutup|hapus|putus|hentikan|tangguhkan|bekukan|cabut|matikan))\b`)
	// "Tarif transfer berubah ... jika tidak setuju, isi formulir di ...".
	tariffThreat = re(`\b(?:tarif|biaya)\s+(?:baru|transfer|transaksi|admin(?:istrasi)?|bulanan|layanan)\b[\s\S]{0,140}?\b(?:jika|bila|apabila|kalau|kalo)\s+(?:anda\s+|bapak/ibu\s+|bapak\s+|ibu\s+|nasabah\s+|kamu\s+)?(?:tidak|tdk|tak|belum)\s+(?:setuju|bersedia|ingin|mau|menyetujui)\b`)
	// A parcel "held" until a fee is paid.
	parcelHeld = re(`\bpaket\b(?:[^.!?\n]|\.\w){0,40}?\b(?:tertahan|ditahan|gagal\s+dikirim|tidak\s+dapat\s+dikirim|tertunda)\b`)
	// Money sent to "this account": an account number after "ke rekening".
	accountNumber = re(`\b(?:rek(?:ening)?|no\.?\s*rek)\.?(?:\s+[^\s\d]+){0,3}?\s*:?\s*\d{3}[\d\s-]{6,16}\d\b`)
	loanBait      = re(`\b(?:pinjaman|pinjol|dana\s+tunai|kredit|kta)\b[\s\S]{0,80}?\b(?:tanpa\s+(?:jaminan|agunan|bi\s*checking|slik|survei|survey)|langsung\s+cair|cair\s+dalam|acc\s+cepat|pasti\s+(?:cair|acc))\b`)

	// A denial or a warning just before a cue: "jangan berikan kode OTP",
	// "Telkomsel tidak pernah meminta ...", "waspada file ... .apk".
	negation = regexp.MustCompile(`(?i)\b(?:jangan|jgn|tidak\s+pernah|tdk\s+pernah|tak\s+pernah|tidak\s+akan|tidak\s+boleh|dilarang|hindari|waspada|waspadai|hati-hati|(?:kalau|kalo|jika|bila|apabila)\s+ada\s+(?:yang|orang|pihak|seseorang))\b`)

	// Words of an awareness message or news report about scams. Scam
	// messages themselves do not talk about "modus penipuan"; a warning that
	// quotes one must not be scored as the scam it describes.
	awareness = re(`\b(?:waspada|waspadai|hati-hati\s+(?:terhadap|dengan|jika|bila)|modus|penipuan|penipu|hoaks|hoax|imbau|mengimbau|himbau|menghimbau|phishing|smishing|rekayasa\s+sosial|social\s+engineering|cek\s+fakta|pelaku|korban|tersangka)\b`)
)

// Analyze finds the scam cues and patterns in a message.
func Analyze(text string) Result {
	text = strings.TrimSpace(text)
	var r Result
	if text == "" {
		return r
	}
	add := func(signal string, rx *regexp.Regexp, guarded bool) bool {
		found := false
		for _, loc := range rx.FindAllStringIndex(text, 4) {
			if guarded && negated(text, loc[0]) {
				continue
			}
			r.Hits = append(r.Hits, Hit{Signal: signal, Phrase: clip(text[loc[0]:loc[1]])})
			found = true
		}
		return found
	}

	cred := addSecret(&r, text, credVerbObj)
	cred = addSecret(&r, text, credObjVerb) || cred
	cred = add(CredentialRequest, cardData, true) || cred
	if unofficialDestination(text) {
		cred = add(CredentialRequest, cardEntry, true) || cred
	}
	apk := add(APKLure, apkFile, true)
	apk = add(APKLure, apkUse, true) || apk
	wrong := add(WrongTransfer, wrongTransfer, false)
	wrongish := wrongLoose.MatchString(text)
	job := add(JobScam, jobOffer, false)
	task := add(JobScam, taskPay, false)
	perDay := payPerDay.MatchString(text)
	jobAd := jobAny.MatchString(text) || recruit.MatchString(text)
	tasked := taskWord.MatchString(text)
	perWeek := payPerWeek.MatchString(text)
	remote := remoteWork.MatchString(text)
	invest := investment(text)
	if invest != "" {
		r.Hits = append(r.Hits, Hit{Signal: InvestmentScam, Phrase: clip(invest)})
	}
	pz := add(prize, prizeCash, true)
	pz = add(prize, prizeSlang, true) || pz
	sf := add(fee, scamFee, false)
	fe := add(fee, plainFee, false) || sf
	mo := add(money, moneyAsk, true)
	ur := add(urgency, urgent, false)
	se := add(secrecy, secret, false)
	th := add(threat, threatCue, true)
	im := add(impersonate, imperson, false)
	em := add(emergency, relativeEmergency, true)
	ca := addCallToAction(&r, text)
	ln := add(loan, loanBait, false)
	fakeLine := addUnofficialContact(&r, text)
	forward := add(CallForward, forwardCode, true)
	topUp := pulsaAsk.MatchString(text)
	toNewNumber := newNumber.MatchString(text)
	toMobile := mobile.MatchString(text)
	diverted := redirected.MatchString(text)
	acct := add(acctThreat, accountThreat, true)
	acct = add(acctThreat, tariffThreat, false) || acct
	held := parcelHeld.MatchString(text)
	// Somewhere to send money or to call that is not an official channel:
	// a call to action, any unofficial link or mobile number, or an account
	// number.
	elsewhere := ca || unofficialDestination(text) || accountNumber.MatchString(text)

	if awareness.MatchString(text) {
		// Cues stay visible as tactics; no pattern is claimed.
		return r
	}
	switch {
	case cred:
		r.Pattern = CredentialRequest
	case forward:
		r.Pattern = CallForward
	case apk:
		r.Pattern = APKLure
	case wrong || (wrongish && mo):
		r.Pattern = WrongTransfer
	case task || (tasked && (perDay || perWeek)) || ((job || remote) && perDay && (remote || elsewhere)) || (jobAd && fe):
		r.Pattern = JobScam
	case invest != "":
		r.Pattern = InvestmentScam
	case pz && (fe || mo || se || th || elsewhere):
		r.Pattern = PrizeFee
	case ln && fe && (mo || upfront.MatchString(text)):
		r.Pattern = LoanFee
	case topUp && (toNewNumber || (toMobile && pressing.MatchString(text))):
		r.Pattern = MoneyPressure
	case diverted && toMobile:
		r.Pattern = MoneyPressure
	case mo && (im || em || (th && elsewhere) || (se && ur) || (ur && sf) || (sf && ca) || (held && (ca || sf))):
		r.Pattern = MoneyPressure
	case acct && elsewhere:
		r.Pattern = AccountThreat
	case fakeLine:
		r.Pattern = UnofficialContact
	}
	return r
}

// A bank's service line ("Halo BCA", "CS BRI", "call center Mandiri") given
// as a mobile number that is not the bank's own. Only the banks whose
// WhatsApp numbers are listed in officialNumbers are checked.
var (
	serviceCue = re(`\b(?:halo|cs|customer\s+service|call\s+cent(?:er|re)|contact\s+cent(?:er|re)|layanan\s+(?:nasabah|pelanggan|pengaduan)|pengaduan)\b`)
	bankCue    = re(`\b(?:bca|bri|mandiri)\b`)
	mobile     = re(`(?:\+?\b62[\s-]?|\b0)8\d{1,2}[\s-]?\d{3,4}[\s-]?\d{3,5}\b`)
)

// addUnofficialContact records a bank service line on a number the bank does
// not use.
func addUnofficialContact(r *Result, text string) bool {
	found := false
	for _, loc := range mobile.FindAllStringIndex(text, 4) {
		window := text[max(0, loc[0]-50):loc[0]]
		if k := strings.LastIndexAny(window, "!?\n"); k >= 0 {
			window = window[k+1:]
		}
		if !serviceCue.MatchString(window) || !bankCue.MatchString(window) || officialNumbers[phoneDigits(text[loc[0]:loc[1]])] {
			continue
		}
		if negated(text, max(0, loc[0]-len(window))) {
			continue
		}
		r.Hits = append(r.Hits, Hit{Signal: UnofficialContact, Phrase: clip(strings.TrimSpace(window) + " " + text[loc[0]:loc[1]])})
		found = true
	}
	return found
}

// phoneDigits normalises a phone number to its national form (08...).
func phoneDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	if strings.HasPrefix(d, "62") {
		d = "0" + d[2:]
	}
	return d
}

// unofficialDestination reports a link or a mobile number anywhere in the
// text that is not one of the official domains.
func unofficialDestination(text string) bool {
	for _, m := range destinationRe.FindAllString(text, 8) {
		if !official(m) {
			return true
		}
	}
	return false
}

// addSecret records a request for a code, PIN or password, but not for a
// location pin or the password of a meeting or a Wi-Fi network.
func addSecret(r *Result, text string, rx *regexp.Regexp) bool {
	found := false
	for _, loc := range rx.FindAllStringIndex(text, 4) {
		if negated(text, loc[0]) {
			continue
		}
		after := strings.TrimLeft(text[loc[1]:min(len(text), loc[1]+24)], " -:")
		if notASecret.MatchString(text[loc[0]:loc[1]]) || notASecretNext.MatchString(after) {
			continue
		}
		r.Hits = append(r.Hits, Hit{Signal: CredentialRequest, Phrase: clip(text[loc[0]:loc[1]])})
		found = true
	}
	return found
}

// addCallToAction records a request to click, call or visit somewhere,
// except a link to an official domain ("cek di cekbansos.kemensos.go.id").
func addCallToAction(r *Result, text string) bool {
	found := false
	for _, loc := range callAction.FindAllStringIndex(text, 4) {
		phrase := text[loc[0]:loc[1]]
		if official(destinationRe.FindString(phrase)) {
			continue
		}
		r.Hits = append(r.Hits, Hit{Signal: callToAct, Phrase: clip(phrase)})
		found = true
	}
	return found
}

// destination is a link or an Indonesian mobile number. Banks, operators and
// agencies use short codes, 1500-numbers and their own domains, not 08xx.
const destination = `(?:https?://)?(?:wa\.me|api\.whatsapp\.com/send\?phone=)/?\+?\d{9,15}|https?://\S+|www\.\S+|\b(?:[a-z0-9-]+\.)+(?:com|id|net|org|co|io|me|ly|xyz|top|site|online|link|info|shop|store|click|icu|live|cc|app|vip|biz|website|space|fun|tk|ml|ga|cf|gq|pw|cfd|sbs|cyou|lol|bond|buzz|rest|cam|monster|ws|gg|to|club|pro|asia|win|work|life|tech)\b(?:/\S*)?|\b0?8\d{2}[\s-]?\d{3,4}[\s-]?\d{3,5}\b|\+?\b62[\s-]?8\d{1,2}[\s-]?\d{3,4}[\s-]?\d{3,5}\b`

var (
	destinationRe = re(destination)
	hostInCTA     = re(`(?:https?://)?(?:[a-z0-9-]+\.)+[a-z]{2,}`)
	waNumber      = re(`(?:wa\.me|api\.whatsapp\.com/send\?phone=)/?\+?\d{9,15}`)
)

// officialNumbers are banks' own WhatsApp numbers, which look like ordinary
// mobile numbers (checked November 2026): Halo BCA on bca.co.id, BRI's
// Sabrina as listed by Bisnis.com (14 August 2024), Mandiri's MITA on
// bankmandiri.co.id.
var officialNumbers = map[string]bool{
	"08111500998": true, // BCA
	"08121214017": true, // BRI
	"08118414000": true, // Mandiri
}

// official reports a destination on an official domain or one of the
// official numbers. A wa.me link to a phone number is that number.
func official(dest string) bool {
	host := hostInCTA.FindString(dest)
	if host == "" || waNumber.MatchString(dest) {
		return officialNumbers[phoneDigits(dest)]
	}
	host = strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(host), "https://"), "http://")
	return linkcheck.IsOfficialHost(host)
}

// investment returns the matched phrase of a guaranteed or implausible
// return: any rate per day or week, 10 % or more per month, or a promise of
// profit without risk. Bank interest per year is not matched.
func investment(text string) string {
	if m := investDaily.FindString(text); m != "" {
		return m
	}
	for _, m := range investRate.FindAllStringSubmatch(text, 4) {
		rate, err := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", "."), 64)
		if err != nil {
			continue
		}
		period := strings.ToLower(m[2])
		if period != "bulan" || rate >= 10 {
			return m[0]
		}
	}
	if m := investPromise.FindString(text); m != "" {
		return m
	}
	return ""
}

// negated reports a denial or warning in the few words before position i.
func negated(text string, i int) bool {
	start := i - 48
	if start < 0 {
		start = 0
	}
	// Do not cross a sentence boundary: "Jangan panik. Kirim kode OTP..."
	window := text[start:i]
	if k := strings.LastIndexAny(window, ".!?\n"); k >= 0 {
		window = window[k+1:]
	}
	return negation.MatchString(window)
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 80 {
		s = string(r[:80])
	}
	return s
}

// tacticFor maps cues onto the persuasion tactics the UI explains (TECH_SPEC
// §6.1). Requests for money or codes are scam patterns, not tactics.
var tacticFor = map[string]string{
	urgency:     "URGENCY",
	secrecy:     "SECRECY",
	threat:      "FEAR",
	emergency:   "FEAR",
	impersonate: "IMPERSONATION",
	prize:       "REWARD_BAIT",
}

// Tactics turns the persuasion cues into candidate tactics. The caller still
// runs them through tactics.Verify against the same text.
func (r Result) Tactics(source string) []models.Tactic {
	out := make([]models.Tactic, 0, 5)
	for _, h := range r.Hits {
		if code, ok := tacticFor[h.Signal]; ok {
			out = append(out, models.Tactic{Code: code, Phrase: h.Phrase, Source: source})
		}
	}
	return out
}
