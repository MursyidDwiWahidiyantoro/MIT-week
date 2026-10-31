// Package linkcheck scores a URL using deterministic heuristics only.
//
// Nothing here performs network I/O except UnrollRedirects, and every request
// it makes goes through the SSRF-safe client.
package linkcheck

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// Reason codes are stable identifiers; the Indonesian text is built by the
// caller from a static table so copy changes never touch this package.
const (
	ReasonLookalikeBrand  = "LOOKALIKE_BRAND"
	ReasonSuspiciousTLD   = "SUSPICIOUS_TLD"
	ReasonIPHost          = "IP_HOST"
	ReasonPunycode        = "PUNYCODE"
	ReasonManySubdomains  = "MANY_SUBDOMAINS"
	ReasonCredentials     = "CREDENTIALS_IN_URL"
	ReasonURLShortener    = "URL_SHORTENER"
	ReasonExcessiveLength = "EXCESSIVE_LENGTH"
	ReasonAtSymbol        = "AT_SYMBOL"
	ReasonNoHTTPS         = "NO_HTTPS"
	ReasonPortNonStandard = "NONSTANDARD_PORT"
	ReasonSuspiciousPath  = "SUSPICIOUS_PATH"
)

// brands are the names impersonated most often in Indonesian phishing.
var brands = []string{
	"bri", "bca", "bni", "mandiri", "bsi", "cimb", "permata", "danamon",
	"gopay", "ovo", "dana", "shopeepay", "linkaja", "qris",
	"tokopedia", "shopee", "lazada", "bukalapak", "blibli",
	"whatsapp", "telegram", "instagram", "facebook", "tiktok",
	"kemendikbud", "kominfo", "komdigi", "pajak", "djp", "bpjs", "jne", "jnt",
}

// suspiciousTLDs are cheap or frequently abused TLDs.
var suspiciousTLDs = map[string]bool{
	"zip": true, "mov": true, "top": true, "xyz": true, "click": true,
	"link": true, "shop": true, "icu": true, "buzz": true, "rest": true,
	"cam": true, "monster": true, "sbs": true, "cfd": true,
}

// shorteners hide the destination behind one more hop.
var shorteners = map[string]bool{
	"bit.ly": true, "tinyurl.com": true, "t.co": true, "goo.gl": true,
	"is.gd": true, "ow.ly": true, "buff.ly": true, "cutt.ly": true,
	"s.id": true, "shorturl.at": true, "rebrand.ly": true,
}

// suspiciousPathWords are common in credential-harvesting pages.
var suspiciousPathWords = []string{
	"login", "verifikasi", "verify", "account", "akun", "password",
	"konfirmasi", "unlock", "secure", "update-data", "hadiah", "klaim",
}

// Finding is one heuristic hit with its contribution to the score.
type Finding struct {
	Code   string
	Weight int
	Detail string
}

// Result is the outcome of the static analysis.
type Result struct {
	Score    int
	Findings []Finding
	Host     string
	FinalURL string
}

// Analyze scores a URL from its text alone. It never performs I/O.
func Analyze(raw string) Result {
	res := Result{FinalURL: raw}

	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		// A URL that cannot be parsed cannot be assessed; callers should have
		// rejected it earlier, so score it as maximally suspicious.
		res.Score = 100
		res.Findings = append(res.Findings, Finding{Code: ReasonIPHost, Weight: 100, Detail: "unparseable"})
		return res
	}

	host := strings.ToLower(u.Hostname())
	res.Host = host

	if u.User != nil {
		res.Findings = append(res.Findings, Finding{Code: ReasonCredentials, Weight: 40})
	}
	if strings.Contains(u.String(), "@") && u.User == nil {
		res.Findings = append(res.Findings, Finding{Code: ReasonAtSymbol, Weight: 30})
	}
	if u.Scheme != "https" {
		res.Findings = append(res.Findings, Finding{Code: ReasonNoHTTPS, Weight: 15})
	}
	if port := u.Port(); port != "" && port != "80" && port != "443" {
		res.Findings = append(res.Findings, Finding{Code: ReasonPortNonStandard, Weight: 20})
	}
	if len(u.String()) > 200 {
		res.Findings = append(res.Findings, Finding{Code: ReasonExcessiveLength, Weight: 10})
	}
	if isIPHost(host) {
		res.Findings = append(res.Findings, Finding{Code: ReasonIPHost, Weight: 35})
	}
	if strings.HasPrefix(host, "xn--") || strings.Contains(host, ".xn--") {
		res.Findings = append(res.Findings, Finding{Code: ReasonPunycode, Weight: 30})
	}
	if strings.Count(host, ".") >= 4 {
		res.Findings = append(res.Findings, Finding{Code: ReasonManySubdomains, Weight: 20})
	}

	tld := lastLabel(host)
	if suspiciousTLDs[tld] {
		res.Findings = append(res.Findings, Finding{Code: ReasonSuspiciousTLD, Weight: 25, Detail: tld})
	}
	if shorteners[registrable(host)] {
		res.Findings = append(res.Findings, Finding{Code: ReasonURLShortener, Weight: 20, Detail: registrable(host)})
	}
	if brand, ok := lookalikeBrand(host); ok {
		res.Findings = append(res.Findings, Finding{Code: ReasonLookalikeBrand, Weight: 45, Detail: brand})
	}

	path := strings.ToLower(u.Path)
	for _, word := range suspiciousPathWords {
		if strings.Contains(path, word) {
			res.Findings = append(res.Findings, Finding{Code: ReasonSuspiciousPath, Weight: 15, Detail: word})
			break
		}
	}

	res.Score = capScore(sumWeights(res.Findings))
	return res
}

