package retrieval

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/models"
)

func TestPublisherURLUnwrapsBingRedirect(t *testing.T) {
	raw := "http://www.bing.com/news/apiclick.aspx?ref=FexRss&aid=&tid=x&url=https%3a%2f%2fwww.kompas.com%2fcekfakta%2fread%2f1&c=1"
	if got := publisherURL(raw); got != "https://www.kompas.com/cekfakta/read/1" {
		t.Fatalf("publisherURL = %q", got)
	}
	if got := publisherURL("https://www.kompas.com/x"); got != "https://www.kompas.com/x" {
		t.Fatalf("direct url = %q", got)
	}
	if got := publisherURL(""); got != "" {
		t.Fatalf("empty = %q", got)
	}
}

func TestSanitizeQueryDropsOperatorsAndURLs(t *testing.T) {
	got := SanitizeQuery(`site:turnbackhoax.id -bantuan "Rp 500.000" hari ini https://x.y/z`)
	for _, banned := range []string{"site:", "-bantuan", "https://"} {
		if strings.Contains(got, banned) {
			t.Errorf("query %q still contains %q", got, banned)
		}
	}
	if !strings.Contains(got, "hari") {
		t.Errorf("query lost ordinary words: %q", got)
	}
}

func TestHeuristicQueryRemovesLinks(t *testing.T) {
	got := HeuristicQuery("Cek tautan https://bit.ly/abc dan https://evil.example/login sekarang")
	if strings.Contains(got, "http") {
		t.Fatalf("query kept a URL: %q", got)
	}
	if !strings.Contains(got, "sekarang") {
		t.Fatalf("query = %q", got)
	}
}

func TestClassify(t *testing.T) {
	tests := map[string]models.FactStatus{
		"[HOAKS] Tautan bantuan sosial Rp 900.000":      models.FactContradicts,
		"Cek fakta: klaim bantuan Rp 500 ribu":          models.FactContradicts,
		"Pemerintah resmi mengumumkan bantuan tersebut": models.FactSupports,
		"Bantuan Rp 500.000 dibagikan hari ini":         models.FactUnverified,
	}
	for text, want := range tests {
		if got := classify(text); got != want {
			t.Errorf("classify(%q) = %s, want %s", text, got, want)
		}
	}
}

func TestAnnotateKeepsRelatedAndRanksContradictionsFirst(t *testing.T) {
	items := []Item{
		{Title: "Kumpulan hoaks Prabowo-Gibran", Snippet: "politik"},
		{Title: "[HOAKS] Tautan bantuan sosial Rp 900.000", Snippet: "cek fakta bantuan sosial"},
		{Title: "Bantuan sosial Rp 900.000 disalurkan", Snippet: "pemerintah"},
	}
	got := Annotate("Bantuan sosial Rp 900.000 dibagikan hari ini", items)
	if len(got) < 2 {
		t.Fatalf("kept %d items, want at least 2: %+v", len(got), got)
	}
	if got[0].Match != models.FactContradicts {
		t.Fatalf("first item = %s (%q), want CONTRADICTS first", got[0].Match, got[0].Title)
	}
	for _, it := range got {
		if strings.Contains(it.Title, "Prabowo") {
			t.Fatalf("unrelated political article kept: %+v", it)
		}
	}
}

func TestAggregate(t *testing.T) {
	if got := Aggregate([]Item{{Match: models.FactSupports}, {Match: models.FactContradicts}}); got != models.FactContradicts {
		t.Errorf("aggregate = %s, want CONTRADICTS", got)
	}
	if got := Aggregate([]Item{{Match: models.FactSupports}, {Match: models.FactUnverified}}); got != models.FactSupports {
		t.Errorf("aggregate = %s, want SUPPORTS", got)
	}
	if got := Aggregate(nil); got != models.FactUnverified {
		t.Errorf("aggregate nil = %s, want UNVERIFIED", got)
	}
}

// End-to-end: the feed is parsed, Bing redirects are unwrapped, off-allowlist
// publishers are dropped, and the result carries a real publisher URL.
func TestSearchFiltersByAllowlist(t *testing.T) {
	feed := `<?xml version="1.0"?><rss version="2.0"><channel>
<item><title>[HOAKS] Bantuan Rp 500 ribu</title><link>http://www.bing.com/news/apiclick.aspx?url=https%3a%2f%2fwww.kompas.com%2fcekfakta%2fread%2f1</link><description>cek fakta</description><pubDate>Thu, 22 Jan 2026 16:00:00 GMT</pubDate><Source>Kompas</Source></item>
<item><title>Bantuan Rp 500 ribu</title><link>https://random-blog.example/post</link><description>blog</description><Source>Blog</Source></item>
</channel></rss>`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("format") != "RSS" {
			t.Errorf("format = %q, want RSS", r.URL.Query().Get("format"))
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(feed))
	}))
	defer srv.Close()

	s := New(grounding.New([]string{"kompas.com"}))
	s.Endpoint = srv.URL + "/news/search"
	s.Client = srv.Client()
	s.Client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	items, err := s.Search(context.Background(), "bantuan 500 ribu")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %+v, want exactly the allowlisted one", items)
	}
	if items[0].URL != "https://www.kompas.com/cekfakta/read/1" {
		t.Fatalf("url = %q", items[0].URL)
	}
	if items[0].Publisher != "Kompas" {
		t.Fatalf("publisher = %q", items[0].Publisher)
	}
}

func TestSearchEmptyQuerySkipsNetwork(t *testing.T) {
	s := New(nil)
	s.Endpoint = "https://invalid.invalid/news"
	items, err := s.Search(context.Background(), "   ")
	if err != nil || len(items) != 0 {
		t.Fatalf("items=%v err=%v", items, err)
	}
}

// A hoax about a different amount must not be cited for this claim.
func TestAnnotateRejectsDifferentAmounts(t *testing.T) {
	items := []Item{{
		Title:   "[HOAKS] Tautan Pengecekan Penerima Bantuan Sosial Rp 900.000 dan Rp 2,5 Juta",
		Snippet: "cek fakta bantuan sosial",
	}}
	got := Annotate("Bantuan Rp 500.000 dibagikan hari ini saja", items)
	if len(got) != 0 {
		t.Fatalf("a different rupiah amount was accepted: %+v", got)
	}
}

func TestAnnotateAcceptsMatchingAmount(t *testing.T) {
	items := []Item{{
		Title:   "[HOAKS] Bantuan Rp 500.000 dari Kemensos",
		Snippet: "cek fakta bantuan sosial",
	}}
	got := Annotate("Bantuan Rp 500.000 dibagikan hari ini saja", items)
	if len(got) != 1 {
		t.Fatalf("matching amount rejected: %+v", got)
	}
	if got[0].Match != models.FactContradicts {
		t.Fatalf("match = %s", got[0].Match)
	}
}

func TestHeuristicQueryFromURLOnly(t *testing.T) {
	got := HeuristicQuery("https://turnbackhoax.id/2026/10/24/salah-beredar-produk-pertalite-bubuk-tinggal-campur-air/")
	if got == "" {
		t.Fatal("a URL-only audit still needs a query")
	}
	for _, want := range []string{"salah", "pertalite", "bubuk"} {
		if !strings.Contains(got, want) {
			t.Errorf("query %q missing %q", got, want)
		}
	}
	if strings.Contains(got, "2026") || strings.Contains(got, "https") {
		t.Errorf("query kept a date or scheme: %q", got)
	}
}
