// Package articles provides article management and processing services.
package articles

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	apperrors "github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/epub"
	"github.com/shaftoe/savetoink/backend/lib/internal/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/profile"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	"github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/internal/validation"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

// ArticleService handles article CRUD operations.
type ArticleService struct {
	articlesRepo    repository.ArticlesRepository
	articleTagsRepo repository.ArticleTagsRepository
	publisher       *epub.Publisher
	userProfile     *profile.UserProfileService
}

// New creates a new ArticleService instance.
func New(
	articles repository.ArticlesRepository,
	articleTags repository.ArticleTagsRepository,
	publisher *epub.Publisher,
	userProfile *profile.UserProfileService,
) *ArticleService {
	return &ArticleService{
		articlesRepo:    articles,
		articleTagsRepo: articleTags,
		publisher:       publisher,
		userProfile:     userProfile,
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

	// Populate tags for the article
	if getTagsErr := s.populateTagsForArticles(ctx, accountID, []*model.Article{article}); getTagsErr != nil {
		return nil, fmt.Errorf("failed to populate tags: %w", getTagsErr)
	}

	return article, nil
}

// GetArticlesMetadata retrieves paginated article metadata for an account.
func (s *ArticleService) GetArticlesMetadata(
	ctx context.Context,
	accountID string,
	page, pageSize int,
	filter *types.ArticleFilter,
) (*servicetypes.GetArticlesResult, error) {
	articles, total, err := s.articlesRepo.GetMetadataByAccount(
		ctx, accountID, page, pageSize, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}

	if articles == nil {
		articles = []*model.Article{}
	}

	// Populate tags for articles
	if populateErr := s.populateTagsForArticles(ctx, accountID, articles); populateErr != nil {
		return nil, fmt.Errorf("failed to populate tags: %w", populateErr)
	}

	return &servicetypes.GetArticlesResult{
		Articles: articles,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  (page * pageSize) < total,
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

	// Delete all tags for this article
	deleteTagErr := s.articleTagsRepo.DeleteTagsForArticle(ctx, accountID, articleID)
	if deleteTagErr != nil {
		return nil, fmt.Errorf("failed to delete tags for article: %w", deleteTagErr)
	}

	return &servicetypes.DeleteArticleResult{Deleted: 1}, nil
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

// AddArticleTags adds tags to an article.
func (s *ArticleService) AddArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	// Validate article exists
	_, err := s.articlesRepo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to get article: %w", err)
	}

	// Validate and normalize tags (AddArticleTags requires at least one tag)
	if len(tags) == 0 {
		return fmt.Errorf("%w: at least one tag is required", apperrors.ErrInvalid)
	}

	normalizedTags, validationErr := validation.ValidateTags(tags)
	if validationErr != nil {
		return fmt.Errorf("failed to validate tags: %w", validationErr)
	}

	if addErr := s.articleTagsRepo.AddTagsToArticle(ctx, accountID, articleID, normalizedTags, nil); addErr != nil {
		return fmt.Errorf("failed to add tags to article: %w", addErr)
	}

	return nil
}

// RemoveArticleTags removes specific tags from an article.
func (s *ArticleService) RemoveArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	// Validate article exists
	_, err := s.articlesRepo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to get article: %w", err)
	}

	// Validate and normalize tags
	normalizedTags, validationErr := validation.ValidateTags(tags)
	if validationErr != nil {
		return fmt.Errorf("failed to validate tags: %w", validationErr)
	}

	if removeErr := s.articleTagsRepo.RemoveTagsFromArticle(ctx, accountID, articleID, normalizedTags); removeErr != nil {
		return fmt.Errorf("failed to remove tags from article: %w", removeErr)
	}

	return nil
}

// SetArticleTags replaces all tags for an article with the provided tags.
func (s *ArticleService) SetArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	// Validate article exists
	_, err := s.articlesRepo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to get article: %w", err)
	}

	// Validate and normalize tags (nil tags means remove all tags)
	var normalizedTags []string
	if len(tags) > 0 {
		var validationErr error
		normalizedTags, validationErr = validation.ValidateTags(tags)
		if validationErr != nil {
			return fmt.Errorf("failed to validate tags: %w", validationErr)
		}
	}

	if setErr := s.articleTagsRepo.SetArticleTags(ctx, accountID, articleID, normalizedTags); setErr != nil {
		return fmt.Errorf("failed to set article tags: %w", setErr)
	}

	return nil
}

// GetArticleTags retrieves all tags for a specific article.
func (s *ArticleService) GetArticleTags(ctx context.Context, accountID, articleID string) ([]string, error) {
	// Validate article exists
	_, err := s.articlesRepo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	tags, err := s.articleTagsRepo.GetArticleTags(ctx, accountID, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get article tags: %w", err)
	}

	// Sort tags for consistency
	return sortTags(tags), nil
}

// GetAllTagsForAccount retrieves all unique tags for an account.
func (s *ArticleService) GetAllTagsForAccount(ctx context.Context, accountID string) ([]string, error) {
	tags, err := s.articleTagsRepo.GetAllTagsForAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tags for account: %w", err)
	}

	// Sort tags for consistency
	return sortTags(tags), nil
}

// sortTags sorts tags alphabetically.
func sortTags(tags []string) []string {
	if len(tags) == 0 {
		return tags
	}

	// Deduplicate tags using the shared helper
	unique := validation.DeduplicateStrings(tags)

	// Sort the deduplicated tags
	sort.Strings(unique)
	return unique
}

// populateTagsForArticles populates the Tags field for all provided articles.
// It fetches tags from the article tags repository for each article and assigns them.
func (s *ArticleService) populateTagsForArticles(
	ctx context.Context,
	accountID string,
	articles []*model.Article,
) error {
	if len(articles) == 0 || s.articleTagsRepo == nil {
		return nil
	}

	// Fetch tags for each article individually
	// This is O(n) repository calls. For optimization, consider adding
	// a batch method to the repository interface that can fetch tags for
	// multiple articles in a single call.
	for _, article := range articles {
		tags, err := s.articleTagsRepo.GetArticleTags(ctx, accountID, article.ID)
		if err != nil {
			return fmt.Errorf("failed to get tags for article %s: %w", article.ID, err)
		}
		article.Tags = sortTags(tags)
	}

	return nil
}
