package security

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testConfig() RateLimitConfig {
	return RateLimitConfig{
		PerMinute:     60, // 1/sec, so a burst of 5 is easy to exhaust
		Burst:         5,
		MaxInFlight:   2,
		IdleTTL:       50 * time.Millisecond,
		SweepInterval: 10 * time.Millisecond,
	}
}

func reqFrom(remote, forwarded string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = remote
	if forwarded != "" {
		r.Header.Set("X-Forwarded-For", forwarded)
	}
	return r
}

func TestBurstThenRefuse(t *testing.T) {
	rl := NewRateLimiter(testConfig())
	defer rl.Close()
	key := rl.ClientKey(reqFrom("203.0.113.9:1111", ""))

	for i := 0; i < 5; i++ {
		if ok, _ := rl.Allow(key); !ok {
			t.Fatalf("request %d refused inside the burst allowance", i+1)
		}
	}
	ok, delay := rl.Allow(key)
	if ok {
		t.Fatal("sixth request allowed, want refusal")
	}
	if delay <= 0 {
		t.Fatalf("delay = %v, want a positive retry hint", delay)
	}
}

func TestRefillAllowsAgain(t *testing.T) {
	cfg := testConfig()
	cfg.PerMinute = 600 // 10/sec
	rl := NewRateLimiter(cfg)
	defer rl.Close()
	key := rl.ClientKey(reqFrom("203.0.113.10:2222", ""))

	for i := 0; i < 5; i++ {
		rl.Allow(key)
	}
	if ok, _ := rl.Allow(key); ok {
		t.Fatal("expected the bucket to be empty")
	}
	time.Sleep(150 * time.Millisecond) // > 1 token at 10/sec
	if ok, _ := rl.Allow(key); !ok {
		t.Fatal("expected a refilled token to be allowed")
	}
}

func TestForwardedHeaderIgnoredFromUntrustedPeer(t *testing.T) {
	rl := NewRateLimiter(testConfig())
	defer rl.Close()

	a := rl.ClientKey(reqFrom("203.0.113.1:1000", "1.1.1.1"))
	b := rl.ClientKey(reqFrom("203.0.113.1:2000", "2.2.2.2"))
	if a != b {
		t.Fatal("spoofed X-Forwarded-For changed the bucket for an untrusted peer")
	}
}

func TestForwardedHeaderHonouredFromTrustedProxy(t *testing.T) {
	cfg := testConfig()
	_, proxyNet, _ := net.ParseCIDR("10.0.0.0/8")
	cfg.TrustedProxies = []*net.IPNet{proxyNet}
	rl := NewRateLimiter(cfg)
	defer rl.Close()

	a := rl.ClientKey(reqFrom("10.1.1.1:1000", "198.51.100.7"))
	b := rl.ClientKey(reqFrom("10.1.1.1:2000", "198.51.100.8"))
	if a == b {
		t.Fatal("trusted proxy header was ignored; different clients shared a bucket")
	}
}

func TestKeysAreHashedNotRaw(t *testing.T) {
	rl := NewRateLimiter(testConfig())
	defer rl.Close()
	key := rl.ClientKey(reqFrom("203.0.113.55:1234", ""))
	if len(key) != 32 {
		t.Fatalf("key length = %d, want 32 hex chars", len(key))
	}
	if key == "203.0.113.55" {
		t.Fatal("key is the raw address, not a hash")
	}
}

func TestSweepEvictsIdleClients(t *testing.T) {
	cfg := testConfig()
	rl := NewRateLimiter(cfg)
	defer rl.Close()

	key := rl.ClientKey(reqFrom("203.0.113.77:1234", ""))
	rl.Allow(key)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		rl.mu.Lock()
		_, present := rl.clients[key]
		rl.mu.Unlock()
		if !present {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("idle client was never evicted")
}

func TestInFlightCap(t *testing.T) {
	rl := NewRateLimiter(testConfig())
	defer rl.Close()

	if !rl.AcquireInFlight() || !rl.AcquireInFlight() {
		t.Fatal("first two slots should be available")
	}
	if rl.AcquireInFlight() {
		t.Fatal("third slot granted past MaxInFlight")
	}
	rl.ReleaseInFlight()
	if !rl.AcquireInFlight() {
		t.Fatal("slot not returned after release")
	}
}

func TestMiddlewareSetsRetryAfter(t *testing.T) {
	rl := NewRateLimiter(testConfig())
	defer rl.Close()

	var limitedCalls int
	handler := rl.Middleware(func(w http.ResponseWriter, _ *http.Request) {
		limitedCalls++
		w.WriteHeader(http.StatusTooManyRequests)
	})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := reqFrom("203.0.113.200:9999", "")
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d = %d, want 200", i+1, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header missing")
	}
	if limitedCalls != 1 {
		t.Fatalf("onLimited called %d times, want 1", limitedCalls)
	}
}
