// Package reply drafts Balas Sopan messages (PRD F12, TECH_SPEC §6.2).
//
// Input is only the verified share-token payload, never raw user content.
// A model draft is used when it passes validation; otherwise the
// deterministic template for that status and tone is returned.
package reply

import (
	"container/list"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/token"
)

// MaxLength is the PRD F12 limit.
const MaxLength = 280

// Tones are the three accepted tones.
var Tones = []string{"hormat", "santai", "singkat"}

//go:embed reply_templates.json
var templatesJSON []byte

// templates is configs/reply_templates.json, embedded. tools/check-demo-diff.sh
// keeps the config, this copy and the frontend copy byte-identical.
var templates = mustTemplates()

func mustTemplates() map[string]map[string]string {
	var doc struct {
		Templates map[string]map[string]string `json:"templates"`
	}
	if err := json.Unmarshal(templatesJSON, &doc); err != nil {
		panic("reply: invalid embedded templates: " + err.Error())
	}
	for _, status := range []string{"TAMPERED", "SUSPICIOUS", "SAFE", "INSUFFICIENT"} {
		for _, tone := range Tones {
			if doc.Templates[status][tone] == "" {
				panic("reply: missing template " + status + "/" + tone)
			}
		}
	}
	return doc.Templates
}

// Template returns the deterministic reply for a status and tone.
func Template(status, tone string) string {
	switch status {
	case "TAMPERED", "SUSPICIOUS", "SAFE", "INSUFFICIENT":
	default:
		status = "INSUFFICIENT"
	}
	return templates[status][tone]
}

// ValidTone reports whether tone is one of Tones.
func ValidTone(tone string) bool {
	for _, t := range Tones {
		if t == tone {
			return true
		}
	}
	return false
}

// Service drafts replies and caches them by SHA-256(token+tone) for 10 min.
type Service struct {
	LLM     *llm.Client
	Timeout time.Duration
	TTL     time.Duration
	Max     int

	mu    sync.Mutex
	cache map[string]*list.Element
	order *list.List
}

type entry struct {
	key, text, source string
	expires           time.Time
}

// New builds a service.
func New(client *llm.Client) *Service {
	return &Service{
		LLM:     client,
		Timeout: 4 * time.Second,
		TTL:     10 * time.Minute,
		Max:     500,
		cache:   map[string]*list.Element{},
		order:   list.New(),
	}
}

// Draft returns the reply text and its source ("llm" or "template").
func (s *Service) Draft(ctx context.Context, rawToken string, p token.Payload, tone string) (string, string) {
	key := cacheKey(rawToken, tone)
	if text, source, ok := s.get(key); ok {
		return text, source
	}

	text, source := Template(p.Status, tone), "template"
	if s.LLM != nil {
		allowed := map[string]bool{}
		sources := make([]string, 0, len(p.Srcs))
		for _, src := range p.Srcs {
			if src.URL == "" {
				continue
			}
			allowed[src.URL] = true
			sources = append(sources, strings.TrimSpace(src.Pub+" "+src.URL))
		}
		dctx, cancel := context.WithTimeout(ctx, s.Timeout)
		draft, err := s.LLM.Reply(dctx, llm.ReplyInput{
			Status: p.Status, Title: p.Title, Claim: p.Claim, Sources: sources, Tone: tone,
		})
		cancel()
		if err == nil {
			if cleaned, ok := Validate(draft, allowed); ok {
				text, source = cleaned, "llm"
			}
		}
	}
	s.put(key, text, source)
	return text, source
}

var (
	urlPattern = regexp.MustCompile(`https?://[^\s<>"')\]]+`)
	// banned covers certainty claims, scores and shaming (TECH_SPEC §6.2).
	banned = regexp.MustCompile(`(?i)(\bpasti\b|100\s?%|\bterbukti\b|\bdijamin\b|\bbohong\b|\bpembohong\b|\bbodoh\b|\btolol\b|\bgoblok\b|\bdungu\b|\bhoaxer\b|penyebar hoa(ks|x)|\bskor\b|\bscore\b|\b\d{1,3}\s?/\s?100\b|kamu (yang )?salah|anda (yang )?salah)`)
	spaces = regexp.MustCompile(`[ \t]+`)
)

// Validate applies the F12 rules to a draft. URLs outside allowed are
// stripped; the draft is rejected if it then breaks any rule or is empty.
func Validate(draft string, allowed map[string]bool) (string, bool) {
	text := strings.TrimSpace(draft)
	text = urlPattern.ReplaceAllStringFunc(text, func(u string) string {
		trimmed := strings.TrimRight(u, ".,;:!?")
		if allowed[trimmed] {
			return u
		}
		return ""
	})
	text = strings.TrimSpace(spaces.ReplaceAllString(text, " "))
	text = strings.ReplaceAll(text, " .", ".")
	text = strings.ReplaceAll(text, " ,", ",")
	switch {
	case utf8.RuneCountInString(text) < 10:
		return "", false
	case utf8.RuneCountInString(text) > MaxLength:
		return "", false
	case banned.MatchString(text):
		return "", false
	case strings.Contains(text, "<") || strings.Contains(text, ">"):
		return "", false
	}
	return text, true
}

func cacheKey(rawToken, tone string) string {
	sum := sha256.Sum256([]byte(rawToken + "\x00" + tone))
	return hex.EncodeToString(sum[:])
}

func (s *Service) get(key string) (string, string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	el, ok := s.cache[key]
	if !ok {
		return "", "", false
	}
	e := el.Value.(*entry)
	if time.Now().After(e.expires) {
		s.order.Remove(el)
		delete(s.cache, key)
		return "", "", false
	}
	s.order.MoveToFront(el)
	return e.text, e.source, true
}

func (s *Service) put(key, text, source string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if el, ok := s.cache[key]; ok {
		s.order.Remove(el)
	}
	s.cache[key] = s.order.PushFront(&entry{key: key, text: text, source: source, expires: time.Now().Add(s.TTL)})
	for s.order.Len() > s.Max {
		back := s.order.Back()
		s.order.Remove(back)
		delete(s.cache, back.Value.(*entry).key)
	}
}
