package token

import (
	"errors"
	"strings"
	"testing"
)

func testKey(b byte) []byte { return []byte(strings.Repeat(string([]byte{b}), 32)) }

func sample() Payload {
	score := 92
	return Payload{
		ID:     "AMBA-20261030-7F3K",
		TS:     "2026-10-30T07:22:15Z",
		Status: "TAMPERED",
		Score:  &score,
		Title:  "Kabar bantuan yang beredar",
		Claim:  "Bantuan disalurkan minggu ini",
		Srcs:   []Source{{Pub: "Sumber Contoh", URL: "https://example.org/a", M: "CONTRADICTS"}},
	}
}

func TestRoundTrip(t *testing.T) {
	key := testKey('k')
	tok, err := Sign(key, sample())
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	got, err := Verify(key, tok)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.ID != "AMBA-20261030-7F3K" || got.Status != "TAMPERED" || got.Score == nil || *got.Score != 92 {
		t.Fatalf("payload mismatch: %+v", got)
	}
	if len(got.Srcs) != 1 || got.Srcs[0].Pub != "Sumber Contoh" {
		t.Fatalf("sources mismatch: %+v", got.Srcs)
	}
}

func TestTamperedPayloadFails(t *testing.T) {
	key := testKey('k')
	tok, _ := Sign(key, sample())
	parts := strings.Split(tok, ".")
	// Flip a character in the payload segment.
	b := []byte(parts[1])
	if b[0] == 'A' {
		b[0] = 'B'
	} else {
		b[0] = 'A'
	}
	forged := parts[0] + "." + string(b) + "." + parts[2]

	if _, err := Verify(key, forged); !errors.Is(err, ErrSignature) {
		t.Fatalf("err = %v, want ErrSignature", err)
	}
}

func TestTamperedSignatureFails(t *testing.T) {
	key := testKey('k')
	tok, _ := Sign(key, sample())
	parts := strings.Split(tok, ".")
	sig := []byte(parts[2])
	if sig[0] == 'A' {
		sig[0] = 'B'
	} else {
		sig[0] = 'A'
	}
	if _, err := Verify(key, parts[0]+"."+parts[1]+"."+string(sig)); !errors.Is(err, ErrSignature) {
		t.Fatalf("err = %v, want ErrSignature", err)
	}
}

func TestTruncatedTokenFails(t *testing.T) {
	key := testKey('k')
	tok, _ := Sign(key, sample())
	for _, cut := range []int{1, len(tok) / 2, len(tok) - 1} {
		if _, err := Verify(key, tok[:cut]); err == nil {
			t.Fatalf("truncation to %d bytes verified", cut)
		}
	}
}

func TestWrongVersionFails(t *testing.T) {
	key := testKey('k')
	tok, _ := Sign(key, sample())
	if _, err := Verify(key, "v2"+tok[2:]); !errors.Is(err, ErrVersion) {
		t.Fatalf("err = %v, want ErrVersion", err)
	}
}

func TestEmptyOrShortKeyRejected(t *testing.T) {
	if _, err := Sign(nil, sample()); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("sign with nil key: %v", err)
	}
	if _, err := Sign([]byte("short"), sample()); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("sign with short key: %v", err)
	}
	if _, err := Verify([]byte("short"), "v1.a.b"); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("verify with short key: %v", err)
	}
}

func TestOversizedPayloadRejected(t *testing.T) {
	key := testKey('k')
	p := sample()
	p.Title = strings.Repeat("x", 81)
	if _, err := Sign(key, p); !errors.Is(err, ErrPayloadSize) {
		t.Fatalf("long title: %v", err)
	}

	p = sample()
	p.Srcs = []Source{{}, {}, {}, {}}
	if _, err := Sign(key, p); !errors.Is(err, ErrPayloadSize) {
		t.Fatalf("too many sources: %v", err)
	}
}

func TestHugeTokenRejectedBeforeParsing(t *testing.T) {
	key := testKey('k')
	huge := strings.Repeat("A", MaxTokenBytes+1)
	if _, err := Verify(key, huge); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
}

func TestKeyRotation(t *testing.T) {
	oldKey, newKey := testKey('a'), testKey('b')
	tok, _ := Sign(oldKey, sample())

	if _, err := Verify(oldKey, tok); err != nil {
		t.Fatalf("old key should verify: %v", err)
	}
	if _, err := Verify(newKey, tok); !errors.Is(err, ErrSignature) {
		t.Fatalf("rotated key accepted an old token: %v", err)
	}
}

func TestMiniTokenRoundTrip(t *testing.T) {
	key := testKey('m')
	score := 12
	mini, err := SignMini(key, MiniPayload{ID: "AMBA-1", TS: "2026-10-30T07:22:15Z", Status: "SAFE", Score: &score})
	if err != nil {
		t.Fatalf("sign mini: %v", err)
	}
	got, err := VerifyMini(key, mini)
	if err != nil {
		t.Fatalf("verify mini: %v", err)
	}
	if got.ID != "AMBA-1" || got.Status != "SAFE" {
		t.Fatalf("mini mismatch: %+v", got)
	}
}

func TestInvalidTimestampRejected(t *testing.T) {
	key := testKey('k')
	p := sample()
	p.TS = "not-a-timestamp"
	if _, err := Sign(key, p); !errors.Is(err, ErrMalformed) {
		t.Fatalf("err = %v, want ErrMalformed", err)
	}
}
