// Package articles provides article management and processing services.
package articles

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/epub"
	"github.com/shaftoe/savetoink/backend/lib/service/profile"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
)

// ArticleService handles article CRUD operations.
type ArticleService struct {
	repo        repository.ArticlesRepository
	extractor   *content.Extractor
	publisher   *epub.Publisher
	userProfile *profile.UserProfileService
}

// New creates a new ArticleService instance.
func New(
	repo repository.ArticlesRepository,
	extractor *content.Extractor,
	publisher *epub.Publisher,
	userProfile *profile.UserProfileService,
) *ArticleService {
	return &ArticleService{
		repo:        repo,
		extractor:   extractor,
		publisher:   publisher,
		userProfile: userProfile,
	}
}

// CreateArticle creates an article entry with minimal metadata.
func (s *ArticleService) CreateArticle(ctx context.Context, rawURL, accountID string) (*model.Article, error) {
	cleanURL, err := content.CleanURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to clean url: %w", err)
	}

	articleID, err := content.ArticleIDFromURL(cleanURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate article id: %w", err)
	}

	article := &model.Article{
		Account:   accountID,
		ID:        articleID,
		URL:       cleanURL,
		CreatedAt: time.Now().UTC(),
	}

	if s.repo != nil {
		if storeErr := s.repo.Store(ctx, article); storeErr != nil {
			return nil, fmt.Errorf("failed to store article: %w", storeErr)
		}
	}

	return article, nil
}

// UpdateArticle updates an existing article with full content and metadata.
func (s *ArticleService) UpdateArticle(ctx context.Context, article *model.Article) error {
	if article.Account == "" || article.ID == "" {
		return errors.New("account and ID required for update")
	}

	if s.repo == nil {
		return nil
	}

	if err := s.repo.Store(ctx, article); err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}

	return nil
}

// GetArticle retrieves an article by account ID and article ID.
func (s *ArticleService) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	if articleID == "" {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrInvalid, consts.ErrInvalidArticleID)
	}

	if s.repo == nil {
		return nil, errors.New("repository not configured")
	}

	article, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
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
	if s.repo == nil {
		return &servicetypes.GetArticlesResult{
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

	if s.repo == nil {
		return &servicetypes.DeleteArticleResult{Deleted: 0}, nil
	}

	_, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return &servicetypes.DeleteArticleResult{Deleted: 0}, nil
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	err = s.repo.DeleteByAccountAndID(ctx, accountID, articleID)
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
	if s.repo == nil {
		return &servicetypes.DeleteArticleResult{Deleted: 0}, nil
	}

	deleted, err := s.repo.DeleteByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete all articles: %w", err)
	}

	return &servicetypes.DeleteArticleResult{Deleted: deleted}, nil
}

// ToggleFavorite toggles the favorite status of an article.
func (s *ArticleService) ToggleFavorite(ctx context.Context, accountID, articleID string) (bool, error) {
	if s.repo == nil {
		return false, errors.New("repository not configured")
	}

	article, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return false, apperrors.ErrNotFound
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
