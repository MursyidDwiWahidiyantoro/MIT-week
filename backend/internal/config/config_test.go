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

func TestPortFallsBackToPaaSPort(t *testing.T) {
	setEnv(t, map[string]string{"SERVER_PORT": "", "PORT": "9123"})
	cfg, err := Load()
	if err != nil || cfg.Port != 9123 {
		t.Fatalf("port = %v, err = %v", cfg, err)
	}
	setEnv(t, map[string]string{"SERVER_PORT": "8081", "PORT": "9123"})
	if cfg, _ := Load(); cfg.Port != 8081 {
		t.Fatalf("SERVER_PORT must win, got %d", cfg.Port)
	}
}

func TestGoogleKeyFallback(t *testing.T) {
	setEnv(t, map[string]string{"GOOGLE_API_KEY": "shared", "GOOGLE_SAFE_BROWSING_KEY": "", "GOOGLE_FACTCHECK_KEY": "own"})
	cfg, err := Load()
	if err != nil || cfg.SafeBrowsing != "shared" || cfg.FactCheckKey != "own" {
		t.Fatalf("cfg = %+v err = %v", cfg, err)
	}
}

func TestSightengineNeedsBothCredentials(t *testing.T) {
	setEnv(t, map[string]string{"SIGHTENGINE_API_USER": "123", "SIGHTENGINE_API_SECRET": ""})
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "SIGHTENGINE") {
		t.Fatalf("want SIGHTENGINE error, got %v", err)
	}
	setEnv(t, map[string]string{"SIGHTENGINE_API_USER": "123", "SIGHTENGINE_API_SECRET": "s"})
	if cfg, err := Load(); err != nil || cfg.SightengineUser != "123" || cfg.SightengineSecret != "s" {
		t.Fatalf("cfg = %+v err = %v", cfg, err)
	}
}

func TestClientIPHeaderNeedsTrustedProxies(t *testing.T) {
	setEnv(t, map[string]string{"CLIENT_IP_HEADER": "Fly-Client-IP", "TRUSTED_PROXY_CIDRS": ""})
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "CLIENT_IP_HEADER") {
		t.Fatalf("want CLIENT_IP_HEADER error, got %v", err)
	}
	setEnv(t, map[string]string{"CLIENT_IP_HEADER": "Fly-Client-IP", "TRUSTED_PROXY_CIDRS": "*"})
	if _, err := Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTelegramWebhookNeedsSecret(t *testing.T) {
	setEnv(t, map[string]string{"TELEGRAM_BOT_TOKEN": "123:abc", "TELEGRAM_MODE": "webhook", "TELEGRAM_WEBHOOK_SECRET": "short"})
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "TELEGRAM_WEBHOOK_SECRET") {
		t.Fatalf("want secret error, got %v", err)
	}
	setEnv(t, map[string]string{"TELEGRAM_MODE": "sideways"})
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "TELEGRAM_MODE") {
		t.Fatalf("want mode error, got %v", err)
	}
}

func TestProductionRequiresHTTPSAppURL(t *testing.T) {
	setEnv(t, map[string]string{
		"ENVIRONMENT": "production", "REPORT_SIGNING_KEY": strings.Repeat("k", 32),
		"PUBLIC_APP_URL": "http://amba.example",
	})
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "PUBLIC_APP_URL") {
		t.Fatalf("want PUBLIC_APP_URL error, got %v", err)
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
		"PUBLIC_APP_URL":     "https://amba.example",
		"LLM_MOCK":           "false",
		"NINEROUTER_API_KEY": "key",
		"NINEROUTER_MODEL":   "model",
	})
	if _, err := Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// The mock model's findings are invented; production must never serve them.
func TestProductionRefusesMockLLM(t *testing.T) {
	setEnv(t, map[string]string{
		"ENVIRONMENT":        "production",
		"REPORT_SIGNING_KEY": strings.Repeat("k", 32),
		"PUBLIC_APP_URL":     "https://amba.example",
		"LLM_MOCK":           "",
	})
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "LLM_MOCK") {
		t.Fatalf("want LLM_MOCK error, got %v", err)
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

func TestLLMExtraBody(t *testing.T) {
	setEnv(t, map[string]string{"LLM_EXTRA_BODY": `{"reasoning":{"enabled":false},"models":["a:free","openrouter/free"]}`})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cfg.LLMExtraBody["reasoning"]; !ok || len(cfg.LLMExtraBody) != 2 {
		t.Fatalf("parsed %v", cfg.LLMExtraBody)
	}
	for _, bad := range []string{`not json`, `["a"]`, `null`, `"text"`} {
		t.Run(bad, func(t *testing.T) {
			setEnv(t, map[string]string{"LLM_EXTRA_BODY": bad})
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "LLM_EXTRA_BODY") {
				t.Fatalf("want LLM_EXTRA_BODY error, got %v", err)
			}
		})
	}
}
