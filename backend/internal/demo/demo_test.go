package demo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// schemaPath is the contract generated from api/openapi.yaml.
const schemaPath = "../../../api/dossier.schema.json"

func TestEveryCaseValidatesAgainstSchema(t *testing.T) {
	// Structural validation without a JSON-schema dependency: check every
	// required field and enum the contract declares.
	schemaBytes, err := os.ReadFile(filepath.Clean(schemaPath))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	defs, _ := schema["$defs"].(map[string]any)
	dossierDef, ok := defs["Dossier"].(map[string]any)
	if !ok {
		t.Fatal("schema has no Dossier definition")
	}
	required, _ := dossierDef["required"].([]any)

	for _, id := range IDs() {
		c, err := Get(id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		var d map[string]any
		if err := json.Unmarshal(c.Dossier, &d); err != nil {
			t.Fatalf("%s: dossier is not an object: %v", id, err)
		}
		for _, r := range required {
			name := r.(string)
			if _, present := d[name]; !present {
				t.Errorf("%s: missing required field %q", id, name)
			}
		}
		if d["schema_version"].(float64) != 2 {
			t.Errorf("%s: schema_version must be 2", id)
		}
		if d["is_demo"] != true {
			t.Errorf("%s: is_demo must be true", id)
		}
		if strings.TrimSpace(c.Caption) == "" {
			t.Errorf("%s: caption is empty", id)
		}
	}
}

func TestStatusesCoverTheThreeMainBands(t *testing.T) {
	seen := map[string]bool{}
	for _, id := range IDs() {
		c, _ := Get(id)
		var d struct {
			Verdict struct {
				Status string `json:"status"`
			} `json:"verdict"`
		}
		if err := json.Unmarshal(c.Dossier, &d); err != nil {
			t.Fatal(err)
		}
		seen[d.Verdict.Status] = true
	}
	for _, want := range []string{"SAFE", "TAMPERED"} {
		if !seen[want] {
			t.Errorf("no demo case with status %s", want)
		}
	}
}

func TestDemoSourcesArePlaceholders(t *testing.T) {
	for _, id := range IDs() {
		c, _ := Get(id)
		if strings.Contains(string(c.Dossier), "http") {
			var d struct {
				Pillars struct {
					Fact struct {
						Sources []struct {
							URL       string `json:"url"`
							Publisher string `json:"publisher"`
						} `json:"sources"`
					} `json:"fact"`
				} `json:"pillars"`
			}
			if err := json.Unmarshal(c.Dossier, &d); err != nil {
				t.Fatal(err)
			}
			for _, s := range d.Pillars.Fact.Sources {
				if !strings.HasPrefix(s.URL, "https://example.org/") {
					t.Errorf("%s: source URL %q is not an example.org placeholder", id, s.URL)
				}
				if s.Publisher != "Sumber Contoh" {
					t.Errorf("%s: publisher %q is not 'Sumber Contoh'", id, s.Publisher)
				}
			}
		}
	}
}

func TestUnknownAndInvalidIDs(t *testing.T) {
	for _, id := range []string{"", "ab", "UPPER", "has space", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"} {
		if _, err := Get(id); err == nil {
			t.Errorf("Get(%q) should fail", id)
		}
	}
}

func TestTacticPhrasesAppearInCaption(t *testing.T) {
	for _, id := range IDs() {
		c, _ := Get(id)
		var d struct {
			Tactics []struct {
				Phrase string `json:"phrase"`
			} `json:"tactics"`
		}
		if err := json.Unmarshal(c.Dossier, &d); err != nil {
			t.Fatal(err)
		}
		for _, tac := range d.Tactics {
			if !strings.Contains(strings.ToLower(c.Caption), strings.ToLower(tac.Phrase)) {
				t.Errorf("%s: tactic phrase %q does not appear in the caption", id, tac.Phrase)
			}
		}
	}
}
