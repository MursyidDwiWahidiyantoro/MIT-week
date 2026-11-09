package aidetect

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func server(t *testing.T, status int, body string, check func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
}

func client(url string) *Sightengine {
	s := NewSightengine("user", "secret")
	s.Endpoint = url
	return s
}

func TestScoreSendsTheImageAndReadsTheProbability(t *testing.T) {
	srv := server(t, 200, `{"status":"success","request":{"id":"req_1"},"type":{"ai_generated":0.987},"media":{"id":"med_1"}}`, func(r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("not multipart: %v", err)
		}
		if r.FormValue("models") != "genai" || r.FormValue("api_user") != "user" || r.FormValue("api_secret") != "secret" {
			t.Errorf("form = %v", r.MultipartForm.Value)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("credentials must not travel in the URL: %q", r.URL.RawQuery)
		}
		f, h, err := r.FormFile("media")
		if err != nil {
			t.Fatalf("no media part: %v", err)
		}
		data, _ := io.ReadAll(f)
		if string(data) != "\xff\xd8\xffimage" || h.Filename != "image" || h.Header.Get("Content-Type") != "image/jpeg" {
			t.Errorf("media = %q %q %q", data, h.Filename, h.Header.Get("Content-Type"))
		}
	})
	defer srv.Close()

	got, err := client(srv.URL).Score(context.Background(), []byte("\xff\xd8\xffimage"), "image/jpeg")
	if err != nil || got != 99 {
		t.Fatalf("Score = %d, %v", got, err)
	}
}

func TestScoreFailsClosed(t *testing.T) {
	for name, tc := range map[string]struct {
		status int
		body   string
	}{
		"quota":           {429, `{"status":"failure","error":{"type":"usage_limit","code":32}}`},
		"bad credentials": {401, `{"status":"failure"}`},
		"failure status":  {200, `{"status":"failure","error":{"message":"x"}}`},
		"no probability":  {200, `{"status":"success","type":{}}`},
		"not json":        {200, `<html>`},
	} {
		t.Run(name, func(t *testing.T) {
			srv := server(t, tc.status, tc.body, nil)
			defer srv.Close()
			if _, err := client(srv.URL).Score(context.Background(), []byte("x"), "image/png"); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestScoreClampsAndTimesOut(t *testing.T) {
	srv := server(t, 200, `{"status":"success","type":{"ai_generated":1.7}}`, nil)
	defer srv.Close()
	if got, err := client(srv.URL).Score(context.Background(), []byte("x"), ""); err != nil || got != 100 {
		t.Fatalf("Score = %d, %v", got, err)
	}

	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer slow.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := client(slow.URL).Score(ctx, []byte("x"), "image/png"); err == nil {
		t.Fatal("a timeout must fail")
	}
}

func TestDisabledWithoutCredentials(t *testing.T) {
	var s *Sightengine
	if s.Enabled() || NewSightengine("", "x").Enabled() || NewSightengine("x", "").Enabled() {
		t.Fatal("must be disabled without both credentials")
	}
	if !NewSightengine("u", "s").Enabled() {
		t.Fatal("must be enabled with credentials")
	}
	if _, err := NewSightengine("", "").Score(context.Background(), []byte("x"), ""); err == nil {
		t.Fatal("a disabled client must not report a score")
	}
}
