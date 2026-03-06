// Package service provides article deletion functionality.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shaftoe/savetoink/backend/internal/consts"
	repoimpl "github.com/shaftoe/savetoink/backend/internal/repository/dynamodb"
)

// DeleteArticleResult holds the result of deleting an article.
type DeleteArticleResult struct {
	Deleted int
}

// DeleteArticle deletes a single article by account and ID.
func (s *Service) DeleteArticle(ctx context.Context, accountID, articleID string) (*DeleteArticleResult, error) {
	if articleID == "" {
		return nil, errors.New(consts.ErrInvalidArticleID)
	}

	if s.repo == nil {
		return &DeleteArticleResult{Deleted: 0}, nil
	}

	_, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return &DeleteArticleResult{Deleted: 0}, nil
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	err = s.repo.DeleteByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete article: %w", err)
	}

	return &DeleteArticleResult{Deleted: 1}, nil
}

// DeleteAllArticles deletes all articles for a given account.
func (s *Service) DeleteAllArticles(ctx context.Context, accountID string) (*DeleteArticleResult, error) {
	if s.repo == nil {
		return &DeleteArticleResult{Deleted: 0}, nil
	}

	deleted, err := s.repo.DeleteByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete all articles: %w", err)
	}

	return &DeleteArticleResult{Deleted: deleted}, nil
}
