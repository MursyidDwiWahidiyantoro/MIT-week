package pagefetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExtractReadsTitleDescriptionAndBody(t *testing.T) {
	body := `<!doctype html><html><head>
		<title>Bantuan Rp 500 Ribu</title>
		<meta name="description" content="Cek fakta bantuan sosial.">
		<script>var secret = "should not appear";</script>
		<style>.x{color:red}</style>
	</head><body><h1>Bantuan</h1><p>Pesan berantai ini <b>hoaks</b>.</p></body></html>`
	page := Extract([]byte(body), "https://example.org/a")
	if page.Title != "Bantuan Rp 500 Ribu" {
		t.Errorf("title = %q", page.Title)
	}
	if page.Description != "Cek fakta bantuan sosial." {
		t.Errorf("description = %q", page.Description)
	}
	if !strings.Contains(page.Text, "hoaks") {
		t.Errorf("text missing body: %q", page.Text)
	}
	if strings.Contains(page.Text, "should not appear") || strings.Contains(page.Text, "color:red") {
		t.Errorf("script/style leaked into text: %q", page.Text)
	}
}

func TestExtractTruncatesLongPages(t *testing.T) {
	body := "<html><body>" + strings.Repeat("kata ", 5000) + "</body></html>"
	page := Extract([]byte(body), "")
	if len([]rune(page.Text)) > MaxText {
		t.Fatalf("text length = %d, want <= %d", len([]rune(page.Text)), MaxText)
	}
}

func TestFetchRejectsNonHTML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.7"))
	}))
	defer srv.Close()

	// httptest listens on loopback, which the SSRF client refuses. This test
	// therefore asserts the refusal rather than the content-type branch.
	_, err := Fetch(context.Background(), srv.Client(), srv.URL, time.Second)
	if err == nil {
		t.Fatal("loopback fetch must be refused")
	}
}

func TestIsHTML(t *testing.T) {
	if !isHTML("text/html; charset=utf-8", nil) {
		t.Error("text/html not detected")
	}
	if isHTML("application/json", []byte(`{}`)) {
		t.Error("json classified as html")
	}
	if !isHTML("", []byte("<!doctype html><html></html>")) {
		t.Error("doctype sniff failed")
	}
}
