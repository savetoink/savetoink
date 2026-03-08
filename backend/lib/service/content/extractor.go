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
	"time"

	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/validation"
	"golang.org/x/net/html"
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

// Extractor handles the extraction of article content from HTML.
type Extractor struct{}

// NewExtractor creates a new Extractor instance.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// GenerateFromHTML extracts article content from HTML bytes.
func (e *Extractor) GenerateFromHTML(_ context.Context, htmlBytes []byte) (*model.Article, error) {
	parsedURL, _ := url.Parse("https://example.com")
	_ = parsedURL
	opts := trafilatura.Options{
		OriginalURL:    parsedURL,
		EnableFallback: true,
		Config: &trafilatura.Config{
			MinExtractedSize: consts.MinimumExtractedSize,
			MinOutputSize:    consts.MinimumOutputSize,
		},
	}

	result, err := trafilatura.Extract(strings.NewReader(string(htmlBytes)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to extract article content: %w", err)
	}

	if result.ContentNode == nil {
		return nil, errors.New("no content extracted")
	}

	article := e.buildArticle(result, htmlBytes)
	return article, nil
}

// ExtractFromURL fetches and extracts article content from a URL.
// This is a convenience method for testing purposes. In production, callers
// should use Fetcher.Fetch() followed by Extractor.GenerateFromHTML().
func (e *Extractor) ExtractFromURL(ctx context.Context, urlStr string) (*model.Article, error) {
	fetcher := NewFetcher("")
	htmlBytes, err := fetcher.Fetch(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	opts := trafilatura.Options{
		OriginalURL:    parsedURL,
		EnableFallback: true,
		Config: &trafilatura.Config{
			MinExtractedSize: consts.MinimumExtractedSize,
			MinOutputSize:    consts.MinimumOutputSize,
		},
	}

	result, err := trafilatura.Extract(strings.NewReader(string(htmlBytes)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to extract article content: %w", err)
	}

	if result.ContentNode == nil {
		return nil, errors.New("no content extracted")
	}

	article := e.buildArticle(result, htmlBytes)
	return article, nil
}

func (e *Extractor) buildArticle(result *trafilatura.ExtractResult, htmlBytes []byte) *model.Article {
	contentHTML := dom.InnerHTML(result.ContentNode)
	plainText := stripHTML(contentHTML)
	wordCount := countWords(plainText)

	title := result.Metadata.Title
	if title == result.Metadata.Sitename {
		if extractedTitle := extractTitleFromHTML(htmlBytes); extractedTitle != "" {
			title = extractedTitle
		}
	}

	return &model.Article{
		Title:              title,
		Author:             result.Metadata.Author,
		Content:            contentHTML,
		Excerpt:            result.Metadata.Description,
		ImageURL:           result.Metadata.Image,
		PublishedAt:        toTimePtr(result.Metadata.Date),
		URL:                result.Metadata.URL,
		CreatedAt:          time.Now().UTC(),
		WordCount:          wordCount,
		ReadingTimeMinutes: (wordCount + consts.WordsPerMinute - 1) / consts.WordsPerMinute,
		SourceDomain:       result.Metadata.Hostname,
		SiteName:           result.Metadata.Sitename,
		ContentType:        result.Metadata.PageType,
		Language:           result.Metadata.Language,
	}
}

func extractTitleFromHTML(htmlBytes []byte) string {
	doc, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return ""
	}

	titleNodes := dom.QuerySelectorAll(doc, "title")
	if len(titleNodes) > 0 {
		fullTitle := dom.TextContent(titleNodes[0])
		title := cleanTitle(fullTitle)
		if title != "" {
			return title
		}
	}

	h2Nodes := dom.QuerySelectorAll(doc, "h2")
	if len(h2Nodes) > 0 {
		h2Text := strings.TrimSpace(dom.TextContent(h2Nodes[0]))
		if h2Text != "" {
			return h2Text
		}
	}

	return ""
}

// cleanTitle extracts the article title from a full title string by removing
// the sitename suffix. This is needed because some websites format their
// <title> tags as "Article Title - Site Name", and we want just the article
// portion when the title extraction fails and we fall back to parsing the
// <title> tag directly.
func cleanTitle(fullTitle string) string {
	parts := strings.Split(fullTitle, consts.TitleSeparator)
	if len(parts) < consts.MinTitleParts {
		return ""
	}

	articleTitle := strings.Join(parts[:len(parts)-1], consts.TitleSeparator)
	articleTitle = strings.TrimSpace(articleTitle)

	if articleTitle == "" {
		return ""
	}

	return articleTitle
}

func toTimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func stripHTML(s string) string {
	re := strings.NewReplacer(
		consts.HTMLTagP, " ",
		consts.HTMLTagPEnd, " ",
		consts.HTMLTagDiv, " ",
		consts.HTMLTagDivEnd, " ",
		consts.HTMLTagBr, " ",
		consts.HTMLTagBrSelfClosing, " ",
	)

	result := re.Replace(s)

	var stripped strings.Builder
	inTag := false
	for _, r := range result {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				stripped.WriteRune(r)
			}
		}
	}

	return stripped.String()
}

func countWords(text string) int {
	fields := strings.Fields(text)
	return len(fields)
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

func validateURL(urlStr string) error {
	if err := validation.ValidateURLOnlyFormat(urlStr); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}
