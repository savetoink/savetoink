// Package epub provides EPUB file generation functionality.
package epub

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/go-shiori/go-epub"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

const (
	// fileModeReadWrite is the file permission for EPUB files (readable by user).
	fileModeReadWrite = 0o644

	// metadataTemplate is the template used to generate the metadata header.
	metadataTemplate = `<div style="font-size: 0.85em; color: #666; margin-bottom: 2em; ` +
		`padding: 1em; border-left: 3px solid #ccc; background-color: #f9f9f9;">
` + `{{- if .Title}}<p><strong>Title: {{.Title}}</strong></p>
` + `{{- end}}{{- if .Author}}<p><strong>Author: {{.Author}}</strong></p>
` + `{{- end}}{{- if (sourceInfo .)}}<p><strong>Source: {{sourceInfo .}}</strong></p>
` + `{{- end}}{{- if gt .ReadingTimeMinutes 0}}<p><strong>Reading time: ` +
		`{{.ReadingTimeMinutes}} min</strong></p>
` + `{{- end}}{{- if and .PublishedAt (not .PublishedAt.IsZero)}}<p><strong>Published: ` +
		`{{.PublishedAt.Format "2006-01-02T15:04:05Z07:00"}}</strong></p>
` + `{{- end}}{{- if not .CreatedAt.IsZero}}<p><strong>Added: ` +
		`{{.CreatedAt.Format "2006-01-02T15:04:05Z07:00"}}</strong></p>
` + `{{- end}}
</div>`
)

// Generator handles EPUB file generation from article content.
type Generator struct{}

// NewGenerator creates a new EPUB generator instance.
func NewGenerator() *Generator {
	return &Generator{}
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
	emptyDiv := `<div style="font-size: 0.85em; color: #666; margin-bottom: 2em; ` +
		`padding: 1em; border-left: 3px solid #ccc; background-color: #f9f9f9;">
</div>`
	if result == emptyDiv {
		return ""
	}

	return result
}

// Generate creates an EPUB file from the given article and returns its bytes.
func (g *Generator) Generate(article *model.Article) ([]byte, error) {
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

	var buffer bytes.Buffer
	_, err = e.WriteTo(&buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to write EPUB: %w", err)
	}

	return buffer.Bytes(), nil
}

// GenerateAndWrite generates an EPUB file and writes it to the specified path.
func (g *Generator) GenerateAndWrite(article *model.Article, outputPath string) error {
	data, err := g.Generate(article)
	if err != nil {
		return err
	}

	// #nosec G306 - EPUB files need to be readable by user
	if writeErr := os.WriteFile(outputPath, data, fileModeReadWrite); writeErr != nil {
		return fmt.Errorf("failed to write EPUB file: %w", writeErr)
	}

	return nil
}
