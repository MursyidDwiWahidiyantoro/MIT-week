// Package reply drafts a short polite reply from a signed verdict token.
//
// The model is optional: the deterministic templates are always valid, and a
// draft only replaces them when it passes every check. The model never sees
// raw user content — only the verdict summary already in the token.
package reply

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/token"
)

// MaxChars is the contract limit.
const MaxChars = 280

// Tones are the only accepted values.
var Tones = map[string]bool{"hormat": true, "santai": true, "singkat": true}

// Source says where the text came from.
type Source string

const (
	SourceLLM      Source = "llm"
	SourceTemplate Source = "template"
)

// Draft is the API response body.
type Draft struct {
	Text   string `json:"text"`
	Source Source `json:"source"`
}

// Templates is the status x tone matrix. It mirrors configs/reply_templates.json.
var Templates = map[string]map[string]string{
	"TAMPERED": {
		"hormat":  "Terima kasih atas informasinya. Mohon maaf, sepertinya informasi ini perlu diperiksa kembali karena belum ada sumber resmi yang mendukung.",
		"santai":  "Makasih infonya. Aku cek dulu ya, soalnya belum ada sumber resminya.",
		"singkat": "Terima kasih. Sepertinya perlu dicek dulu sebelum disebarkan.",
	},
	"SUSPICIOUS": {
		"hormat":  "Terima kasih sudah berbagi. Informasi ini sebaiknya diperiksa dulu ke sumber resmi sebelum diteruskan.",
		"santai":  "Makasih share-nya. Aku cek dulu ya sebelum ikut sebarkan.",
		"singkat": "Terima kasih. Sebaiknya dicek dulu ke sumber resmi.",
	},
	"SAFE": {
		"hormat":  "Terima kasih atas informasinya. Sejauh pemeriksaan, tidak ditemukan tanda bahaya, namun tetap baik untuk memeriksa sumber resminya.",
		"santai":  "Makasih infonya. Sejauh ini nggak ada tanda bahaya, tapi tetap cek sumber resminya ya.",
		"singkat": "Terima kasih. Tidak ditemukan tanda bahaya.",
	},
	"INSUFFICIENT": {
		"hormat":  "Terima kasih atas informasinya. Mohon maaf, datanya belum cukup untuk dinilai, jadi sebaiknya diperiksa langsung ke sumber resmi.",
		"santai":  "Makasih infonya. Datanya belum cukup buat dinilai, mending cek langsung ke sumbernya.",
		"singkat": "Terima kasih. Datanya belum cukup untuk dinilai.",
	},
}

// Template returns the deterministic fallback for a payload.
func Template(p token.Payload, tone string) string {
	status := p.Status
	if _, ok := Templates[status]; !ok {
		status = "INSUFFICIENT"
	}
	if !Tones[tone] {
		tone = "hormat"
	}
	return Templates[status][tone]
}

// Service drafts replies with a cache.
type Service struct {
	Client  *llm.Client
	AppURL  string
	Timeout time.Duration

	mu    sync.Mutex
	cache map[string]cacheEntry
}

// New builds the service.
func New(client *llm.Client, appURL string) *Service {
	return &Service{
		Client:  client,
		AppURL:  strings.TrimRight(appURL, "/"),
		Timeout: 7 * time.Second,
		cache:   make(map[string]cacheEntry),
	}
}

type cacheEntry struct {
	draft   Draft
	expires time.Time
}

