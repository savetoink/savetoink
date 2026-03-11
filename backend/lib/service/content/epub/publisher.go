// Package epub provides EPUB file generation functionality.
package epub

import (
	"bytes"
	_ "embed"
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

//go:embed metadata.tpl
var metadataTemplate string

// Publisher handles EPUB file generation from article content.
type Publisher struct{}

// NewPublisher creates a new EPUB publisher instance.
func NewPublisher() *Publisher {
	return &Publisher{}
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

	articleContent := buildMetadataHeader(article) + article.Content

	_, err = e.AddSection(articleContent, consts.DefaultChapterTitle, consts.DefaultChapterFilename, "")
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
