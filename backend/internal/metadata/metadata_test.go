package metadata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// testResolver points every oEmbed endpoint at srv. Tests use a plain client:
// the SSRF-safe client refuses loopback, which is its own tested behaviour.
func testResolver(srv *httptest.Server) *Resolver {
	r := New(&http.Client{Timeout: 2 * time.Second})
	r.OEmbed = map[string]string{
		"youtube": srv.URL + "/oembed/youtube?url=",
		"tiktok":  srv.URL + "/oembed/tiktok?url=",
		"x":       srv.URL + "/oembed/x?url=",
	}
	return r
}

func TestPlatform(t *testing.T) {
	cases := map[string]string{
		"www.youtube.com": "youtube", "youtu.be": "youtube", "m.youtube.com": "youtube",
		"www.tiktok.com": "tiktok", "vt.tiktok.com": "tiktok",
		"x.com": "x", "twitter.com": "x", "mobile.twitter.com": "x",
		"www.instagram.com": "instagram", "fb.watch": "facebook", "www.facebook.com": "facebook",
		"www.threads.net": "threads", "kompas.com": "web", "notyoutube.com": "web",
	}
	for host, want := range cases {
		if got := Platform(host); got != want {
			t.Errorf("Platform(%q) = %q, want %q", host, got, want)
		}
	}
}

func TestOEmbedTitleAndAuthor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oembed/tiktok" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"title":"Bansos Rp 500.000 cair hari ini!! Klik link di bio","author_name":"info.bansos","provider_name":"TikTok"}`))
	}))
	defer srv.Close()

	// The oEmbed endpoint is asked about the original URL; the platform is
	// decided from the URL's host, not the endpoint's.
	r := testResolver(srv)
	m := r.Resolve(context.Background(), "https://www.tiktok.com/@info.bansos/video/1")
	if m.Status != StatusOK || m.Via != "oembed" || m.Platform != "tiktok" {
		t.Fatalf("got %+v", m)
	}
	if m.Title != "Bansos Rp 500.000 cair hari ini!! Klik link di bio" || m.Author != "info.bansos" || m.SiteName != "TikTok" {
		t.Fatalf("fields: %+v", m)
	}
	if m.Caption() != m.Title {
		t.Fatalf("caption = %q", m.Caption())
	}
	if !strings.Contains(m.ForModel("https://www.tiktok.com/@info.bansos/video/1"), "Pengunggah: info.bansos") {
		t.Fatal("model text misses the author line")
	}
}

func TestXOEmbedUsesPostText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"author_name":"Akun Contoh","html":"<blockquote class=\"twitter-tweet\"><p lang=\"in\" dir=\"ltr\">Sebarkan! Besok listrik padam <a href=\"https://t.co/x\">pic</a> se-Jawa</p>&mdash; Akun</blockquote>"}`))
	}))
	defer srv.Close()

	m := testResolver(srv).Resolve(context.Background(), "https://x.com/akun/status/1")
	if m.Status != StatusOK || m.Description != "Sebarkan! Besok listrik padam pic se-Jawa" {
		t.Fatalf("got %+v", m)
	}
}

func TestOpenGraphFallback(t *testing.T) {
	page := `<!doctype html><html><head>
<title>Fallback title</title>
<meta property="og:title" content="Pemerintah &amp; BUMN bagi hadiah">
<meta name="description" content='Isi   formulir
 sekarang juga.'>
<meta property="og:site_name" content=Contoh>
</head><body><meta property="og:title" content="body tags are ignored"></body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}))
	defer srv.Close()

	m := testResolver(srv).Resolve(context.Background(), srv.URL+"/artikel")
	if m.Status != StatusOK || m.Via != "opengraph" || m.Platform != "web" {
		t.Fatalf("got %+v", m)
	}
	if m.Title != "Pemerintah & BUMN bagi hadiah" || m.Description != "Isi formulir sekarang juga." || m.SiteName != "Contoh" {
		t.Fatalf("fields: %+v", m)
	}
	if m.Caption() != "Pemerintah & BUMN bagi hadiah\nIsi formulir sekarang juga." {
		t.Fatalf("caption = %q", m.Caption())
	}
}

func TestBlockedPlatforms(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"forbidden", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "no", http.StatusForbidden) }},
		{"rate limited", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "slow down", http.StatusTooManyRequests) }},
		{"login redirect", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/accounts/login/" {
				_, _ = w.Write([]byte(`<html><head><title>Login</title></head></html>`))
				return
			}
			http.Redirect(w, r, "/accounts/login/", http.StatusFound)
		}},
		{"generic login page", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><head><title>Instagram</title></head></html>`))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()
			r := testResolver(srv)
			// Pretend the test server is Instagram by resolving its URL through
			// a client that rewrites the host.
			r.Client = rewriteTo(srv)
			m := r.Resolve(context.Background(), "https://www.instagram.com/p/abc/")
			if m.Status != StatusBlocked || m.Platform != "instagram" {
				t.Fatalf("got %+v", m)
			}
		})
	}
}

func TestUnavailable(t *testing.T) {
	notHTML := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 'P', 'N', 'G'})
	}))
	defer notHTML.Close()
	if m := testResolver(notHTML).Resolve(context.Background(), notHTML.URL+"/a.png"); m.Status != StatusUnavailable {
		t.Fatalf("image url: %+v", m)
	}

	gone := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) }))
	defer gone.Close()
	if m := testResolver(gone).Resolve(context.Background(), gone.URL+"/x"); m.Status != StatusUnavailable {
		t.Fatalf("404: %+v", m)
	}

	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><head></head><body>hi</body></html>`))
	}))
	defer empty.Close()
	if m := testResolver(empty).Resolve(context.Background(), empty.URL); m.Status != StatusUnavailable {
		t.Fatalf("no tags: %+v", m)
	}

	if m := testResolver(empty).Resolve(context.Background(), "::not a url"); m.Status != StatusUnavailable {
		t.Fatalf("bad url: %+v", m)
	}
}

func TestTimeoutIsUnavailable(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer slow.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if m := testResolver(slow).Resolve(ctx, slow.URL); m.Status != StatusUnavailable {
		t.Fatalf("got %+v", m)
	}
}

func TestCleanTruncatesOnRuneBoundary(t *testing.T) {
	long := strings.Repeat("é", maxTitle+20)
	got := clean(long, maxTitle)
	if n := len([]rune(got)); n != maxTitle {
		t.Fatalf("rune length = %d", n)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatal("truncation should be marked")
	}
	if clean("a\x00b​c", 10) != "ab​c" {
		t.Fatalf("control characters must go: %q", clean("a\x00b​c", 10))
	}
}

// rewriteTo returns a client that sends every request to srv, whatever host
// the URL names, so a test server can stand in for a real platform.
func rewriteTo(srv *httptest.Server) *http.Client {
	return &http.Client{
		Timeout: 2 * time.Second,
		Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
			r2 := req.Clone(req.Context())
			r2.URL.Scheme = "http"
			r2.URL.Host = strings.TrimPrefix(srv.URL, "http://")
			r2.Host = r2.URL.Host
			return http.DefaultTransport.RoundTrip(r2)
		}),
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
