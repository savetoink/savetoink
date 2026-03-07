// Package servicetypes provides shared data structures used across service subpackages.
package servicetypes

import (
	"github.com/shaftoe/savetoink/backend/lib/model"
)

// GetArticlesResult holds the result of listing articles with pagination (without content).
type GetArticlesResult struct {
	Articles []*model.Article
	Page     int
	PageSize int
	Total    int
	HasMore  bool
}

// DeleteArticleResult holds the result of deleting an article.
type DeleteArticleResult struct {
	Deleted int
}

// ProcessResult holds the result of processing an article.
type ProcessResult struct {
	article  *model.Article
	epubData []byte
	url      string
}

// Article returns the extracted article.
func (r *ProcessResult) Article() *model.Article {
	return r.article
}

// EPUBData returns the generated EPUB data.
func (r *ProcessResult) EPUBData() []byte {
	return r.epubData
}

// URL returns the URL that was processed.
func (r *ProcessResult) URL() string {
	return r.url
}

// NewProcessResult creates a new ProcessResult for testing purposes.
func NewProcessResult(article *model.Article, epubData []byte, url string) *ProcessResult {
	return &ProcessResult{
		article:  article,
		epubData: epubData,
		url:      url,
	}
}
