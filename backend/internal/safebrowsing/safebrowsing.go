// Package safebrowsing checks a URL against Google Safe Browsing v4.
//
// It is optional: without GOOGLE_SAFE_BROWSING_KEY the caller reports the
// integration as "not_configured" instead of pretending a check happened.
package safebrowsing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultEndpoint is the fixed v4 threatMatches endpoint.
const DefaultEndpoint = "https://safebrowsing.googleapis.com/v4/threatMatches:find"

// ErrNotConfigured means no API key was supplied.
var ErrNotConfigured = errors.New("safe browsing not configured")

// Client performs lookups.
type Client struct {
	APIKey   string
	Endpoint string
	HTTP     *http.Client
}

// New builds a client. An empty key makes every Check return ErrNotConfigured.
func New(apiKey string) *Client {
	return &Client{
		APIKey:   apiKey,
		Endpoint: DefaultEndpoint,
		HTTP: &http.Client{
			Timeout: 2500 * time.Millisecond,
			Transport: &http.Transport{
				Proxy:                  nil,
				MaxResponseHeaderBytes: 32 << 10,
				DisableKeepAlives:      true,
			},
		},
	}
}

// Check reports whether the URL matches any threat list. A transport or API
// error is returned as an error, never as "safe".
func (c *Client) Check(ctx context.Context, target string) (bool, error) {
	if c == nil || c.APIKey == "" {
		return false, ErrNotConfigured
	}
	body := map[string]any{
		"client": map[string]string{"clientId": "amba", "clientVersion": "1.0"},
		"threatInfo": map[string]any{
			"threatTypes":      []string{"MALWARE", "SOCIAL_ENGINEERING", "UNWANTED_SOFTWARE", "POTENTIALLY_HARMFUL_APPLICATION"},
			"platformTypes":    []string{"ANY_PLATFORM"},
			"threatEntryTypes": []string{"URL"},
			"threatEntries":    []map[string]string{{"url": target}},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint+"?key="+c.APIKey, bytes.NewReader(payload))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		// The URL carries the API key as a query parameter; never wrap the
		// transport error, which would print the full request URL.
		return false, errors.New("safe browsing request failed")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("safe browsing status %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
	if err != nil {
		return false, err
	}
	var out struct {
		Matches []json.RawMessage `json:"matches"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return false, fmt.Errorf("safe browsing parse: %w", err)
	}
	return len(out.Matches) > 0, nil
}
