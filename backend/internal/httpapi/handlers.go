package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/amba/amba/backend/internal/aidetect"
	"github.com/amba/amba/backend/internal/config"
	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/linkcheck"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/metrics"
	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/pipeline"
	"github.com/amba/amba/backend/internal/reply"
	"github.com/amba/amba/backend/internal/security"
	"github.com/amba/amba/backend/internal/token"
)

// handlers carries every dependency the routes need.
type handlers struct {
	cfg        *config.Config
	pipeline   *pipeline.Pipeline
	replies    *reply.Service
	limiter    *security.RateLimiter
	metrics    *metrics.Registry
	signingKey []byte
}

// Deps lets the server entrypoint share one pipeline between the HTTP API
// and the Telegram bot.
type Deps struct {
	Pipeline *pipeline.Pipeline
	Replies  *reply.Service
	Metrics  *metrics.Registry
	Key      []byte
}

// NewDeps builds the shared services from configuration. In development a
// random signing key is generated so share tokens still work; production
// refuses to boot without one, which the config layer enforces.
func NewDeps(cfg *config.Config) Deps {
	key := []byte(cfg.SigningKey)
	if len(key) < 32 {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			panic("httpapi: cannot read entropy for the development signing key")
		}
	}

	client := llm.New(cfg.RouterBaseURL, cfg.RouterAPIKey, cfg.RouterModel, cfg.LLMMock)
	client.Extra = cfg.LLMExtraBody
	grounder := grounding.New(grounding.DefaultAllowlist)
	p := pipeline.New(client, grounder, cfg.CacheTTL)
	p.SafeBrowsing = linkcheck.NewSafeBrowsing(cfg.SafeBrowsing)
	if cfg.SightengineUser != "" {
		p.Detector = aidetect.NewSightengine(cfg.SightengineUser, cfg.SightengineSecret)
	}
	if cfg.FactCheckKey != "" {
		p.Searcher = grounding.NewFactCheckSearcher(cfg.FactCheckKey, grounder)
	}
	return Deps{Pipeline: p, Replies: reply.New(client), Metrics: metrics.New(), Key: key}
}

// LogReadiness reports, without secrets, which optional integrations are on
// and whether the model gateway answers (TECH_SPEC §6: check /models at boot).
func LogReadiness(ctx context.Context, cfg *config.Config, d Deps) {
	reachable, listed := d.Pipeline.LLM.HealthCheck(ctx)
	slog.Info("readiness",
		"llm", d.Pipeline.LLM.Describe(),
		"llm_reachable", reachable,
		"llm_model_listed", listed,
		"llm_extra_fields", extraKeys(cfg.LLMExtraBody),
		"fact_search", d.Pipeline.Searcher != nil && d.Pipeline.Searcher.Enabled(),
		"safe_browsing", d.Pipeline.SafeBrowsing.Enabled(),
		"ai_detector", d.Pipeline.Detector != nil && d.Pipeline.Detector.Enabled(),
		"telegram", cfg.TelegramToken != "" && cfg.TelegramMode != "off",
		"metrics", cfg.MetricsAddr != "",
	)
	if !cfg.LLMMock && (!reachable || !listed) {
		slog.Warn("llm gateway not ready: audits will be degraded until it answers and lists the configured model")
	}
}

// extraKeys names the LLM_EXTRA_BODY fields in the readiness log. Only the
// names are logged, never the values.
func extraKeys(extra map[string]any) string {
	keys := make([]string, 0, len(extra))
	for k := range extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func newHandlers(cfg *config.Config, limiter *security.RateLimiter, d Deps) *handlers {
	return &handlers{
		cfg:        cfg,
		pipeline:   d.Pipeline,
		replies:    d.Replies,
		limiter:    limiter,
		metrics:    d.Metrics,
		signingKey: d.Key,
	}
}

// Sign builds the share token for a dossier.
func Sign(key []byte, d models.Dossier) (string, error) {
	srcs := make([]token.Source, 0, 3)
	for _, s := range d.Pillars.Fact.Sources {
		if len(srcs) >= 3 {
			break
		}
		srcs = append(srcs, token.Source{Pub: s.Publisher, URL: s.URL, M: string(s.Match)})
	}
	claim := truncateRunes(d.Pillars.Fact.Claim, 120)
	title := truncateRunes(d.Verdict.ShortTitle, 80)
	return token.Sign(key, token.Payload{
		ID:     d.ID,
		TS:     d.Timestamp,
		Status: string(d.Verdict.Status),
		Score:  d.Verdict.Score,
		Title:  title,
		Claim:  claim,
		Srcs:   srcs,
	})
}

func (h *handlers) sign(d models.Dossier) (string, error) { return Sign(h.signingKey, d) }

// truncateRunes cuts on a rune boundary: slicing bytes can split a multi-byte
// character and produce invalid UTF-8 in a signed payload.
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// newID returns an identifier of the form AMBA-YYYYMMDD-XXXX.
var idEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// NewID returns a fresh dossier identifier.
func NewID() string {
	buf := make([]byte, 3) // 24 bits -> 5 base32 chars, take 4
	if _, err := rand.Read(buf); err != nil {
		return "AMBA-" + time.Now().UTC().Format("20060102") + "-0000"
	}
	enc := idEncoding.EncodeToString(buf)
	if len(enc) < 4 {
		enc = strings.Repeat("0", 4)
	}
	return "AMBA-" + time.Now().UTC().Format("20060102") + "-" + enc[:4]
}

func newID() string { return NewID() }

// encodeToken is used by tests that need a valid token without a full dossier.
func encodeToken(p token.Payload, key []byte) (string, error) {
	return token.Sign(key, p)
}
