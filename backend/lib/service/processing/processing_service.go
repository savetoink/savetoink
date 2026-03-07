// Package processing provides article content extraction and EPUB generation.
package processing

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/epub"
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

// Fetch fetches HTML content from a URL.
func (s *ArticleProcessingService) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	reader, err := s.extractor.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url: %w", err)
	}
	return reader, nil
}

// Extract extracts an article from HTML content.
func (s *ArticleProcessingService) Extract(ctx context.Context, htmlReader io.Reader) (*model.Article, error) {
	article, err := s.extractor.ExtractFromReader(ctx, htmlReader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract from reader: %w", err)
	}
	return article, nil
}

// WriteToFile writes an article as an EPUB file to disk.
func (s *ArticleProcessingService) WriteToFile(article *model.Article, outputPath string) error {
	if article == nil {
		return errors.New("article is nil")
	}

	if outputPath == "" {
		return errors.New("output path is empty")
	}

	err := s.Generator.GenerateAndWrite(article, outputPath)
	if err != nil {
		return fmt.Errorf("failed to write EPUB document: %w", err)
	}

	return nil
}
