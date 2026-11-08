// Package metadata reads the public description of a link: oEmbed for the
// platforms that offer it without a login (YouTube, TikTok, X) and OpenGraph
// tags for everything else (PRD F1). Every request goes through the
// SSRF-safe client the caller provides; nothing is fetched behind a login and
// nothing is stored.
package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Status says whether the description could be read.
type Status string

const (
	// StatusOK means at least a title or description was read.
	StatusOK Status = "OK"
	// StatusBlocked means the platform refused automated reading (login wall,
	// 401/403/429/451). The UI asks for a screenshot instead (PRD F1 AC).
	StatusBlocked Status = "BLOCKED"
	// StatusUnavailable means the page could not be reached or had nothing to
	// read (timeout, dead link, not a web page).
	StatusUnavailable Status = "UNAVAILABLE"
)

// Metadata is the public description of a link, already trimmed to the
// TECH_SPEC §6 budget (about 2 KB of text).
type Metadata struct {
	Status      Status
	Platform    string
	Via         string
	Title       string
	Author      string
	Description string
	SiteName    string
}

// Caption is the user-visible text of the post: title and description. It is
// the text tactic phrases are verified against and highlighted in, so it must
// match what the frontend renders.
func (m Metadata) Caption() string {
	parts := make([]string, 0, 2)
	if m.Title != "" {
		parts = append(parts, m.Title)
	}
	if m.Description != "" && m.Description != m.Title {
		parts = append(parts, m.Description)
	}
	return strings.Join(parts, "\n")
}

// ForModel renders the metadata as labelled lines for the LLM. The caller
// wraps it in <untrusted_content>; nothing here is an instruction.
func (m Metadata) ForModel(rawURL string) string {
	var b strings.Builder
	b.WriteString("Tautan: " + rawURL + "\n")
	if m.Platform != "" {
		b.WriteString("Platform: " + m.Platform + "\n")
	}
	switch m.Status {
	case StatusBlocked:
		b.WriteString("Isi tautan: platform tidak mengizinkan pembacaan otomatis.\n")
	case StatusUnavailable:
		b.WriteString("Isi tautan: tidak dapat dibaca.\n")
	}
	if m.SiteName != "" {
		b.WriteString("Situs: " + m.SiteName + "\n")
	}
	if m.Author != "" {
		b.WriteString("Pengunggah: " + m.Author + "\n")
	}
	if m.Title != "" {
		b.WriteString("Judul: " + m.Title + "\n")
	}
	if m.Description != "" && m.Description != m.Title {
		b.WriteString("Deskripsi: " + m.Description + "\n")
	}
	return b.String()
}

// Resolver fetches metadata. Client must be the SSRF-safe client.
type Resolver struct {
	Client *http.Client
	// OEmbed maps a platform to its oEmbed endpoint. Overridable in tests.
	OEmbed map[string]string
}

// DefaultOEmbed lists the endpoints that answer without an API token.
// Instagram and Facebook oEmbed need a Meta app token, so those platforms go
// through OpenGraph and usually end up BLOCKED, which is the honest outcome.
var DefaultOEmbed = map[string]string{
	"youtube": "https://www.youtube.com/oembed?format=json&url=",
	"tiktok":  "https://www.tiktok.com/oembed?url=",
	"x":       "https://publish.twitter.com/oembed?omit_script=1&dnt=true&url=",
}

// New builds a resolver around the given (SSRF-safe) client.
func New(client *http.Client) *Resolver {
	return &Resolver{Client: client, OEmbed: DefaultOEmbed}
}

const (
	maxBody        = 2 << 20 // TECH_SPEC §7.1: every body read is capped at 2 MB
	maxHeadScan    = 512 << 10
	maxTitle       = 300
	maxDescription = 1000
	maxAuthor      = 120
	maxSite        = 80
	userAgent      = "AMBA/1.0 (+https://github.com/fujimaruC/MIT-week; link preview for a literacy check)"
)

