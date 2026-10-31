package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amba/amba/backend/internal/models"
)

func postJSON(t *testing.T, path, body string, accept string) *httptest.ResponseRecorder {
	t.Helper()
	return postJSONTo(t, testRouter(t), path, body, accept)
}

// postJSONTo sends a request through a specific router. Token round-trip tests
// must share one router: without REPORT_SIGNING_KEY each new handler set gets a
// fresh random dev key, so a token signed by one cannot verify in another.
func postJSONTo(t *testing.T, r http.Handler, path, body string, accept string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestAuditTextReturnsDossier(t *testing.T) {
	rec := postJSON(t, "/api/v1/audit/text", `{"text":"Bantuan dibagikan hari ini saja"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var d models.Dossier
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatalf("not a Dossier: %v", err)
	}
	if d.SchemaVersion != 2 || d.InputType != models.InputText {
		t.Fatalf("unexpected dossier: %+v", d)
	}
	if d.IsDemo {
		t.Error("a live audit must never be marked as demo data")
	}
	if d.Verdict.ExplanationSimple == "" {
		t.Error("explanation_simple must always be present")
	}
	if d.ShareToken == "" {
		t.Error("expected a signed share token")
	}
}

func TestAuditTextRejectsEmptyAndOversize(t *testing.T) {
	for _, body := range []string{`{"text":""}`, `{"text":"   "}`} {
		if rec := postJSON(t, "/api/v1/audit/text", body, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("body %q: status = %d, want 400", body, rec.Code)
		}
	}
	long := strings.Repeat("a", 5001)
	if rec := postJSON(t, "/api/v1/audit/text", `{"text":"`+long+`"}`, ""); rec.Code != http.StatusBadRequest {
		t.Errorf("oversize text: status = %d, want 400", rec.Code)
	}
}

func TestAuditURLRejectsBadScheme(t *testing.T) {
	for _, url := range []string{"", "file:///etc/passwd", "javascript:alert(1)", "http://"} {
		body := `{"url":"` + url + `"}`
		rec := postJSON(t, "/api/v1/audit/url", body, "")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("url %q: status = %d, want 400", url, rec.Code)
		}
	}
}

func TestAuditURLAcceptsHTTPS(t *testing.T) {
	rec := postJSON(t, "/api/v1/audit/url", `{"url":"https://example.org/artikel"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var d models.Dossier
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if !d.Pillars.Link.Available {
		t.Error("link pillar should be available for a URL audit")
	}
}

func TestAuditSSEOrder(t *testing.T) {
	rec := postJSON(t, "/api/v1/audit/text", `{"text":"Bantuan hari ini saja"}`, "text/event-stream")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()

	// Every event type must appear, and done must be last.
	for _, want := range []string{"event: stage", "event: partial", "event: verdict", "event: done"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in stream:\n%s", want, body)
		}
	}
	doneIdx := strings.LastIndex(body, "event: done")
	verdictIdx := strings.LastIndex(body, "event: verdict")
	if doneIdx < verdictIdx {
		t.Error("done must come after verdict")
	}
	if strings.Index(body, "event: verdict") < strings.Index(body, "event: stage") {
		t.Error("verdict must come after stage")
	}
}

func TestAuditSSENeverEmitsHeartbeatAsProgress(t *testing.T) {
	rec := postJSON(t, "/api/v1/audit/text", `{"text":"halo"}`, "text/event-stream")
	if strings.Contains(rec.Body.String(), "event: ping") {
		t.Error("heartbeat must be a comment, never an event")
	}
}

func TestVerifyRoundTrip(t *testing.T) {
	router := testRouter(t)
	rec := postJSONTo(t, router, "/api/v1/audit/text", `{"text":"contoh untuk token"}`, "")
	var d models.Dossier
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}

	body := `{"token":"` + d.ShareToken + `"}`
	vr := postJSONTo(t, router, "/api/v1/verify", body, "")
	if vr.Code != http.StatusOK {
		t.Fatalf("verify status = %d", vr.Code)
	}
	var out struct {
		Valid   bool `json:"valid"`
		Payload struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(vr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !out.Valid || out.Payload.ID != d.ID {
		t.Fatalf("verify result = %+v, want valid with id %s", out, d.ID)
	}
}

func TestVerifyRejectsTamperedToken(t *testing.T) {
	rec := postJSON(t, "/api/v1/verify", `{"token":"v1.eyJpZCI6IngifQ.deadbeef"}`, "")
	var out struct {
		Valid bool `json:"valid"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out.Valid {
		t.Fatal("tampered token verified as valid")
	}
}

func TestReplyRequiresValidToken(t *testing.T) {
	rec := postJSON(t, "/api/v1/reply", `{"token":"nonsense","tone":"hormat"}`, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	var env models.ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Error.Message != "Laporan tidak sah." {
		t.Errorf("message = %q", env.Error.Message)
	}
}

func TestReplyReturnsTemplate(t *testing.T) {
	router := testRouter(t)
	rec := postJSONTo(t, router, "/api/v1/audit/text", `{"text":"contoh balasan"}`, "")
	var d models.Dossier
	_ = json.Unmarshal(rec.Body.Bytes(), &d)

	body := `{"token":"` + d.ShareToken + `","tone":"santai"}`
	rr := postJSONTo(t, router, "/api/v1/reply", body, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var out struct {
		Text   string `json:"text"`
		Source string `json:"source"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Text == "" || len([]rune(out.Text)) > 280 {
		t.Errorf("reply text invalid: %q", out.Text)
	}
	if out.Source != "template" {
		t.Errorf("source = %q, want template", out.Source)
	}
	for _, banned := range []string{"pasti", "100%", "terbukti"} {
		if strings.Contains(strings.ToLower(out.Text), banned) {
			t.Errorf("reply contains certainty word %q: %s", banned, out.Text)
		}
	}
}

func TestReplyRejectsUnknownTone(t *testing.T) {
	rec := postJSON(t, "/api/v1/reply", `{"token":"x","tone":"marah"}`, "")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAuditFileRejectsNonImage(t *testing.T) {
	body := strings.NewReader("not an image at all, just text")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/audit/file", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	rec := httptest.NewRecorder()
	testRouter(t).ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("non-image upload accepted")
	}
}