func sumWeights(findings []Finding) int {
	total := 0
	for _, f := range findings {
		total += f.Weight
	}
	return total
}

func capScore(n int) int {
	if n > 100 {
		return 100
	}
	if n < 0 {
		return 0
	}
	return n
}

// lookalikeBrand reports a brand name that appears in the host but is not the
// brand's own registrable domain, which is the classic typosquat pattern.
func lookalikeBrand(host string) (string, bool) {
	reg := registrable(host)
	labels := strings.Split(host, ".")
	for _, brand := range brands {
		if !strings.Contains(host, brand) {
			continue
		}
		// The brand's own domain (brand.co.id, brand.com) is not a lookalike.
		if reg == brand+".com" || reg == brand+".co.id" || reg == brand+".id" ||
			reg == brand+".net" || reg == brand+".org" {
			return "", false
		}
		// A hyphen or an extra label around the brand is the usual giveaway.
		for _, label := range labels {
			if label == brand {
				return "", false
			}
			if strings.Contains(label, brand) && label != brand {
				return brand, true
			}
		}
		return brand, true
	}
	return "", false
}

func lastLabel(host string) string {
	parts := strings.Split(host, ".")
	return parts[len(parts)-1]
}

// registrable approximates the eTLD+1 for the TLDs this project cares about.
// It is deliberately simple: a full public-suffix list is out of scope, and
// every caller only needs a stable comparison key.
func registrable(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return host
	}
	// Handle common Indonesian two-part suffixes.
	if len(parts) >= 3 {
		second := parts[len(parts)-2]
		if second == "co" || second == "ac" || second == "or" || second == "go" ||
			second == "net" || second == "sch" || second == "web" {
			return strings.Join(parts[len(parts)-3:], ".")
		}
	}
	return strings.Join(parts[len(parts)-2:], ".")
}

func isIPHost(host string) bool {
	if strings.Contains(host, ":") {
		return true // bare IPv6
	}
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if p == "" || len(p) > 3 {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

// Unroll follows redirects through the safe client and returns the chain.
// It stops at maxHops and never returns a body.
func Unroll(ctx context.Context, client *http.Client, raw string, maxHops int) ([]string, string, error) {
	chain := []string{raw}
	current := raw

	for i := 0; i < maxHops; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, current, nil)
		if err != nil {
			return chain, current, err
		}
		req.Header.Set("User-Agent", "AMBA/0.1 (+literacy check)")

		// Do not follow redirects here: the chain is the output.
		c := *client
		c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

		resp, err := c.Do(req)
		if err != nil {
			return chain, current, err
		}
		_ = resp.Body.Close()

		if resp.StatusCode < 300 || resp.StatusCode >= 400 {
			return chain, current, nil
		}
		loc := resp.Header.Get("Location")
		if loc == "" {
			return chain, current, nil
		}
		next, err := resp.Request.URL.Parse(loc)
		if err != nil {
			return chain, current, err
		}
		current = next.String()
		chain = append(chain, current)
	}
	return chain, current, nil
}
