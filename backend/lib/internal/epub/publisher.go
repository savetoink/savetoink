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
	"github.com/vincent-petithory/dataurl"
)

const (
	emptyDiv = "<div style=\"font-size: 0.85em; color: #666; margin-bottom: 2em; " +
		"padding: 1em; border-left: 3px solid #ccc; background-color: #f9f9f9;\">\n</div>\n"

	defaultCSSFilename = "content.css"

	defaultCSSContent = `pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  margin: 1em 0;
  padding: 0.75em;
  background-color: #f4f4f4;
  border: 1px solid #ddd;
  border-radius: 3px;
  font-size: 0.9em;
  line-height: 1.4;
  overflow-wrap: break-word;
}

code {
  font-family: monospace;
  font-size: 0.9em;
  background-color: #f4f4f4;
  padding: 0.15em 0.3em;
  border-radius: 3px;
}

pre code {
  padding: 0;
  background-color: transparent;
  border-radius: 0;
}
`
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

// addContentCSS adds the default content stylesheet to the EPUB and returns
// the internal path to the CSS file.
func (p *Publisher) addContentCSS(e *epub.Epub) (string, error) {
	cssDataURL := dataurl.EncodeBytes([]byte(defaultCSSContent))
	cssPath, err := e.AddCSS(cssDataURL, defaultCSSFilename)
	if err != nil {
		return "", fmt.Errorf("failed to add content css: %w", err)
	}
	return cssPath, nil
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

	articleContent := buildMetadataHeader(article) + article.Content

	cssPath, err := p.addContentCSS(e)
	if err != nil {
		return nil, err
	}

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
