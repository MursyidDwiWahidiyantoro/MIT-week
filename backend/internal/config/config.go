// Package config loads and validates environment configuration.
// It fails fast: a bad or missing required value stops the server at boot
// rather than surfacing as a confusing runtime error later.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the fully validated runtime configuration.
type Config struct {
	Host          string
	Port          int
	Environment   string
	PublicAppURL  string
	TrustedProxy  []string
	RouterBaseURL string
	RouterAPIKey  string
	RouterModel   string
	Reasoning     string
	Retrieval     bool
	RetrievalURL  string
	LLMMock       bool
	SigningKey    string
	SafeBrowsing  string
	MaxUploadMB   int64
	HardTimeout   time.Duration
	CacheTTL      time.Duration
	MetricsAddr   string
}

// IsProduction reports whether the server is running in production mode.
func (c *Config) IsProduction() bool { return c.Environment == "production" }

// Addr returns the listen address.
func (c *Config) Addr() string { return fmt.Sprintf("%s:%d", c.Host, c.Port) }

// Load reads configuration from the environment and validates it.
// Every error returned here is safe to print at boot: it names the variable,
// never its value.
func Load() (*Config, error) {
	c := &Config{
		Host:          env("SERVER_HOST", "127.0.0.1"),
		Environment:   env("ENVIRONMENT", "development"),
		PublicAppURL:  env("PUBLIC_APP_URL", "http://localhost:5173"),
		RouterBaseURL: env("NINEROUTER_BASE_URL", "http://localhost:20128/v1"),
		RouterAPIKey:  os.Getenv("NINEROUTER_API_KEY"),
		RouterModel:   os.Getenv("NINEROUTER_MODEL"),
		Reasoning:     env("NINEROUTER_REASONING_EFFORT", "low"),
		RetrievalURL:  env("FACT_SEARCH_URL", "https://www.bing.com/news/search"),
		SigningKey:    os.Getenv("REPORT_SIGNING_KEY"),
		SafeBrowsing:  os.Getenv("GOOGLE_SAFE_BROWSING_KEY"),
		MetricsAddr:   os.Getenv("METRICS_ADDR"),
	}

	var errs []error

	port, err := strconv.Atoi(env("SERVER_PORT", "8080"))
	if err != nil || port < 1 || port > 65535 {
		errs = append(errs, errors.New("SERVER_PORT: must be an integer between 1 and 65535"))
	}
	c.Port = port

	c.LLMMock, err = strconv.ParseBool(env("LLM_MOCK", "true"))
	if err != nil {
		errs = append(errs, errors.New("LLM_MOCK: must be true or false"))
	}

	c.Retrieval, err = strconv.ParseBool(env("FACT_RETRIEVAL", "true"))
	if err != nil {
		errs = append(errs, errors.New("FACT_RETRIEVAL: must be true or false"))
	}

	switch c.Reasoning {
	case "low", "medium", "high", "minimal":
	default:
		errs = append(errs, errors.New("NINEROUTER_REASONING_EFFORT: must be minimal, low, medium or high"))
	}

	mb, err := strconv.ParseInt(env("MAX_UPLOAD_SIZE_MB", "10"), 10, 64)
	if err != nil || mb < 1 || mb > 25 {
		errs = append(errs, errors.New("MAX_UPLOAD_SIZE_MB: must be an integer between 1 and 25"))
	}
	c.MaxUploadMB = mb

	ms, err := strconv.Atoi(env("HARD_TIMEOUT_MS", "20000"))
	if err != nil || ms < 1000 || ms > 60000 {
		errs = append(errs, errors.New("HARD_TIMEOUT_MS: must be an integer between 1000 and 60000"))
	}
	c.HardTimeout = time.Duration(ms) * time.Millisecond

	ttl, err := strconv.Atoi(env("CACHE_TTL_SECONDS", "600"))
	if err != nil || ttl < 0 {
		errs = append(errs, errors.New("CACHE_TTL_SECONDS: must be a non-negative integer"))
	}
	c.CacheTTL = time.Duration(ttl) * time.Second

	if raw := os.Getenv("TRUSTED_PROXY_CIDRS"); strings.TrimSpace(raw) != "" {
		for _, p := range strings.Split(raw, ",") {
			if p = strings.TrimSpace(p); p != "" {
				c.TrustedProxy = append(c.TrustedProxy, p)
			}
		}
	}

	if !c.LLMMock {
		if c.RouterAPIKey == "" {
			errs = append(errs, errors.New("NINEROUTER_API_KEY: required when LLM_MOCK=false"))
		}
		if c.RouterModel == "" {
			errs = append(errs, errors.New("NINEROUTER_MODEL: required when LLM_MOCK=false"))
		}
	}

	// Production must be able to sign share tokens; without a key every
	// /verify would fail closed, so refuse to boot instead.
	if c.IsProduction() && len(c.SigningKey) < 32 {
		errs = append(errs, errors.New("REPORT_SIGNING_KEY: required (>= 32 bytes) in production"))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return c, nil
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
