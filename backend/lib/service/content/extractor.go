// Package content provides article extraction functionality from web pages.
package content

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

// Extractor handles the extraction of article content from HTML.
type Extractor struct{}

// NewExtractor creates a new Extractor instance.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// GenerateFromHTML extracts article content from HTML bytes.
func (e *Extractor) GenerateFromHTML(_ context.Context, htmlBytes []byte, u *url.URL) (*model.Article, error) {
	if u == nil {
		return nil, errors.New("url is nil")
	}

	opts := trafilatura.Options{
		OriginalURL:    u,
		EnableFallback: true,
		IncludeImages:  true,
		IncludeLinks:   true,
		Config: &trafilatura.Config{
			MinExtractedSize: consts.MinimumExtractedSize,
			MinOutputSize:    consts.MinimumOutputSize,
		},
	}

	result, err := trafilatura.Extract(bytes.NewReader(htmlBytes), opts)
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
	u, err := validation.ValidateURL(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	result, err := fetcher.Fetch(ctx, u)
	if err != nil {
		return nil, err
	}

	opts := trafilatura.Options{
		OriginalURL:    u,
		EnableFallback: true,
		Config: &trafilatura.Config{
			MinExtractedSize: consts.MinimumExtractedSize,
			MinOutputSize:    consts.MinimumOutputSize,
		},
	}

	trafilaturaResult, err := trafilatura.Extract(bytes.NewReader(result.HTML), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to extract article content: %w", err)
	}

	if trafilaturaResult.ContentNode == nil {
		return nil, errors.New("no content extracted")
	}

	article := e.buildArticle(trafilaturaResult, result.HTML)
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
