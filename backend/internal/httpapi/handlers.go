package httpapi

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"strings"
	"time"

	"github.com/amba/amba/backend/internal/config"
	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/pipeline"
	"github.com/amba/amba/backend/internal/security"
	"github.com/amba/amba/backend/internal/token"
)

// handlers carries every dependency the routes need.
type handlers struct {
	cfg        *config.Config
	pipeline   *pipeline.Pipeline
	limiter    *security.RateLimiter
	signingKey []byte
}

// newHandlers builds the handler set. In development a random key is generated
// so share tokens still work; production refuses to boot without one, which
// the config layer enforces.
func newHandlers(cfg *config.Config, limiter *security.RateLimiter) *handlers {
	key := []byte(cfg.SigningKey)
	if len(key) < 32 {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			panic("httpapi: cannot read entropy for the development signing key")
		}
	}

	client := llm.New(cfg.RouterBaseURL, cfg.RouterAPIKey, cfg.RouterModel, cfg.LLMMock)
	return &handlers{
		cfg:        cfg,
		pipeline:   pipeline.New(client, grounding.New(grounding.DefaultAllowlist), cfg.CacheTTL),
		limiter:    limiter,
		signingKey: key,
	}
}

// sign builds the share token for a dossier.
func (h *handlers) sign(d models.Dossier) (string, error) {
	srcs := make([]token.Source, 0, 3)
	for _, s := range d.Pillars.Fact.Sources {
		if len(srcs) >= 3 {
			break
		}
		srcs = append(srcs, token.Source{Pub: s.Publisher, URL: s.URL, M: string(s.Match)})
	}
	claim := d.Pillars.Fact.Claim
	if len(claim) > 120 {
		claim = claim[:120]
	}
	title := d.Verdict.ShortTitle
	if len(title) > 80 {
		title = title[:80]
	}
	return token.Sign(h.signingKey, token.Payload{
		ID:     d.ID,
		TS:     d.Timestamp,
		Status: string(d.Verdict.Status),
		Score:  d.Verdict.Score,
		Title:  title,
		Claim:  claim,
		Srcs:   srcs,
	})
}

// newID returns an identifier of the form AMBA-YYYYMMDD-XXXX.
var idEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func newID() string {
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

// encodeToken is used by tests that need a valid token without a full dossier.
func encodeToken(p token.Payload, key []byte) (string, error) {
	return token.Sign(key, p)
}

var _ = base64.RawURLEncoding
