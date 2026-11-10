// Package pagefetch turns a public web page into the small amount of text an
// audit needs: title, description, and visible body text. It uses the
// SSRF-safe client, caps the body, and never writes to disk.
package pagefetch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"

	"github.com/amba/amba/backend/internal/security"
)

// MaxText bounds what is sent to the model. A page can be megabytes; the
// first few thousand characters carry the claim.
const MaxText = 6000

// Page is the extracted content.
type Page struct {
	Title       string
	Description string
	Text        string
	FinalURL    string
	ContentType string
}

// ErrNotHTML means the response is not a page this package can read.
var ErrNotHTML = errors.New("not an html page")

// Fetch retrieves rawURL and extracts readable text.
func Fetch(ctx context.Context, client *http.Client, rawURL string, timeout time.Duration) (Page, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	body, contentType, finalURL, err := fetchBody(ctx, client, rawURL)
	if err != nil {
		return Page{}, err
	}
	if !isHTML(contentType, body) {
		return Page{ContentType: contentType, FinalURL: finalURL}, ErrNotHTML
	}
	return Extract(body, finalURL), nil
}

func fetchBody(ctx context.Context, client *http.Client, rawURL string) ([]byte, string, string, error) {
	u, err := security.ValidateFetchURL(rawURL)
	if err != nil {
		return nil, "", "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", "", err
	}
	req.Header.Set("User-Agent", "AMBA/1.0 (+literacy check)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,text/plain;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	// Redirects are normal (many sites 301 http->https or add a trailing
	// slash). The client's own CheckRedirect policy still bounds the hops and
	// re-checks every dialled address, so following here is safe.
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, "", "", errors.New("unexpected status")
	}
	body, err := readLimited(resp.Body)
	if err != nil {
		return nil, "", "", err
	}
	final := u.String()
	if resp.Request != nil && resp.Request.URL != nil {
		final = resp.Request.URL.String()
	}
	return body, resp.Header.Get("Content-Type"), final, nil
}

func readLimited(r io.Reader) ([]byte, error) {
	const maxBody = 2 << 20
	return io.ReadAll(io.LimitReader(r, maxBody))
}

func isHTML(contentType string, body []byte) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml") {
		return true
	}
	if ct != "" {
		return false
	}
	head := strings.ToLower(strings.TrimSpace(string(body[:min(len(body), 512)])))
	return strings.HasPrefix(head, "<!doctype html") || strings.HasPrefix(head, "<html")
}

// Extract parses HTML and returns the readable parts.
func Extract(body []byte, finalURL string) Page {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return Page{FinalURL: finalURL, Text: collapse(stripTags(string(body)))}
	}

	page := Page{FinalURL: finalURL}
	var text strings.Builder

	var walk func(*html.Node, bool)
	walk = func(n *html.Node, skip bool) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "noscript", "svg", "template", "iframe":
				skip = true
			case "title":
				if page.Title == "" {
					page.Title = collapse(nodeText(n))
				}
			case "meta":
				if page.Description == "" {
					page.Description = metaDescription(n)
				}
			}
		}
		if !skip && n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, skip)
		}
	}
	walk(doc, false)

	page.Text = truncate(collapse(text.String()), MaxText)
	if page.Title == "" {
		page.Title = truncate(page.Text, 120)
	}
	return page
}

func metaDescription(n *html.Node) string {
	var name, content string
	for _, a := range n.Attr {
		switch strings.ToLower(a.Key) {
		case "name", "property":
			name = strings.ToLower(a.Val)
		case "content":
			content = a.Val
		}
	}
	if (name == "description" || name == "og:description") && content != "" {
		return truncate(collapse(content), 300)
	}
	return ""
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

func stripTags(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch r {
		case '<':
			in = true
		case '>':
			in = false
		default:
			if !in {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func collapse(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	space := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			space = true
			continue
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(r)
	}
	return b.String()
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
