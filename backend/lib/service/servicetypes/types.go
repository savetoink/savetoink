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