// Resolve reads the link's public description. It never returns an error:
// failure is a Status the UI can explain.
func (r *Resolver) Resolve(ctx context.Context, rawURL string) Metadata {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return Metadata{Status: StatusUnavailable, Platform: "web"}
	}
	platform := Platform(u.Hostname())
	out := Metadata{Platform: platform}

	if endpoint, ok := r.OEmbed[platform]; ok {
		m, err := r.oembed(ctx, endpoint+url.QueryEscape(rawURL))
		if err == nil && (m.Title != "" || m.Description != "") {
			m.Platform, m.Via, m.Status = platform, "oembed", StatusOK
			return m
		}
		if errors.Is(err, errBlocked) {
			out.Status, out.Via = StatusBlocked, "oembed"
			return out
		}
		// Fall through to OpenGraph: oEmbed can be missing for some URL forms
		// (e.g. short links) while the page itself still has tags.
	}

	m, err := r.opengraph(ctx, rawURL, platform)
	switch {
	case errors.Is(err, errBlocked):
		out.Status, out.Via = StatusBlocked, "opengraph"
		return out
	case err != nil:
		out.Status = StatusUnavailable
		return out
	}
	m.Platform, m.Via = platform, "opengraph"
	if m.Title == "" && m.Description == "" {
		m.Status = StatusUnavailable
		if isSocial(platform) {
			// A social page with no readable caption is a login wall.
			m.Status = StatusBlocked
		}
		return m
	}
	m.Status = StatusOK
	return m
}

// Platform names the social platform a host belongs to, or "web".
func Platform(host string) string {
	host = strings.ToLower(strings.TrimPrefix(host, "www."))
	switch {
	case host == "youtube.com" || host == "youtu.be" || strings.HasSuffix(host, ".youtube.com"):
		return "youtube"
	case host == "tiktok.com" || strings.HasSuffix(host, ".tiktok.com"):
		return "tiktok"
	case host == "x.com" || host == "twitter.com" || strings.HasSuffix(host, ".twitter.com") || strings.HasSuffix(host, ".x.com"):
		return "x"
	case host == "instagram.com" || strings.HasSuffix(host, ".instagram.com"):
		return "instagram"
	case host == "facebook.com" || host == "fb.com" || host == "fb.watch" || strings.HasSuffix(host, ".facebook.com"):
		return "facebook"
	case host == "threads.net" || host == "threads.com" || strings.HasSuffix(host, ".threads.net"):
		return "threads"
	default:
		return "web"
	}
}

func isSocial(platform string) bool { return platform != "web" }

var errBlocked = errors.New("platform refused automated reading")

