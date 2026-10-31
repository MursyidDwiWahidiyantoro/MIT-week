// Package token signs and verifies the share tokens carried in a URL fragment.
//
// Format: "v1." + base64url(JSON payload) + "." + base64url(HMAC-SHA256).
// The payload is readable by anyone with the link — it must never contain
// user input, only the verdict summary needed to render a shared report.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Version is the only accepted token prefix.
const Version = "v1"

// MaxTokenBytes bounds a token so a huge payload cannot be used to exhaust
// memory during verification.
const MaxTokenBytes = 2 << 10

// MaxSourceCount caps embedded sources (TECH_SPEC §7.5).
const MaxSourceCount = 3

// Errors are deliberately coarse: the UI shows one generic invalid state.
var (
	ErrMalformed   = errors.New("token malformed")
	ErrSignature   = errors.New("token signature invalid")
	ErrVersion     = errors.New("token version unsupported")
	ErrTooLarge    = errors.New("token too large")
	ErrPayloadSize = errors.New("token payload exceeds limits")
	ErrEmptyKey    = errors.New("signing key missing")
)

// Source is a verified publisher reference embedded in a share token.
type Source struct {
	Pub string `json:"pub"`
	URL string `json:"url"`
	M   string `json:"m"`
}

// Payload is the signed share summary.
type Payload struct {
	ID     string   `json:"id"`
	TS     string   `json:"ts"`
	Status string   `json:"status"`
	Score  *int     `json:"score"`
	Title  string   `json:"title"`
	Claim  string   `json:"claim,omitempty"`
	Srcs   []Source `json:"srcs,omitempty"`
}

// MiniPayload is the short form encoded in a QR code.
type MiniPayload struct {
	ID     string `json:"id"`
	TS     string `json:"ts"`
	Status string `json:"status"`
	Score  *int   `json:"score"`
}

var enc = base64.RawURLEncoding

// Sign encodes and signs a payload. The key must be at least 32 bytes.
func Sign(key []byte, p Payload) (string, error) {
	if len(key) < 32 {
		return "", ErrEmptyKey
	}
	if err := validatePayload(p); err != nil {
		return "", err
	}
	body, err := json.Marshal(p)
	if err != nil {
		return "", ErrMalformed
	}
	encoded := enc.EncodeToString(body)
	signed := Version + "." + encoded
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signed))
	out := signed + "." + enc.EncodeToString(mac.Sum(nil))
	if len(out) > MaxTokenBytes {
		return "", ErrPayloadSize
	}
	return out, nil
}

// SignMini encodes the compact QR payload.
func SignMini(key []byte, p MiniPayload) (string, error) {
	if len(key) < 32 {
		return "", ErrEmptyKey
	}
	body, err := json.Marshal(p)
	if err != nil {
		return "", ErrMalformed
	}
	encoded := enc.EncodeToString(body)
	signed := Version + "." + encoded
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signed))
	return signed + "." + enc.EncodeToString(mac.Sum(nil)), nil
}

// Verify checks the signature and returns the payload.
func Verify(key []byte, raw string) (Payload, error) {
	var p Payload
	if len(key) < 32 {
		return p, ErrEmptyKey
	}
	if len(raw) > MaxTokenBytes {
		return p, ErrTooLarge
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return p, ErrMalformed
	}
	if parts[0] != Version {
		return p, ErrVersion
	}
	signed := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signed))
	want := mac.Sum(nil)

	got, err := enc.DecodeString(parts[2])
	if err != nil {
		return p, ErrMalformed
	}
	if !hmac.Equal(got, want) {
		return p, ErrSignature
	}
	body, err := enc.DecodeString(parts[1])
	if err != nil {
		return p, ErrMalformed
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return p, ErrMalformed
	}
	if err := validatePayload(p); err != nil {
		return Payload{}, err
	}
	return p, nil
}

// VerifyMini is Verify for the QR payload.
func VerifyMini(key []byte, raw string) (MiniPayload, error) {
	var m MiniPayload
	if len(key) < 32 {
		return m, ErrEmptyKey
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 3 || parts[0] != Version {
		return m, ErrMalformed
	}
	signed := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signed))
	got, err := enc.DecodeString(parts[2])
	if err != nil {
		return m, ErrMalformed
	}
	if !hmac.Equal(got, mac.Sum(nil)) {
		return m, ErrSignature
	}
	body, err := enc.DecodeString(parts[1])
	if err != nil {
		return m, ErrMalformed
	}
	if err := json.Unmarshal(body, &m); err != nil {
		return m, ErrMalformed
	}
	return m, nil
}

func validatePayload(p Payload) error {
	if len(p.Title) > 80 || len(p.Claim) > 120 {
		return ErrPayloadSize
	}
	if len(p.Srcs) > MaxSourceCount {
		return ErrPayloadSize
	}
	if p.TS != "" {
		if _, err := time.Parse(time.RFC3339, p.TS); err != nil {
			return ErrMalformed
		}
	}
	return nil
}
