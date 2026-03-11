package content

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"golang.org/x/net/html"
)

// TrafilaturaCleaner extracts article content from html.Node using go-trafilatura.
type TrafilaturaCleaner struct{}

// NewTrafilaturaCleaner creates a new TrafilaturaCleaner instance.
func NewTrafilaturaCleaner() *TrafilaturaCleaner {
	return &TrafilaturaCleaner{}
}

// Clean extracts article content from an html.Node.
func (c *TrafilaturaCleaner) Clean(_ context.Context, doc, node *html.Node, u *url.URL) (*model.Article, error) {
	if u == nil {
		return nil, fmt.Errorf("%w: url is nil", ErrInvalidURL)
	}
	if doc == nil {
		return nil, fmt.Errorf("%w: html document is nil", ErrNilContentNode)
	}
	if node == nil {
		return nil, fmt.Errorf("%w: html node is nil", ErrNilContentNode)
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

	result, err := trafilatura.ExtractDocument(node, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExtractionFailed, err)
	}

	if result.ContentNode == nil {
		return nil, ErrNoContentExtracted
	}

	article := c.buildArticle(result, doc)
	return article, nil
}

func (c *TrafilaturaCleaner) buildArticle(result *trafilatura.ExtractResult, doc *html.Node) *model.Article {
	contentHTML := dom.InnerHTML(result.ContentNode)
	plainText := stripHTML(contentHTML)
	wordCount := countWords(plainText)

	title := result.Metadata.Title
	if title == result.Metadata.Sitename {
		extractedTitle := c.extractTitleFromDocument(doc)
		if extractedTitle != "" {
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

func (c *TrafilaturaCleaner) extractTitleFromDocument(doc *html.Node) string {
	if doc == nil {
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
