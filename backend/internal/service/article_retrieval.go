// Package service provides article retrieval functionality.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/shaftoe/savetoink/backend/internal/model"
	repoimpl "github.com/shaftoe/savetoink/backend/internal/repository/dynamodb"
)

// GetArticlesResult holds the result of listing articles with pagination (without content).
type GetArticlesResult struct {
	Articles []*model.Article
	Page     int
	PageSize int
	Total    int
	HasMore  bool
}

// GetArticle retrieves a single article by account ID and article ID.
// Returns the full article including all metadata and content.
func (s *Service) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	if articleID == "" {
		return nil, errors.New(consts.ErrInvalidArticleID)
	}

	if s.repo == nil {
		return nil, errors.New("repository not configured")
	}

	article, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return nil, errors.New("article not found")
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return article, nil
}

// GetArticlesMetadata retrieves article metadata for a given account with pagination.
// page starts at 1, pageSize limits the number of articles returned.
// Content field is excluded from returned articles.
// favoriteFilter can be set to true to filter only favorite articles.
func (s *Service) GetArticlesMetadata(
	ctx context.Context,
	accountID string,
	page, pageSize int,
	favoriteFilter *bool,
) (*GetArticlesResult, error) {
	if s.repo == nil {
		return &GetArticlesResult{
			Articles: []*model.Article{},
			Page:     page,
			PageSize: pageSize,
			Total:    0,
			HasMore:  false,
		}, nil
	}

	articles, lastEvaluatedKey, total, err := s.repo.GetMetadataByAccount(ctx, accountID, page, pageSize, favoriteFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}

	if articles == nil {
		articles = []*model.Article{}
	}

	return &GetArticlesResult{
		Articles: articles,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  lastEvaluatedKey != nil,
	}, nil
}

// ToggleFavorite toggles the favorite status of an article.
func (s *Service) ToggleFavorite(ctx context.Context, accountID, articleID string) (bool, error) {
	if s.repo == nil {
		return false, errors.New("repository not configured")
	}

	article, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return false, errors.New("article not found")
		}
		return false, fmt.Errorf("failed to get article: %w", err)
	}

	newFavoriteStatus := !article.Favorite

	err = s.repo.UpdateFavorite(ctx, accountID, articleID, newFavoriteStatus)
	if err != nil {
		return false, fmt.Errorf("failed to update favorite: %w", err)
	}

	return newFavoriteStatus, nil
}
