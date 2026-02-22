// Package content provides article extraction functionality from web pages.
package content

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
)

// Extractor handles the extraction of article content from URLs and HTML.
type Extractor struct {
	client *http.Client
}

// NewExtractor creates a new Extractor instance.
func NewExtractor() *Extractor {
	return &Extractor{
		client: &http.Client{},
	}
}

// ExtractFromURL fetches and extracts article content from given URL.
func (e *Extractor) ExtractFromURL(ctx context.Context, urlStr string) (*model.Article, error) {
	parsedURL, body, err := e.fetchURL(ctx, urlStr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := body.Close(); closeErr != nil {
			log.Printf("warning: failed to close response body: %v", closeErr)
		}
	}()

	opts := trafilatura.Options{
		OriginalURL: parsedURL,
	}

	result, err := trafilatura.Extract(body, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to extract article content: %w", err)
	}

	if result.ContentNode == nil {
		return nil, errors.New("no content extracted")
	}

	return e.buildArticle(result, urlStr), nil
}

func (e *Extractor) fetchURL(ctx context.Context, urlStr string) (*url.URL, io.ReadCloser, error) {
	if err := validateURL(urlStr); err != nil {
		return nil, nil, fmt.Errorf("invalid URL: %w", err)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, http.NoBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.client.Do(req) //nolint:gosec // SSRF is acceptable for fetching article content from user-provided URLs
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("warning: failed to close response body: %v", closeErr)
		}
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("warning: failed to close response body: %v", closeErr)
		}
		return nil, nil, fmt.Errorf("expected HTML content, got: %s", contentType)
	}

	return parsedURL, resp.Body, nil
}

func (e *Extractor) buildArticle(result *trafilatura.ExtractResult, urlStr string) *model.Article {
	contentHTML := dom.InnerHTML(result.ContentNode)
	plainText := stripHTML(contentHTML)
	wordCount := countWords(plainText)

	return &model.Article{
		Title:              result.Metadata.Title,
		Author:             result.Metadata.Author,
		Content:            contentHTML,
		Excerpt:            result.Metadata.Description,
		ImageURL:           result.Metadata.Image,
		PublishedAt:        toTimePtr(result.Metadata.Date),
		URL:                urlStr,
		CreatedAt:          time.Now().UTC(),
		WordCount:          wordCount,
		ReadingTimeMinutes: (wordCount + consts.WordsPerMinute - 1) / consts.WordsPerMinute,
		SourceDomain:       result.Metadata.Hostname,
		SiteName:           result.Metadata.Sitename,
		ContentType:        result.Metadata.PageType,
		Language:           result.Metadata.Language,
	}
}

func toTimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func stripHTML(html string) string {
	re := strings.NewReplacer(
		"<p>", " ",
		"</p>", " ",
		"<div>", " ",
		"</div>", " ",
		"<br>", " ",
		"<br/>", " ",
		"<br />", " ",
	)

	result := re.Replace(html)

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

func validateURL(urlStr string) error {
	if urlStr == "" {
		return errors.New("url cannot be empty")
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url must use http or https scheme")
	}

	if u.Host == "" {
		return errors.New("url must have a host")
	}

	return nil
}
