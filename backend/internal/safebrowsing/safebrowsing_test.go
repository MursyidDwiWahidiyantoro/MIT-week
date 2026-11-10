package safebrowsing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMissingKeyIsNotConfigured(t *testing.T) {
	c := New("")
	if _, err := c.Check(context.Background(), "https://example.org"); err != ErrNotConfigured {
		t.Fatalf("err = %v, want ErrNotConfigured", err)
	}
}

func TestHit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "k" {
			t.Errorf("key not sent")
		}
		var body struct {
			ThreatInfo struct {
				ThreatEntries []struct {
					URL string `json:"url"`
				} `json:"threatEntries"`
			} `json:"threatInfo"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if len(body.ThreatInfo.ThreatEntries) != 1 || body.ThreatInfo.ThreatEntries[0].URL != "https://evil.example" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"matches":[{"threatType":"MALWARE"}]}`))
	}))
	defer srv.Close()

	c := New("k")
	c.Endpoint = srv.URL
	hit, err := c.Check(context.Background(), "https://evil.example")
	if err != nil || !hit {
		t.Fatalf("hit=%v err=%v", hit, err)
	}
}

func TestMiss(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New("k")
	c.Endpoint = srv.URL
	hit, err := c.Check(context.Background(), "https://example.org")
	if err != nil || hit {
		t.Fatalf("hit=%v err=%v", hit, err)
	}
}

func TestAPIErrorIsNotSafe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := New("k")
	c.Endpoint = srv.URL
	if _, err := c.Check(context.Background(), "https://example.org"); err == nil {
		t.Fatal("an API error must not be reported as safe")
	}
}

func TestErrorDoesNotContainKey(t *testing.T) {
	c := New("super-secret-key")
	c.Endpoint = "http://127.0.0.1:1"
	_, err := c.Check(context.Background(), "https://example.org")
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "super-secret-key") {
		t.Fatalf("error leaked the key: %v", err)
	}
}