// Draft produces a reply. It never returns an error: a failure is a template.
func (s *Service) Draft(ctx context.Context, p token.Payload, tone, shareToken string) Draft {
	key := cacheKey(shareToken, tone)
	if d, ok := s.fromCache(key); ok {
		return d
	}

	fallback := Draft{Text: Template(p, tone), Source: SourceTemplate}
	if s == nil || s.Client == nil || s.Client.Mock || s.Client.APIKey == "" {
		return fallback
	}

	allowed := make(map[string]bool, len(p.Srcs)+1)
	for _, src := range p.Srcs {
		allowed[src.URL] = true
	}
	shareLink := ""
	if s.AppURL != "" && shareToken != "" {
		shareLink = s.AppURL + "/hasil#t=" + shareToken
		allowed[shareLink] = true
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	text, err := s.Client.DraftReply(ctx, buildPrompt(p, tone, allowed))
	if err != nil {
		return fallback
	}
	cleaned, ok := ValidateTone(text, allowed, tone)
	if !ok {
		return fallback
	}
	cleaned = appendShareLink(cleaned, shareLink, tone)
	if cleaned == "" {
		return fallback
	}
	draft := Draft{Text: cleaned, Source: SourceLLM}
	s.toCache(key, draft)
	return draft
}

func buildPrompt(p token.Payload, tone string, allowed map[string]bool) string {
	var b strings.Builder
	switch tone {
	case "hormat":
		b.WriteString("Nada HORMAT: sapa dengan sopan (misalnya \"Mohon izin berbagi\"), pakai kata baku, tanpa kata santai seperti \"makasih\", \"nggak\", \"kayaknya\", \"aku\".\n")
	case "santai":
		b.WriteString("Nada SANTAI: akrab dan ringkas, boleh memakai \"aku\" dan \"makasih\", tetap sopan, tanpa \"Bapak/Ibu\".\n")
	case "singkat":
		b.WriteString("Nada SINGKAT: satu baris saja, langsung ke inti, tanpa sapaan panjang.\n")
	}
	b.WriteString("Buat draf balasan untuk pesan yang hasil pemeriksaannya:\n")
	fmt.Fprintf(&b, "- Status: %s\n", p.Status)
	if p.Score != nil {
		fmt.Fprintf(&b, "- Skor risiko: %d dari 100 (indikatif, bukan kepastian)\n", *p.Score)
	}
	if p.Title != "" {
		fmt.Fprintf(&b, "- Judul hasil: %s\n", p.Title)
	}
	if p.Claim != "" {
		fmt.Fprintf(&b, "- Klaim yang diperiksa: %s\n", p.Claim)
	}
	fmt.Fprintf(&b, "- Nada: %s\n", tone)
	b.WriteString("- JANGAN menulis URL, tautan, atau alamat apa pun. Tautan akan ditambahkan otomatis oleh sistem.\n")
	b.WriteString("\nBalas JSON {\"text\":\"...\"}.")
	return b.String()
}

var (
	urlPattern      = regexp.MustCompile(`https?://[^\s<>"')\]]+`)
	certaintyWords  = []string{"pasti", "terbukti", "100%", "dijamin", "jaminan", "tidak diragukan", "mutlak"}
	shamingPatterns = []string{"bodoh", "goblok", "tolol", "dungu", "idiot", "bego", "sinting", "gila kamu"}
)

// Validate enforces the PRD F12 rules. It returns the cleaned text and whether
// the draft is usable at all. An empty tone skips the tone check (used by
// tests that only exercise URL and certainty rules).
func Validate(text string, allowed map[string]bool) (string, bool) {
	return ValidateTone(text, allowed, "")
}

// ValidateTone is Validate plus the register check for a specific tone.
func ValidateTone(text string, allowed map[string]bool, tone string) (string, bool) {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return "", false
	}
	lower := strings.ToLower(text)
	for _, w := range certaintyWords {
		if strings.Contains(lower, w) {
			return "", false
		}
	}
	for _, w := range shamingPatterns {
		if strings.Contains(lower, w) {
			return "", false
		}
	}
	if tone != "" && !toneMatches(tone, lower) {
		return "", false
	}

	// Strip every URL that is not in the allowed set. A model that invents a
	// citation must not get one into a message the user sends to a relative.
	text = urlPattern.ReplaceAllStringFunc(text, func(raw string) string {
		clean := strings.TrimRight(raw, ".,;:!?")
		if allowed[clean] {
			return clean
		}
		return ""
	})
	text = stripDanglingURLContext(strings.Join(strings.Fields(text), " "))
	if len([]rune(text)) == 0 {
		return "", false
	}
	if len([]rune(text)) > MaxChars {
		text = truncateSafely(text, MaxChars)
	}
	return text, true
}

