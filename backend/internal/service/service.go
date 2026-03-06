// Package service provides main orchestration logic for processing articles.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shaftoe/savetoink/backend/internal/config"
	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/shaftoe/savetoink/backend/internal/email"
	"github.com/shaftoe/savetoink/backend/internal/email/mailjet"
	"github.com/shaftoe/savetoink/backend/internal/model"
	"github.com/shaftoe/savetoink/backend/internal/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/internal/service/articles"
	"github.com/shaftoe/savetoink/backend/internal/service/content"
	"github.com/shaftoe/savetoink/backend/internal/service/epub"
	"github.com/shaftoe/savetoink/backend/internal/service/processing"
	"github.com/shaftoe/savetoink/backend/internal/service/profile"
	"github.com/shaftoe/savetoink/backend/internal/service/servicetypes"
)

// Interface defines the contract for service operations.
type Interface interface {
	Process(ctx context.Context, url string) (*servicetypes.ProcessResult, error)
	Send(
		ctx context.Context,
		result *servicetypes.ProcessResult,
		destEmail string,
	) (*email.SendEmailResponse, error)
	SendArticle(
		ctx context.Context,
		article *model.Article,
		accountID string,
	) (*email.SendEmailResponse, error)
	WriteToFile(result *servicetypes.ProcessResult, outputPath string) error
	CreateArticle(ctx context.Context, rawURL, accountID string) (*model.Article, error)
	GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error)
	GetArticlesMetadata(
		ctx context.Context, accountID string, page, pageSize int, favoriteFilter *bool,
	) (*servicetypes.GetArticlesResult, error)
	DeleteArticle(ctx context.Context, accountID, articleID string) (*servicetypes.DeleteArticleResult, error)
	DeleteAllArticles(ctx context.Context, accountID string) (*servicetypes.DeleteArticleResult, error)
	GetDBError() error
	GetUserDeviceEmail(ctx context.Context, accountID string) (string, bool, error)
	SetUserDeviceEmailWithAutoSend(ctx context.Context, accountID, deviceEmail string, autoSend bool) error
	DeleteUserDeviceEmail(ctx context.Context, accountID string) error
	GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error)
	SetUserEmail(ctx context.Context, accountID, email string) error
	DeleteUserProfile(ctx context.Context, accountID string) error
	ToggleFavorite(ctx context.Context, accountID, articleID string) (bool, error)
	CountSendsByAccountDateRange(ctx context.Context, accountID string, startDate, endDate time.Time) (int, error)
	HandleBounce(ctx context.Context, deviceEmail, errorMessage string) error
	IsEmailBouncing(ctx context.Context, accountID, deviceEmail string) (bool, error)
	GetAccountIDByDeviceEmail(ctx context.Context, deviceEmail string) (string, error)
}

// Dependencies holds all external dependencies required by Service.
type Dependencies struct {
	Extractor       *content.Extractor
	Generator       *epub.Generator
	Sender          email.Sender
	ArticlesRepo    repository.ArticlesRepository
	UserProfileRepo repository.UserProfileRepository
	SendsRepo       repository.SendsRepository
	Config          *config.Config
}

// Service orchestrator composes sub-services and implements the Interface.
type Service struct {
	processor *processing.ArticleProcessingService
	articles  *articles.ArticleService
	profile   *profile.UserProfileService
	sender    email.Sender
	cfg       *config.Config
}

// New creates a Service instance with the provided dependencies.
func New(deps *Dependencies) *Service {
	processor := processing.New(deps.Extractor, deps.Generator)
	userProfile := profile.New(deps.UserProfileRepo)
	articleSvc := articles.New(
		deps.ArticlesRepo,
		deps.SendsRepo,
		processor,
		userProfile,
		deps.Config,
		deps.Sender,
	)

	return &Service{
		processor: processor,
		articles:  articleSvc,
		profile:   userProfile,
		sender:    deps.Sender,
		cfg:       deps.Config,
	}
}

// NewDependenciesFromConfig creates all dependencies from configuration.
func NewDependenciesFromConfig(cfg *config.Config) Dependencies {
	var sender email.Sender
	if cfg.EmailProvider == consts.EmailBackendMailjet {
		sender = mailjet.NewSender(cfg.MailjetAPIKey, cfg.MailjetAPISecret, cfg.SenderEmail)
	}

	var articlesRepo repository.ArticlesRepository
	var userProfileRepo repository.UserProfileRepository
	var sendsRepo repository.SendsRepository
	if cfg.AWSConfig != nil {
		dynamoDB := repoimpl.NewDynamoDB(cfg.AWSConfig, cfg.ArticlesTable, cfg.UserProfileTable, cfg.SendsTable)
		articlesRepo = dynamoDB
		userProfileRepo = dynamoDB
		sendsRepo = dynamoDB
	}

	return Dependencies{
		Extractor:       content.NewExtractor(),
		Generator:       epub.NewGenerator(),
		Sender:          sender,
		ArticlesRepo:    articlesRepo,
		UserProfileRepo: userProfileRepo,
		SendsRepo:       sendsRepo,
		Config:          cfg,
	}
}

// NewFromConfig creates a Service and all dependencies from configuration.
func NewFromConfig(cfg *config.Config) *Service {
	deps := NewDependenciesFromConfig(cfg)
	return New(&deps)
}

// Delegate methods to appropriate sub-services

// Process delegates to ArticleProcessingService.
func (s *Service) Process(ctx context.Context, url string) (*servicetypes.ProcessResult, error) {
	result, err := s.processor.Process(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to process article: %w", err)
	}
	return result, nil
}

// WriteToFile delegates to ArticleProcessingService.
func (s *Service) WriteToFile(result *servicetypes.ProcessResult, outputPath string) error {
	if err := s.processor.WriteToFile(result, outputPath); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

// Send sends an email with processed article and EPUB.
func (s *Service) Send(
	ctx context.Context,
	result *servicetypes.ProcessResult,
	destEmail string,
) (*email.SendEmailResponse, error) {
	if result == nil {
		return nil, errors.New("result is nil, must call Process first")
	}
	if result.Article() == nil {
		return nil, errors.New("article is nil, must call Process first")
	}
	if s.sender == nil {
		return nil, errors.New("email sender is not configured")
	}
	if destEmail == "" {
		return nil, errors.New("destination email is required")
	}

	emailReq := &email.Request{
		Article:   result.Article(),
		EPUBData:  result.EPUBData(),
		DestEmail: destEmail,
		AppURL:    s.cfg.AppURL,
	}

	resp, err := s.sender.SendEmail(ctx, emailReq)
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

// SendArticle delegates to ArticleService.
func (s *Service) SendArticle(
	ctx context.Context,
	article *model.Article,
	accountID string,
) (*email.SendEmailResponse, error) {
	resp, err := s.articles.SendArticle(ctx, article, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to send article: %w", err)
	}
	return resp, nil
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
