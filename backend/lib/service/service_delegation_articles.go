package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	"github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

// CreateArticle delegates to ArticleService.
func (s *Service) CreateArticle(ctx context.Context, u *url.URL, accountID string) (*model.Article, error) {
	article, err := s.articles.CreateArticle(ctx, u, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}
	return article, nil
}

// GetArticle delegates to ArticleService.
func (s *Service) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	article, err := s.articles.GetArticle(ctx, accountID, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}
	return article, nil
}

// UpdateArticle delegates to ArticleService.
func (s *Service) UpdateArticle(ctx context.Context, article *model.Article) error {
	if err := s.articles.UpdateArticle(ctx, article); err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}
	return nil
}

// GetArticlesMetadata delegates to ArticleService.
func (s *Service) GetArticlesMetadata(
	ctx context.Context,
	accountID string,
	page, pageSize int,
	filter *types.ArticleFilter,
) (*servicetypes.GetArticlesResult, error) {
	result, err := s.articles.GetArticlesMetadata(ctx, accountID, page, pageSize, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles metadata: %w", err)
	}
	return result, nil
}

// DeleteArticle delegates to ArticleService.
func (s *Service) DeleteArticle(
	ctx context.Context,
	accountID, articleID string,
) (*servicetypes.DeleteArticleResult, error) {
	result, err := s.articles.DeleteArticle(ctx, accountID, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete article: %w", err)
	}
	return result, nil
}

// ToggleFavorite delegates to ArticleService.
func (s *Service) ToggleFavorite(ctx context.Context, accountID, articleID string) (bool, error) {
	favorite, err := s.articles.ToggleFavorite(ctx, accountID, articleID)
	if err != nil {
		return false, fmt.Errorf("failed to toggle favorite: %w", err)
	}
	return favorite, nil
}

// AddArticleTags delegates to ArticleService.
func (s *Service) AddArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	if err := s.articles.AddArticleTags(ctx, accountID, articleID, tags); err != nil {
		return fmt.Errorf("failed to add tags to article: %w", err)
	}
	return nil
}

// RemoveArticleTags delegates to ArticleService.
func (s *Service) RemoveArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	if err := s.articles.RemoveArticleTags(ctx, accountID, articleID, tags); err != nil {
		return fmt.Errorf("failed to remove tags from article: %w", err)
	}
	return nil
}

// SetArticleTags delegates to ArticleService.
func (s *Service) SetArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	if err := s.articles.SetArticleTags(ctx, accountID, articleID, tags); err != nil {
		return fmt.Errorf("failed to set article tags: %w", err)
	}
	return nil
}

// GetArticleTags delegates to ArticleService.
func (s *Service) GetArticleTags(ctx context.Context, accountID, articleID string) ([]string, error) {
	tags, err := s.articles.GetArticleTags(ctx, accountID, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get article tags: %w", err)
	}
	return tags, nil
}

// GetAllTagsForAccount delegates to ArticleService.
func (s *Service) GetAllTagsForAccount(ctx context.Context, accountID string) ([]string, error) {
	tags, err := s.articles.GetAllTagsForAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tags for account: %w", err)
	}
	return tags, nil
}
