package config

import (
	"strings"
	"testing"
)

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func TestLoadDefaults(t *testing.T) {
	setEnv(t, map[string]string{
		"SERVER_PORT": "", "LLM_MOCK": "", "MAX_UPLOAD_SIZE_MB": "",
		"HARD_TIMEOUT_MS": "", "CACHE_TTL_SECONDS": "", "ENVIRONMENT": "",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 || !cfg.LLMMock || cfg.MaxUploadMB != 10 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.HardTimeout.Milliseconds() != 12000 {
		t.Fatalf("hard timeout = %v", cfg.HardTimeout)
	}
}

func TestLoadRejectsBadPort(t *testing.T) {
	for _, port := range []string{"0", "70000", "abc", "-1"} {
		t.Run(port, func(t *testing.T) {
			setEnv(t, map[string]string{"SERVER_PORT": port})
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), "SERVER_PORT") {
				t.Fatalf("want SERVER_PORT error, got %v", err)
			}
		})
	}
}

func TestProductionRequiresSigningKey(t *testing.T) {
	setEnv(t, map[string]string{
		"ENVIRONMENT": "production", "REPORT_SIGNING_KEY": "short",
	})
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "REPORT_SIGNING_KEY") {
		t.Fatalf("want signing key error, got %v", err)
	}
}

func TestProductionAcceptsLongKey(t *testing.T) {
	setEnv(t, map[string]string{
		"ENVIRONMENT":        "production",
		"REPORT_SIGNING_KEY": strings.Repeat("k", 32),
	})
	if _, err := Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockFalseRequiresRouterCreds(t *testing.T) {
	setEnv(t, map[string]string{
		"LLM_MOCK": "false", "NINEROUTER_API_KEY": "", "NINEROUTER_MODEL": "",
	})
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when LLM_MOCK=false without credentials")
	}
	if !strings.Contains(err.Error(), "NINEROUTER_API_KEY") ||
		!strings.Contains(err.Error(), "NINEROUTER_MODEL") {
		t.Fatalf("both missing vars should be named, got %v", err)
	}
}

func TestTrustedProxyParsing(t *testing.T) {
	setEnv(t, map[string]string{"TRUSTED_PROXY_CIDRS": "10.0.0.0/8, ,192.168.0.0/16"})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.TrustedProxy) != 2 {
		t.Fatalf("want 2 proxies, got %v", cfg.TrustedProxy)
	}
}
