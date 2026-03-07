// Package content provides article extraction functionality from web pages.
package content

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/shaftoe/savetoink/backend/internal/model"
	"github.com/shaftoe/savetoink/backend/internal/validation"
	"golang.org/x/net/html"
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
		_ = body.Close()
	}()

	htmlBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
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
	req.Header.Set("User-Agent", consts.GetRandomUserAgent())

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		_ = resp.Body.Close()
		return nil, nil, fmt.Errorf("expected HTML content, got: %s", contentType)
	}

	return parsedURL, resp.Body, nil
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

func validateURL(urlStr string) error {
	if err := validation.ValidateURLOnlyFormat(urlStr); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}