func (r *Resolver) get(ctx context.Context, target string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/json;q=0.9,*/*;q=0.5")
	req.Header.Set("Accept-Language", "id,en;q=0.8")
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests, http.StatusUnavailableForLegalReasons:
		_ = resp.Body.Close()
		return nil, errBlocked
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_ = resp.Body.Close()
		return nil, errors.New("unexpected status")
	}
	if loginWall(resp.Request.URL) {
		_ = resp.Body.Close()
		return nil, errBlocked
	}
	return resp, nil
}

func (r *Resolver) oembed(ctx context.Context, endpoint string) (Metadata, error) {
	resp, err := r.get(ctx, endpoint)
	if err != nil {
		return Metadata{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	var body struct {
		Title        string `json:"title"`
		AuthorName   string `json:"author_name"`
		ProviderName string `json:"provider_name"`
		HTML         string `json:"html"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(&body); err != nil {
		return Metadata{}, err
	}
	m := Metadata{
		Title:    clean(body.Title, maxTitle),
		Author:   clean(body.AuthorName, maxAuthor),
		SiteName: clean(body.ProviderName, maxSite),
	}
	// X returns the post as an HTML blockquote; its first paragraph is the
	// text. Tags are stripped, never rendered.
	if m.Title == "" && body.HTML != "" {
		if p := paragraph.FindStringSubmatch(body.HTML); p != nil {
			m.Description = clean(stripTags(p[1]), maxDescription)
		}
	}
	return m, nil
}

func (r *Resolver) opengraph(ctx context.Context, rawURL, platform string) (Metadata, error) {
	resp, err := r.get(ctx, rawURL)
	if err != nil {
		return Metadata{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if ct != "" && !strings.Contains(ct, "html") {
		return Metadata{}, errors.New("not a web page")
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxHeadScan))
	if err != nil && len(raw) == 0 {
		return Metadata{}, err
	}
	doc := strings.ToValidUTF8(string(raw), "")
	if i := strings.Index(strings.ToLower(doc), "</head>"); i > 0 {
		doc = doc[:i]
	}

	meta := map[string]string{}
	for _, tag := range metaTag.FindAllString(doc, 200) {
		key := strings.ToLower(attr(tag, "property"))
		if key == "" {
			key = strings.ToLower(attr(tag, "name"))
		}
		if key == "" {
			continue
		}
		if _, seen := meta[key]; !seen {
			meta[key] = attr(tag, "content")
		}
	}
	title := first(meta["og:title"], meta["twitter:title"])
	if title == "" {
		if t := titleTag.FindStringSubmatch(doc); t != nil {
			title = t[1]
		}
	}
	m := Metadata{
		Title:       clean(title, maxTitle),
		Description: clean(first(meta["og:description"], meta["twitter:description"], meta["description"]), maxDescription),
		SiteName:    clean(meta["og:site_name"], maxSite),
		Author:      clean(first(meta["author"], meta["article:author"]), maxAuthor),
	}
	if isSocial(platform) && genericSocialTitle(m.Title) && m.Description == "" {
		// "Instagram", "Log in • Instagram": the login page, not the post.
		return Metadata{}, errBlocked
	}
	return m, nil
}

var (
	metaTag   = regexp.MustCompile(`(?is)<meta\s[^>]*>`)
	titleTag  = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	paragraph = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	anyTag    = regexp.MustCompile(`(?s)<[^>]*>`)
)

var attrPatterns = map[string]*regexp.Regexp{}

func init() {
	for _, name := range []string{"property", "name", "content"} {
		attrPatterns[name] = regexp.MustCompile(`(?is)\s` + name + `\s*=\s*("([^"]*)"|'([^']*)'|([^\s>]+))`)
	}
}

// attr reads one attribute ("property", "name" or "content") from a single
// tag, quoted or not.
func attr(tag, name string) string {
	m := attrPatterns[name].FindStringSubmatch(tag)
	if m == nil {
		return ""
	}
	for _, v := range m[2:] {
		if v != "" {
			return html.UnescapeString(v)
		}
	}
	return ""
}

func stripTags(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	return html.UnescapeString(anyTag.ReplaceAllString(s, " "))
}

// clean collapses whitespace, drops control characters and truncates on a
// rune boundary.
func clean(s string, limit int) string {
	s = html.UnescapeString(s)
	var b strings.Builder
	space := false
	for _, r := range s {
		if r == utf8.RuneError || (unicode.IsControl(r) && r != '\n') {
			continue
		}
		if unicode.IsSpace(r) {
			if !space && b.Len() > 0 {
				b.WriteRune(' ')
			}
			space = true
			continue
		}
		space = false
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	if utf8.RuneCountInString(out) > limit {
		out = string([]rune(out)[:limit-1]) + "…"
	}
	return out
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func loginWall(u *url.URL) bool {
	if u == nil {
		return false
	}
	p := strings.ToLower(u.Path)
	return strings.Contains(p, "/login") || strings.Contains(p, "/accounts/login") ||
		strings.Contains(p, "/checkpoint") || strings.HasPrefix(p, "/i/flow/login")
}

func genericSocialTitle(title string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	if t == "" {
		return true
	}
	for _, generic := range []string{"instagram", "facebook", "log in", "login", "masuk", "tiktok - make your day", "threads"} {
		if t == generic || strings.HasPrefix(t, generic+" •") || strings.HasPrefix(t, generic+" |") || strings.HasSuffix(t, "• instagram") {
			return true
		}
	}
	return false
}
