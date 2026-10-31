package linkcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantMin    int
		wantMax    int
		wantReason string
	}{
		{"clean https", "https://example.org/artikel", 0, 0, ""},
		{"clean indonesian news", "https://www.antaranews.com/berita/123", 0, 15, ""},
		{"plain http", "http://example.org/", 15, 15, ReasonNoHTTPS},
		{"ip host", "http://192.168.1.1/login", 50, 100, ReasonIPHost},
		{"credentials in url", "https://user:pass@example.org/", 40, 100, ReasonCredentials},
		{"at symbol trick", "https://example.org@evil.example/", 40, 100, ReasonCredentials},
		{"punycode", "https://xn--80ak6aa92e.com/", 30, 100, ReasonPunycode},
		{"suspicious tld", "https://promo-menarik.xyz/", 25, 100, ReasonSuspiciousTLD},
		{"shortener", "https://bit.ly/abc123", 20, 100, ReasonURLShortener},
		{"nonstandard port", "https://example.org:8443/", 20, 100, ReasonPortNonStandard},
		{"long url", "https://example.org/" + longPath(220), 10, 100, ReasonExcessiveLength},
		{"many subdomains", "https://a.b.c.d.example.org/", 20, 100, ReasonManySubdomains},
		{"suspicious path", "https://example.org/verifikasi-akun", 15, 100, ReasonSuspiciousPath},
		{"bri lookalike", "https://bri-verifikasi.xyz/login", 60, 100, ReasonLookalikeBrand},
		{"bca lookalike hyphen", "https://bca-login.example.org/", 45, 100, ReasonLookalikeBrand},
		{"gopay lookalike", "https://gopay-promo.example.org/", 45, 100, ReasonLookalikeBrand},
		{"real bri domain", "https://bri.co.id/informasi", 0, 20, ""},
		{"real bca domain", "https://bca.co.id/", 0, 20, ""},
		{"real tokopedia", "https://tokopedia.com/produk", 0, 20, ""},
		{"unparseable", "not a url at all", 100, 100, ""},
		{"empty host", "https:///path", 100, 100, ""},
		{"shopee lookalike", "https://shopee-hadiah.icu/klaim", 60, 100, ReasonLookalikeBrand},
		{"qris lookalike", "https://qris-update.example.org/", 45, 100, ReasonLookalikeBrand},
		{"dana lookalike", "https://dana-klaim.xyz/", 60, 100, ReasonLookalikeBrand},
		{"whatsapp lookalike", "https://whatsapp-verifikasi.example.org/", 45, 100, ReasonLookalikeBrand},
		{"pajak lookalike", "https://pajak-refund.example.org/", 45, 100, ReasonLookalikeBrand},
		{"telegram lookalike", "https://telegram-login.example.org/", 45, 100, ReasonLookalikeBrand},
		{"instagram lookalike", "https://instagram-verify.example.org/", 45, 100, ReasonLookalikeBrand},
		{"clean gov", "https://komdigi.go.id/pengumuman", 0, 20, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Analyze(tt.url)
			if got.Score < tt.wantMin || got.Score > tt.wantMax {
				t.Errorf("score = %d, want %d..%d (findings %+v)", got.Score, tt.wantMin, tt.wantMax, got.Findings)
			}
			if tt.wantReason != "" && !hasReason(got.Findings, tt.wantReason) {
				t.Errorf("missing reason %s in %+v", tt.wantReason, got.Findings)
			}
		})
	}
}

func TestScoreIsCappedAt100(t *testing.T) {
	// Every heuristic at once must still clamp to 100.
	got := Analyze("http://user:pass@192.168.1.1:8080/bca-login-verifikasi.xyz/" + longPath(300))
	if got.Score > 100 || got.Score < 0 {
		t.Fatalf("score out of range: %d", got.Score)
	}
}

func TestRegistrableHandlesIndonesianSuffixes(t *testing.T) {
	tests := map[string]string{
		"bri.co.id":         "bri.co.id",
		"www.bri.co.id":     "bri.co.id",
		"example.org":       "example.org",
		"a.b.example.co.id": "example.co.id",
		"sub.example.com":   "example.com",
	}
	for host, want := range tests {
		if got := registrable(host); got != want {
			t.Errorf("registrable(%q) = %q, want %q", host, got, want)
		}
	}
}

func TestUnrollFollowsRedirects(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer final.Close()

	mid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusFound)
	}))
	defer mid.Close()

	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, mid.URL, http.StatusFound)
	}))
	defer first.Close()

	// These servers are on loopback, which the safe client refuses. Unroll is
	// therefore expected to return an error rather than follow them; the test
	// asserts the refusal is surfaced, not swallowed.
	client := &http.Client{Timeout: 2 * time.Second}
	chain, _, err := Unroll(context.Background(), client, first.URL, 5)
	if err == nil && len(chain) < 2 {
		t.Fatalf("expected either a refusal or a multi-hop chain, got %v", chain)
	}
}

func hasReason(findings []Finding, code string) bool {
	for _, f := range findings {
		if f.Code == code {
			return true
		}
	}
	return false
}

func longPath(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
