package grounding

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestVerifiedRequiresAllowlistedHost(t *testing.T) {
	g := New(DefaultAllowlist)
	tools := map[string]bool{
		"https://www.antaranews.com/berita/1": true,
		"https://evil.example/berita/1":       true,
	}

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"allowlisted exact", "https://www.antaranews.com/berita/1", true},
		{"allowlisted subdomain", "https://www.antaranews.com/berita/1", true},
		{"not allowlisted", "https://evil.example/berita/1", false},
		{"allowlisted but not in tool results", "https://kompas.com/berita/9", false},
		{"http not https", "http://antaranews.com/berita/1", false},
		{"lookalike suffix", "https://antaranews.com.evil.example/x", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The tool-result set only contains the two URLs above.
			got := g.Verified(tt.url, tools)
			if got != tt.want {
				t.Errorf("Verified(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestSubdomainOfAllowedHostIsAccepted(t *testing.T) {
	g := New([]string{"kompas.com"})
	if !g.Verified("https://news.kompas.com/a", nil) {
		t.Error("subdomain should be accepted")
	}
	if g.Verified("https://notkompas.com/a", nil) {
		t.Error("a suffix without a dot boundary must not be accepted")
	}
}

func TestEmptyAllowlistRejectsEverything(t *testing.T) {
	g := New(nil)
	if g.Verified("https://antaranews.com/x", nil) {
		t.Error("empty allowlist accepted a host")
	}
}

// The allowlist is duplicated on purpose: the YAML is the reviewed, human
// editable contract; DefaultAllowlist is what ships in the binary. This test
// makes drift between them a build failure instead of a silent policy change.
func TestAllowlistMatchesConfigFile(t *testing.T) {
	raw, err := os.ReadFile("../../../configs/trusted_sources.yaml")
	if err != nil {
		t.Fatalf("read trusted_sources.yaml: %v", err)
	}
	var fromFile []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			fromFile = append(fromFile, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
		}
	}
	if !slices.Equal(fromFile, DefaultAllowlist) {
		t.Fatalf("allowlist drift:\n file: %v\n code: %v", fromFile, DefaultAllowlist)
	}
}
