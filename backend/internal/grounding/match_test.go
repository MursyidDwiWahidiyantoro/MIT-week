package grounding

import "testing"

func TestIsFactCheck(t *testing.T) {
	for _, tc := range []struct {
		title, url string
		want       bool
	}{
		{"Video Pasukan PBB", "https://turnbackhoax.id/2026/11/01/salah-video-pasukan-pbb/", true},
		{"[HOAKS] Wapres Gibran Serahkan Surat", "https://www.kompas.com/cekfakta/read/2026/05/11/175600982/x", true},
		{"Hoaks! video Wapres Gibran mengundurkan diri", "https://www.antaranews.com/berita/5258633/hoaks-video-wapres", true},
		{"[HOAKS] Gibran Mengundurkan Diri", "https://www.komdigi.go.id/berita/berita-hoaks/detail/hoaks-gibran", true},
		// Seen as false "contradictions" on 9 Nov:
		{"Harga BBM di SPBU BP AKR Hari Ini, Naik Semua!", "https://www.liputan6.com/bisnis/read/5859493/harga-bbm", false},
		{"CekFakta", "https://cekfakta.com/institusi/19", false},
		{"spbu bp - Kumpulan Berita ANTARA News", "https://www.antaranews.com/tag/spbu-bp/8", false},
	} {
		if got := IsFactCheck(tc.title, tc.url); got != tc.want {
			t.Errorf("IsFactCheck(%q, %q) = %v, want %v", tc.title, tc.url, got, tc.want)
		}
	}
}

func TestSameTopic(t *testing.T) {
	for _, tc := range []struct {
		claim, text string
		want        bool
	}{
		{"Gibran Rakabuming mengundurkan diri dari jabatan Wakil Presiden Republik Indonesia.",
			"[HOAKS] Gibran Mengundurkan Diri sebagai Wakil Presiden Terpilih", true},
		{"Prabowo menang karena adanya Gibran", "Prabowo menang pilpres berkat Gibran, benarkah?", true},
		// Seen as false "contradictions" on 9 Nov:
		{"The Birth of a New Life - Witness the magic of creation!",
			"Cek fakta, bayi WNI baru lahir otomatis masuk peserta BPJS", false},
		{"BP Global Investments Limited melepas seluruh bisnis SPBU di Indonesia",
			"Harga BBM di SPBU BP AKR Hari Ini, Naik Semua!", false},
		{"Akun akan diblokir", "[HOAKS] Akun WhatsApp akan diblokir", true},
		{"", "apa saja", false},
	} {
		if got := SameTopic(tc.claim, tc.text); got != tc.want {
			t.Errorf("SameTopic(%q, %q) = %v, want %v", tc.claim, tc.text, got, tc.want)
		}
	}
}
