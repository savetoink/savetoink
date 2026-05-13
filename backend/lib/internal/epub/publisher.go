// Package epub provides EPUB file generation functionality.
package epub

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/go-shiori/go-epub"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

const (
	emptyDiv = "<div style=\"font-size: 0.85em; color: #666; margin-bottom: 2em; " +
		"padding: 1em; border-left: 3px solid #ccc; background-color: #f9f9f9;\">\n</div>\n"
)

// PublisherOption configures a Publisher.
type PublisherOption func(*Publisher)

// WithMemoryStorage configures the publisher to use in-memory storage
// instead of the filesystem, avoiding the need for /tmp.
func WithMemoryStorage() PublisherOption {
	return func(p *Publisher) {
		p.useMemoryStorage = true
	}
}

//go:embed metadata.tpl
var metadataTemplate string

// Publisher handles EPUB file generation from article content.
type Publisher struct {
	useMemoryStorage bool
}

// NewPublisher creates a new EPUB publisher instance.
func NewPublisher(opts ...PublisherOption) *Publisher {
	p := &Publisher{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// addStylesheet embeds the default CSS stylesheet into the EPUB and returns
// the relative path to the CSS file for use in sections.
func addStylesheet(e *epub.Epub) (string, error) {
	dataURL := "data:text/css;base64," + base64.StdEncoding.EncodeToString([]byte(consts.EPUBStylesheet))
	cssPath, err := e.AddCSS(dataURL, consts.EPUBStylesheetFilename)
	if err != nil {
		return "", fmt.Errorf("failed to add CSS stylesheet: %w", err)
	}
	return cssPath, nil
}

func buildMetadataHeader(article *model.Article) string {
	tmpl := template.Must(template.New("metadata").Funcs(template.FuncMap{
		"sourceInfo": func(a *model.Article) string {
			if a.SiteName == "" && a.SourceDomain == "" {
				return ""
			}
			if a.SiteName != "" && a.SourceDomain != "" {
				return a.SiteName + " (" + a.SourceDomain + ")"
			}
			if a.SiteName != "" {
				return a.SiteName
			}
			return a.SourceDomain
		},
	}).Parse(metadataTemplate))

	var buf strings.Builder
	err := tmpl.Execute(&buf, article)
	if err != nil {
		return ""
	}

	result := buf.String()
	if result == emptyDiv {
		return ""
	}

	return result
}

// GenerateEPUB creates an EPUB file from the given article and returns a ReadCloser.
func (p *Publisher) GenerateEPUB(article *model.Article) (io.ReadCloser, error) {
	if p.useMemoryStorage {
		if err := epub.Use(epub.MemoryFS); err != nil {
			return nil, fmt.Errorf("failed to set memory storage: %w", err)
		}
	}

	e, err := epub.NewEpub(article.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to create EPUB: %w", err)
	}

	if article.Title != "" {
		e.SetTitle(article.Title)
	}

	if article.Author != "" {
		e.SetAuthor(article.Author)
	}

	if article.Excerpt != "" {
		e.SetDescription(article.Excerpt)
	}

	e.SetLang(article.Language)
	if article.Language == "" {
		e.SetLang("en")
	}

	cssPath, err := addStylesheet(e)
	if err != nil {
		return nil, fmt.Errorf("failed to add stylesheet: %w", err)
	}

	articleContent := buildMetadataHeader(article) + article.Content

	_, err = e.AddSection(articleContent, consts.DefaultChapterTitle, consts.DefaultChapterFilename, cssPath)
	if err != nil {
		return nil, fmt.Errorf("failed to add chapter: %w", err)
	}

	e.EmbedImages()

	// Notice: go-epub doesn't support returning a ReadCloser directly
	// so we use a bytes.Buffer and io.NopCloser
	var buffer bytes.Buffer
	_, err = e.WriteTo(&buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to write EPUB: %w", err)
	}

	return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
}
