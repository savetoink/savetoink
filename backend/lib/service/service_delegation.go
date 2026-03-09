package service

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
)

// Fetch fetches HTML content from a URL.
func (s *Service) Fetch(ctx context.Context, url string) ([]byte, content.FetcherType, error) {
	result, err := s.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, content.FetcherTypeGo, fmt.Errorf("failed to fetch url: %w", err)
	}

	return result.HTML, result.Type, nil
}

// Extract extracts article metadata and content from HTML bytes.
func (s *Service) Extract(ctx context.Context, htmlBytes []byte) (*model.Article, error) {
	article, err := s.extractor.GenerateFromHTML(ctx, htmlBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to extract: %w", err)
	}
	return article, nil
}

// GenerateEPUB generates an EPUB from an existing article.
func (s *Service) GenerateEPUB(article *model.Article) ([]byte, error) {
	epubData, err := s.publisher.GenerateEPUB(article)
	if err != nil {
		return nil, fmt.Errorf("failed to generate epub: %w", err)
	}
	return epubData, nil
}

// SendArticle sends an EPUB via email.
func (s *Service) SendArticle(
	ctx context.Context,
	destEmail string,
	epubBytes []byte,
	title string,
) (*email.SendEmailResponse, error) {
	req := &email.Request{
		EPUBData:  epubBytes,
		DestEmail: destEmail,
		Body:      consts.BuildCLIEmailBody(),
		Subject:   email.BuildSubject(title),
	}
	resp, err := s.sender.SendEmail(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}
	return resp, nil
}

// CreateArticle delegates to ArticleService.
func (s *Service) CreateArticle(ctx context.Context, rawURL, accountID string) (*model.Article, error) {
	article, err := s.articles.CreateArticle(ctx, rawURL, accountID)
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
	favoriteFilter *bool,
) (*servicetypes.GetArticlesResult, error) {
	result, err := s.articles.GetArticlesMetadata(ctx, accountID, page, pageSize, favoriteFilter)
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

// DeleteAllArticles delegates to ArticleService.
func (s *Service) DeleteAllArticles(
	ctx context.Context,
	accountID string,
) (*servicetypes.DeleteArticleResult, error) {
	result, err := s.articles.DeleteAllArticles(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete all articles: %w", err)
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

// CountSendsByAccountDateRange delegates to ArticleService.
func (s *Service) CountSendsByAccountDateRange(
	ctx context.Context,
	accountID string,
	startDate, endDate time.Time,
) (int, error) {
	count, err := s.articles.CountSendsByAccountDateRange(ctx, accountID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to count sends: %w", err)
	}
	return count, nil
}

// GetDBError delegates to ArticleService.
func (s *Service) GetDBError() error {
	err := s.articles.GetDBError()
	if err != nil {
		return fmt.Errorf("failed to get db error: %w", err)
	}
	return nil
}

// GetUserDeviceEmail delegates to UserProfileService.
func (s *Service) GetUserDeviceEmail(
	ctx context.Context,
	accountID string,
) (deviceEmail string, autoSend bool, err error) {
	deviceEmail, autoSend, err = s.profile.GetUserDeviceEmail(ctx, accountID)
	return
}

// SetUserDeviceEmail delegates to UserProfileService.
func (s *Service) SetUserDeviceEmail(ctx context.Context, accountID, deviceEmail string) error {
	if err := s.profile.SetUserDeviceEmail(ctx, accountID, deviceEmail); err != nil {
		return fmt.Errorf("failed to set user device email: %w", err)
	}
	return nil
}

// SetUserDeviceEmailWithAutoSend delegates to UserProfileService.
func (s *Service) SetUserDeviceEmailWithAutoSend(
	ctx context.Context,
	accountID, deviceEmail string,
	autoSend bool,
) error {
	if err := s.profile.SetUserDeviceEmailWithAutoSend(ctx, accountID, deviceEmail, autoSend); err != nil {
		return fmt.Errorf("failed to set user device email with auto send: %w", err)
	}
	return nil
}

// DeleteUserDeviceEmail delegates to UserProfileService.
func (s *Service) DeleteUserDeviceEmail(ctx context.Context, accountID string) error {
	if err := s.profile.DeleteUserDeviceEmail(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user device email: %w", err)
	}
	return nil
}

// GetUserProfile delegates to UserProfileService.
func (s *Service) GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error) {
	userProfile, err := s.profile.GetUserProfile(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return userProfile, nil
}

// SetUserEmail delegates to UserProfileService.
func (s *Service) SetUserEmail(ctx context.Context, accountID, userEmail string) error {
	if err := s.profile.SetUserEmail(ctx, accountID, userEmail); err != nil {
		return fmt.Errorf("failed to set user email: %w", err)
	}
	return nil
}

// DeleteUserProfile delegates to UserProfileService.
func (s *Service) DeleteUserProfile(ctx context.Context, accountID string) error {
	if err := s.profile.DeleteUserProfile(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}
	return nil
}

// HandleBounce delegates to UserProfileService.
func (s *Service) HandleBounce(ctx context.Context, deviceEmail, errorMessage string) error {
	if err := s.profile.HandleBounce(ctx, deviceEmail, errorMessage); err != nil {
		return fmt.Errorf("failed to handle bounce: %w", err)
	}
	return nil
}

// IsEmailBouncing delegates to UserProfileService.
func (s *Service) IsEmailBouncing(ctx context.Context, accountID, deviceEmail string) (bool, error) {
	bouncing, err := s.profile.IsEmailBouncing(ctx, accountID, deviceEmail)
	if err != nil {
		return false, fmt.Errorf("failed to check if email is bouncing: %w", err)
	}
	return bouncing, nil
}

// GetAccountIDByDeviceEmail delegates to UserProfileService.
func (s *Service) GetAccountIDByDeviceEmail(ctx context.Context, deviceEmail string) (string, error) {
	accountID, err := s.profile.GetAccountIDByDeviceEmail(ctx, deviceEmail)
	if err != nil {
		return "", fmt.Errorf("failed to get account id by device email: %w", err)
	}
	return accountID, nil
}

// SendArticleByID retrieves an article, generates an EPUB, and sends it to the user's device email.
func (s *Service) SendArticleByID(
	ctx context.Context,
	accountID,
	articleID string,
) (*servicetypes.SendArticleResult, error) {
	deviceEmail, _, err := s.GetUserDeviceEmail(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if deviceEmail == "" {
		return nil, fmt.Errorf("%w: user device email not configured", apperrors.ErrInvalid)
	}

	article, err := s.GetArticle(ctx, accountID, articleID)
	if err != nil {
		return nil, err
	}

	epubBytes, err := s.GenerateEPUB(article)
	if err != nil {
		return nil, err
	}

	emailResp, err := s.SendArticle(ctx, deviceEmail, epubBytes, article.Title)
	if err != nil {
		return nil, err
	}

	return &servicetypes.SendArticleResult{
		Article:     article,
		DeviceEmail: deviceEmail,
		EmailResp:   emailResp,
	}, nil
}
