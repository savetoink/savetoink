// Package epub provides EPUB file generation functionality.
package epub

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/go-shiori/go-epub"
	"github.com/shaftoe/savetoink/backend/consts"
	"github.com/shaftoe/savetoink/backend/model"
)

const (
	// fileModeReadWrite is the file permission for EPUB files (readable by user).
	fileModeReadWrite = 0o644
)

// Generator handles EPUB file generation from article content.
type Generator struct{}

// NewGenerator creates a new EPUB generator instance.
func NewGenerator() *Generator {
	return &Generator{}
}

func buildMetadataHeader(article *model.Article) string {
	var metaLines []string

	sourceInfo := ""
	if article.SiteName != "" {
		sourceInfo = article.SiteName
	}
	if article.SourceDomain != "" {
		if sourceInfo != "" {
			sourceInfo += " (" + article.SourceDomain + ")"
		} else {
			sourceInfo = article.SourceDomain
		}
	}
	if sourceInfo != "" {
		metaLines = append(metaLines, fmt.Sprintf("<p><strong>Source:</strong> %s</p>", sourceInfo))
	}

	if article.ReadingTimeMinutes > 0 {
		metaLines = append(metaLines, fmt.Sprintf("<p><strong>Reading time:</strong> %d min</p>", article.ReadingTimeMinutes))
	}

	if article.PublishedAt != nil && !article.PublishedAt.IsZero() {
		metaLines = append(metaLines,
			fmt.Sprintf("<p><strong>Published:</strong> %s</p>", article.PublishedAt.Format("2006-01-02")))
	}

	if !article.CreatedAt.IsZero() {
		metaLines = append(metaLines,
			fmt.Sprintf("<p><strong>Added:</strong> %s</p>", article.CreatedAt.Format("2006-01-02")))
	}

	if len(metaLines) == 0 {
		return ""
	}

	return `<div style="font-size: 0.85em; color: #666; margin-bottom: 2em; ` +
		`padding: 1em; border-left: 3px solid #ccc; background-color: #f9f9f9;">
` + strings.Join(metaLines, "") + `
</div>`
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
