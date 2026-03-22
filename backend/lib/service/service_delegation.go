package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"golang.org/x/net/html"
)

// Fetch fetches HTML content from a URL.
func (s *Service) Fetch(ctx context.Context, u *url.URL) (*content.FetchedContent, error) {
	result, err := s.fetcher.Fetch(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url: %w", err)
	}

	return result, nil
}

// ParseHTMLFromSource fetches or reads HTML content from a URL or local file.
func (s *Service) ParseHTMLFromSource(ctx context.Context, u *url.URL) (*html.Node, error) {
	var fetched *content.FetchedContent
	var err error

	if u.Scheme == "file" {
		fetched, err = s.fetchFromFile(u)
	} else {
		fetched, err = s.Fetch(ctx, u)
	}
	if err != nil {
		return nil, err
	}

	doc, err := s.ParseHTML(ctx, fetched)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *Service) fetchFromFile(u *url.URL) (*content.FetchedContent, error) {
	filePath := filepath.Clean(u.Path)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &content.FetchedContent{
		HTML: file,
		URL:  u,
		Type: content.FetcherTypeGo,
	}, nil
}

// ParseHTML parses HTML content from fetched content into a DOM node.
func (s *Service) ParseHTML(ctx context.Context, fetched *content.FetchedContent) (*html.Node, error) {
	defer func() {
		_ = fetched.HTML.Close()
	}()

	doc, err := s.extractor.Extract(ctx, fetched.HTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	return doc, nil
}

// Clean extracts article content from a DOM node.
func (s *Service) Clean(ctx context.Context, doc *html.Node, u *url.URL) (*model.Article, error) {
	article, err := s.cleaner.Clean(ctx, doc, u)
	if err != nil {
		return nil, fmt.Errorf("failed to clean content: %w", err)
	}

	return article, nil
}

// GenerateEPUB generates an EPUB from an article.
func (s *Service) GenerateEPUB(article *model.Article) (io.ReadCloser, error) {
	epubReader, err := s.publisher.GenerateEPUB(article)
	if err != nil {
		return nil, fmt.Errorf("failed to generate epub: %w", err)
	}
	return epubReader, nil
}

// ReadEPUB reads an EPUB file from a URL and returns the file reader and title.
func (s *Service) ReadEPUB(ctx context.Context, u *url.URL) (io.ReadCloser, string, error) {
	epubReader, title, err := s.reader.ReadFromURL(ctx, u)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read epub: %w", err)
	}
	return epubReader, title, nil
}
