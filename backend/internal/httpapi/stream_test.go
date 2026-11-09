package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/scam"
)

type sseEvent struct {
	name string
	data string
}

func parseSSE(body string) []sseEvent {
	var out []sseEvent
	for _, block := range strings.Split(body, "\n\n") {
		var e sseEvent
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "event: "):
				e.name = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				e.data = strings.TrimPrefix(line, "data: ")
			}
		}
		if e.name != "" {
			out = append(out, e)
		}
	}
	return out
}

// Every pillar starts, ends exactly once, and its partial follows its own
// "done" stage; the verdict comes after every pillar settled.
func TestSSEIsProgressivePerPillar(t *testing.T) {
	rec := postJSON(t, "/api/v1/audit/text", `{"text":"Segera klaim hadiah https://bri-verifikasi.xyz/login"}`, "text/event-stream")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	events := parseSSE(rec.Body.String())
	ended := map[string]string{}
	lastStageDone := map[string]int{}
	verdictAt := -1
	for i, e := range events {
		switch e.name {
		case "stage":
			var s struct{ Pillar, Status string }
			_ = json.Unmarshal([]byte(e.data), &s)
			if s.Status != "running" {
				if _, dup := ended[s.Pillar]; dup {
					t.Fatalf("%s ended twice", s.Pillar)
				}
				ended[s.Pillar] = s.Status
				lastStageDone[s.Pillar] = i
			}
		case "partial":
			var p struct{ Pillar string }
			_ = json.Unmarshal([]byte(e.data), &p)
			if at, ok := lastStageDone[p.Pillar]; !ok || at > i || ended[p.Pillar] != "done" {
				t.Fatalf("partial for %s without a preceding done stage", p.Pillar)
			}
		case "verdict":
			verdictAt = i
		}
	}
	for _, p := range []string{"ela", "ai", "fact", "link"} {
		if _, ok := ended[p]; !ok {
			t.Fatalf("pillar %s never settled: %v", p, ended)
		}
		if lastStageDone[p] > verdictAt {
			t.Fatalf("pillar %s settled after the verdict", p)
		}
	}
	if ended["link"] != "done" || ended["ela"] != "skipped" {
		t.Fatalf("stages = %v", ended)
	}
	if events[len(events)-1].name != "done" {
		t.Fatalf("last event = %s", events[len(events)-1].name)
	}
}

func TestURLDossierCarriesSourceMetadata(t *testing.T) {
	// The sandboxed fetch cannot reach the host, which is exactly the
	// "could not read" case the UI has to explain.
	rec := postJSON(t, "/api/v1/audit/url", `{"url":"https://amba-test.invalid/post"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var d models.Dossier
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d.SourceMetadata == nil || d.SourceMetadata.Status != models.MetadataUnavailable {
		t.Fatalf("source_metadata = %+v", d.SourceMetadata)
	}
	if !strings.Contains(d.Verdict.Explanation, "tidak dapat dibaca") {
		t.Fatalf("explanation should mention the unreadable page: %q", d.Verdict.Explanation)
	}

	text := postJSON(t, "/api/v1/audit/text", `{"text":"halo"}`, "")
	if !strings.Contains(text.Body.String(), `"source_metadata":null`) {
		t.Fatal("text audits must carry source_metadata: null")
	}
}

func TestExplanationUsesPlainNames(t *testing.T) {
	rec := postJSON(t, "/api/v1/audit/text", `{"text":"halo semuanya"}`, "")
	var d models.Dossier
	_ = json.Unmarshal(rec.Body.Bytes(), &d)
	if strings.Contains(d.Verdict.Explanation, "ela") || !strings.Contains(d.Verdict.Explanation, "analisis gambar") {
		t.Fatalf("explanation should name pillars in plain words: %q", d.Verdict.Explanation)
	}
}

func TestInvalidTextSaysWhatIsWrong(t *testing.T) {
	rec := postJSON(t, "/api/v1/audit/text", `{"text":""}`, "")
	if !strings.Contains(rec.Body.String(), "Teks kosong") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

// Every scam pattern the rules can return has an explanation sentence and is
// in the published contract.
func TestScamPatternsAreInTheContract(t *testing.T) {
	raw, err := os.ReadFile(filepath.Clean("../../../api/dossier.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Defs struct {
			Verdict struct {
				Properties struct {
					ScamPattern struct {
						Enum []any `json:"enum"`
					} `json:"scam_pattern"`
				} `json:"properties"`
			} `json:"Verdict"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	inSchema := map[string]bool{}
	for _, v := range schema.Defs.Verdict.Properties.ScamPattern.Enum {
		if s, ok := v.(string); ok {
			inSchema[s] = true
		}
	}
	if len(inSchema) != len(scam.Patterns) {
		t.Errorf("schema lists %d patterns, the rules %d", len(inSchema), len(scam.Patterns))
	}
	for _, p := range scam.Patterns {
		if !inSchema[p] {
			t.Errorf("%s is missing from dossier.schema.json", p)
		}
		if scamText[p] == "" {
			t.Errorf("%s has no explanation sentence", p)
		}
	}
}

// A scam message gets the pattern, the floor and the plain sentence.
func TestScamMessageExplainsThePattern(t *testing.T) {
	rec := postJSON(t, "/api/v1/audit/text", `{"text":"Akun BRImo Anda akan diblokir dalam 24 jam. Segera verifikasi data di brimo-verifikasi.my.id"}`, "")
	var d models.Dossier
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatalf("%v: %s", err, rec.Body.String())
	}
	if d.Verdict.ScamPattern == nil || *d.Verdict.ScamPattern != "ACCOUNT_THREAT" {
		t.Fatalf("scam pattern = %v", d.Verdict.ScamPattern)
	}
	if d.Verdict.Status != models.StatusTampered {
		t.Fatalf("status = %s", d.Verdict.Status)
	}
	if !strings.Contains(d.Verdict.Explanation, scamText["ACCOUNT_THREAT"]) {
		t.Fatalf("explanation = %q", d.Verdict.Explanation)
	}
}
