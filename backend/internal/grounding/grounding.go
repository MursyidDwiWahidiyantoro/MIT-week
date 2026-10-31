// Package grounding decides whether a source returned by the model may be
// shown to the user. A source is only trusted when it came from the model's
// own tool results AND its host is on the reviewed allowlist.
package grounding

import (
	"net/url"
	"strings"
)

// DefaultAllowlist is the starter set from TECH_SPEC §6, pending team review.
var DefaultAllowlist = []string{
	"turnbackhoax.id",
	"cekfakta.com",
	"komdigi.go.id",
	"kominfo.go.id",
	"antaranews.com",
	"kompas.com",
	"detik.com",
	"tempo.co",
	"cnnindonesia.com",
	"liputan6.com",
	"tirto.id",
}

// Grounder verifies candidate sources against an allowlist.
type Grounder struct {
	allowed map[string]bool
}

// New builds a Grounder from an allowlist. A subdomain of an allowed host is
// accepted; anything else is not.
func New(allowlist []string) *Grounder {
	m := make(map[string]bool, len(allowlist))
	for _, host := range allowlist {
		m[strings.ToLower(strings.TrimSpace(host))] = true
	}
	return &Grounder{allowed: m}
}

// Verified reports whether rawURL is https and its registrable host is
// allowlisted. toolURLs is the set of URLs the model's own tools returned:
// a source the model invented cannot pass even if its host is allowlisted.
func (g *Grounder) Verified(rawURL string, toolURLs map[string]bool) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return false
	}
	if toolURLs != nil && !toolURLs[rawURL] {
		return false
	}
	return g.hostAllowed(strings.ToLower(u.Hostname()))
}

func (g *Grounder) hostAllowed(host string) bool {
	if host == "" {
		return false
	}
	if g.allowed[host] {
		return true
	}
	for allowed := range g.allowed {
		if strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}
