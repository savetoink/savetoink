// Package processing provides article content extraction and EPUB generation.
package processing

import (
	"context"
	"errors"
	"fmt"

	"github.com/shaftoe/savetoink/backend/internal/service/content"
	"github.com/shaftoe/savetoink/backend/internal/service/epub"
	"github.com/shaftoe/savetoink/backend/internal/service/servicetypes"
)

// ArticleProcessingService handles article extraction and EPUB generation.
type ArticleProcessingService struct {
	extractor *content.Extractor
	Generator *epub.Generator
}

// New creates a new ArticleProcessingService instance.
func New(extractor *content.Extractor, generator *epub.Generator) *ArticleProcessingService {
	return &ArticleProcessingService{
		extractor: extractor,
		Generator: generator,
	}
}

// Process extracts an article from a URL and generates an EPUB.
func (s *ArticleProcessingService) Process(ctx context.Context, url string) (*servicetypes.ProcessResult, error) {
	article, err := s.extractor.ExtractFromURL(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to extract article: %w", err)
	}

	if article.Title == "" {
		article.Title = "Untitled"
	}

	epubData, err := s.Generator.Generate(article)
	if err != nil {
		return nil, fmt.Errorf("failed to generate EPUB: %w", err)
	}

	return servicetypes.NewProcessResult(article, epubData, url), nil
}

// WriteToFile writes the processed article as an EPUB file to disk.
func (s *ArticleProcessingService) WriteToFile(result *servicetypes.ProcessResult, outputPath string) error {
	if result == nil {
		return errors.New("result is nil, must call Process first")
	}

	if result.Article() == nil {
		return errors.New("article is nil, must call Process first")
	}

	if outputPath == "" {
		return errors.New("output path is empty")
	}

	err := s.Generator.GenerateAndWrite(result.Article(), outputPath)
	if err != nil {
		return fmt.Errorf("failed to write EPUB document: %w", err)
	}

	return nil
}