// appendShareLink adds the signed link after the model's sentence. A model
// asked to include a long token reliably shortens it, and a shortened token
// is a broken report, so the link is never model-written.
func appendShareLink(text, shareLink, tone string) string {
	text = strings.TrimSpace(text)
	if shareLink == "" || tone == "singkat" {
		return text
	}
	sep := " Detailnya: "
	if !strings.HasSuffix(text, ".") && !strings.HasSuffix(text, "!") && !strings.HasSuffix(text, "?") {
		text += "."
	}

	// Trim the body, not the link: a cut token is a broken report, a cut
	// sentence is merely shorter.
	budget := MaxChars - len([]rune(sep+shareLink))
	if budget <= 0 {
		return text
	}
	if len([]rune(text)) > budget {
		text = truncateSafely(text, budget)
	}
	return text + sep + shareLink
}

// danglingWords are connectives that only make sense before a link. When the
// link is stripped, "cek detailnya di" becomes "cek detailnya".
var danglingWords = map[string]bool{
	"di": true, "ke": true, "dari": true, "pada": true, "lewat": true,
	"via": true, "lihat": true, "buka": true, "kunjungi": true, "sumbernya": true,
}

func stripDanglingURLContext(text string) string {
	fields := strings.Fields(text)
	for len(fields) > 0 {
		last := strings.ToLower(strings.Trim(fields[len(fields)-1], " .,;:!?-"))
		if !danglingWords[last] {
			break
		}
		fields = fields[:len(fields)-1]
	}
	out := strings.Join(fields, " ")
	out = strings.TrimRight(out, " ,;:-")
	if out != "" && !strings.HasSuffix(out, ".") && !strings.HasSuffix(out, "!") && !strings.HasSuffix(out, "?") {
		out += "."
	}
	return strings.TrimSpace(out)
}

// truncateSafely cuts at a word boundary and never leaves half a URL behind.
// A 280-char cap that slices "http://localhost/hasil#t=v1.eyJ..." produces a
// broken link in a message the user is about to send to a relative.
func truncateSafely(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	cut := strings.TrimSpace(string(runes[:limit]))
	if idx := strings.LastIndex(cut, " "); idx > 0 {
		cut = cut[:idx]
	}
	// Drop a trailing token that looks like a URL but has no complete match.
	fields := strings.Fields(cut)
	for len(fields) > 0 {
		last := fields[len(fields)-1]
		if strings.HasPrefix(strings.ToLower(last), "http://") || strings.HasPrefix(strings.ToLower(last), "https://") {
			if urlPattern.MatchString(last) && strings.HasSuffix(last, lastURLSuffix(last)) {
				break
			}
			fields = fields[:len(fields)-1]
			continue
		}
		break
	}
	out := strings.Join(fields, " ")
	out = strings.TrimRight(out, " ,;:-")
	return strings.TrimSpace(out)
}

// lastURLSuffix returns the final character of a token, used to check that a
// URL was not cut by the length limit.
func lastURLSuffix(s string) string {
	if s == "" {
		return ""
	}
	return s[len(s)-1:]
}

var (
	// " sih " with spaces: bare "sih" also occurs inside "kasih" and would
	// reject every polite "terima kasih" draft.
	casualMarkers = []string{"makasih", "nggak", "gak ", "kayaknya", "aku ", "banget", " sih "}
	formalMarkers = []string{"bapak/ibu", "mohon izin", "dengan hormat", "terima kasih atas"}
)

// toneMatches keeps a draft from claiming a register it did not write in.
// The model reliably ignores the tone line unless the output is checked.
func toneMatches(tone, lower string) bool {
	switch tone {
	case "hormat":
		for _, w := range casualMarkers {
			if strings.Contains(lower, w) {
				return false
			}
		}
		return true
	case "santai":
		for _, w := range formalMarkers {
			if strings.Contains(lower, w) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

func cacheKey(token, tone string) string {
	sum := sha256.Sum256([]byte(token + "\x00" + tone))
	return hex.EncodeToString(sum[:])
}

func (s *Service) fromCache(key string) (Draft, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.cache[key]
	if !ok || time.Now().After(e.expires) {
		delete(s.cache, key)
		return Draft{}, false
	}
	return e.draft, true
}

func (s *Service) toCache(key string, d Draft) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.cache) > 500 {
		s.cache = make(map[string]cacheEntry)
	}
	s.cache[key] = cacheEntry{draft: d, expires: time.Now().Add(10 * time.Minute)}
}
