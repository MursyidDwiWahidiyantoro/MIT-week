package reply

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// configs/reply_templates.json is the human-editable copy that the frontend
// bundles for offline demo. The Go matrix is what /reply actually serves, so
// the two drifting apart would make the offline and online drafts disagree.
func TestTemplatesMatchConfigFile(t *testing.T) {
	path := filepath.Join("..", "..", "..", "configs", "reply_templates.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("config file not reachable from the test working directory: %v", err)
	}
	var file struct {
		Templates map[string]map[string]string `json:"templates"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("configs/reply_templates.json is invalid: %v", err)
	}
	if len(file.Templates) != len(Templates) {
		t.Fatalf("config has %d statuses, Go has %d", len(file.Templates), len(Templates))
	}
	for status, tones := range Templates {
		for tone, want := range tones {
			got, ok := file.Templates[status][tone]
			if !ok {
				t.Errorf("config is missing %s/%s", status, tone)
				continue
			}
			if got != want {
				t.Errorf("%s/%s drifted:\n config: %q\n go:     %q", status, tone, got, want)
			}
		}
	}
}
