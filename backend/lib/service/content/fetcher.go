// Package content provides article extraction functionality from web pages.
package content

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/validation"
)

// Fetcher handles fetching HTML content from URLs.
type Fetcher struct {
	client         *http.Client
	browserlessKey string
}

// NewFetcher creates a new Fetcher instance.
func NewFetcher(browserlessKey string) *Fetcher {
	return &Fetcher{
		client:         &http.Client{},
		browserlessKey: browserlessKey,
	}
}

// Fetch fetches HTML content from a URL and returns the bytes.
// Falls back to Browserless API if the simple HTTP fetch fails and a browserless key is configured.
func (f *Fetcher) Fetch(ctx context.Context, urlStr string) ([]byte, error) {
	if err := validateURL(urlStr); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	_, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", consts.GetRandomUserAgent())

	resp, err := f.client.Do(req)
	if err != nil {
		if f.browserlessKey != "" {
			return f.fetchWithBrowserless(ctx, urlStr)
		}
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		if f.browserlessKey != "" {
			return f.fetchWithBrowserless(ctx, urlStr)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		_ = resp.Body.Close()
		if f.browserlessKey != "" {
			return f.fetchWithBrowserless(ctx, urlStr)
		}
		return nil, fmt.Errorf("expected HTML content, got: %s", contentType)
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		if f.browserlessKey != "" {
			return f.fetchWithBrowserless(ctx, urlStr)
		}
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return htmlBytes, nil
}

// fetchWithBrowserless fetches HTML content using the Browserless content API.
func (f *Fetcher) fetchWithBrowserless(ctx context.Context, urlStr string) ([]byte, error) {
	browserlessURL := fmt.Sprintf("%s?token=%s", consts.BrowserlessContentURL, f.browserlessKey)

	requestBody := struct {
		URL string `json:"url"`
	}{
		URL: urlStr,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal browserless request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, browserlessURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create browserless request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch via browserless: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("browserless returned status code: %d", resp.StatusCode)
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read browserless response: %w", err)
	}

	if err = validateHTMLContent(htmlBytes); err != nil {
		return nil, fmt.Errorf("browserless returned invalid content: %w", err)
	}

	return htmlBytes, nil
}

func validateURL(urlStr string) error {
	if err := validation.ValidateURLOnlyFormat(urlStr); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}

func validateHTMLContent(htmlBytes []byte) error {
	content := strings.ToLower(string(htmlBytes))

	for _, pattern := range consts.HTMLErrorPatterns {
		if strings.Contains(content, strings.ToLower(pattern)) {
			return fmt.Errorf("content appears to be an error page (contains %q)", pattern)
		}
	}

	if !strings.Contains(content, "<html") && !strings.Contains(content, "<!doctype") {
		return errors.New("content does not appear to be valid HTML")
	}

	return nil
}
