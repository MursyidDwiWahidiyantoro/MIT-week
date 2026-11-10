package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpsStatusIsHonestInMockMode(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ops/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Status string `json:"status"`
		LLM    struct {
			Mode        string `json:"mode"`
			Placeholder bool   `json:"placeholder"`
		} `json:"llm"`
		Retrieval struct {
			Enabled     bool `json:"enabled"`
			Placeholder bool `json:"placeholder"`
		} `json:"retrieval"`
		SafeBrowsing struct {
			Placeholder bool `json:"placeholder"`
		} `json:"safe_browsing"`
		Pillars map[string]string `json:"pillars"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Errorf("status = %q", body.Status)
	}
	if body.LLM.Mode != "mock" || !body.LLM.Placeholder {
		t.Errorf("mock mode must be labelled a placeholder: %+v", body.LLM)
	}
	if body.Retrieval.Enabled {
		t.Error("retrieval must not claim to run in mock mode")
	}
	if body.Pillars["ai"] != "placeholder" || body.Pillars["fact"] != "placeholder" {
		t.Errorf("pillars = %v", body.Pillars)
	}
	if !body.SafeBrowsing.Placeholder {
		t.Error("safe browsing must be labelled a placeholder without a key")
	}
}

func TestOpsStatusNeverLeaksSecrets(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ops/status", nil))
	body := rec.Body.String()
	for _, banned := range []string{"Bearer", "NINEROUTER_API_KEY", "api_key", "20128"} {
		if contains(body, banned) {
			t.Errorf("ops response contains %q", banned)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
