// Package articles provides article management and processing services.
package articles

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	apperrors "github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/content/epub"
	"github.com/shaftoe/savetoink/backend/lib/service/profile"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
)

// ArticleService handles article CRUD operations.
type ArticleService struct {
	articlesRepo repository.ArticlesRepository
	publisher    *epub.Publisher
	userProfile  *profile.UserProfileService
}

// New creates a new ArticleService instance.
func New(
	articles repository.ArticlesRepository,
	publisher *epub.Publisher,
	userProfile *profile.UserProfileService,
) *ArticleService {
	return &ArticleService{
		articlesRepo: articles,
		publisher:    publisher,
		userProfile:  userProfile,
	}
}

// CreateArticle creates an article entry with minimal metadata.
func (s *ArticleService) CreateArticle(ctx context.Context, u *url.URL, accountID string) (*model.Article, error) {
	cleanURL := content.CleanURL(u)

	articleID, err := content.ArticleIDFromURL(u)
	if err != nil {
		return nil, fmt.Errorf("failed to generate article id: %w", err)
	}

	article := &model.Article{
		Account:   accountID,
		ID:        articleID,
		URL:       cleanURL,
		CreatedAt: time.Now().UTC(),
	}

	if storeErr := s.articlesRepo.Store(ctx, article); storeErr != nil {
		return nil, fmt.Errorf("failed to store article: %w", storeErr)
	}

	return article, nil
}

// UpdateArticle updates an existing article with full content and metadata.
func (s *ArticleService) UpdateArticle(ctx context.Context, article *model.Article) error {
	if article.Account == "" || article.ID == "" {
		return errors.New("account and ID required for update")
	}

	if err := s.articlesRepo.Store(ctx, article); err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}

	return nil
}

// GetArticle retrieves an article by account ID and article ID.
func (s *ArticleService) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	if articleID == "" {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrInvalid, consts.ErrInvalidArticleID)
	}

	article, err := s.articlesRepo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return article, nil
}

// GetArticlesMetadata retrieves paginated article metadata for an account.
func (s *ArticleService) GetArticlesMetadata(
	ctx context.Context,
	accountID string,
	page, pageSize int,
	favoriteFilter *bool,
) (*servicetypes.GetArticlesResult, error) {
	articles, lastEvaluatedKey, total, err := s.articlesRepo.GetMetadataByAccount(
		ctx, accountID, page, pageSize, favoriteFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}

	if articles == nil {
		articles = []*model.Article{}
	}

	return &servicetypes.GetArticlesResult{
		Articles: articles,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  lastEvaluatedKey != nil,
	}, nil
}

// DeleteArticle removes an article by account ID and article ID.
func (s *ArticleService) DeleteArticle(
	ctx context.Context,
	accountID, articleID string,
) (*servicetypes.DeleteArticleResult, error) {
	if articleID == "" {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrInvalid, consts.ErrInvalidArticleID)
	}

	_, err := s.articlesRepo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return &servicetypes.DeleteArticleResult{Deleted: 0}, nil
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	err = s.articlesRepo.DeleteByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete article: %w", err)
	}

	return &servicetypes.DeleteArticleResult{Deleted: 1}, nil
}

// DeleteAllArticles removes all articles for an account.
func (s *ArticleService) DeleteAllArticles(
	ctx context.Context,
	accountID string,
) (*servicetypes.DeleteArticleResult, error) {
	deleted, err := s.articlesRepo.DeleteByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete all articles: %w", err)
	}

	return &servicetypes.DeleteArticleResult{Deleted: deleted}, nil
}

// ToggleFavorite toggles the favorite status of an article.
func (s *ArticleService) ToggleFavorite(ctx context.Context, accountID, articleID string) (bool, error) {
	article, err := s.articlesRepo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return false, apperrors.ErrNotFound
		}
		return false, fmt.Errorf("failed to get article: %w", err)
	}

	newFavoriteStatus := !article.Favorite

	err = s.articlesRepo.UpdateFavorite(ctx, accountID, articleID, newFavoriteStatus)
	if err != nil {
		return false, fmt.Errorf("failed to update favorite: %w", err)
	}

	return newFavoriteStatus, nil
}
