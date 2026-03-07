// Package service provides main entry point to interact with Save to Ink various services e.g.:
// - content extraction
// - articles
// - profiles
// - email sending
// - storage access
// - etc...
package service

import (
	"context"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/email/mailjet"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/service/articles"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/epub"
	"github.com/shaftoe/savetoink/backend/lib/service/processing"
	"github.com/shaftoe/savetoink/backend/lib/service/profile"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
)

// Interface defines the contract for service operations.
type Interface interface {
	// Fetch fetches the HTML content of a given URL.
	Fetch(ctx context.Context, url string) ([]byte, error)

	// Extract extracts the clean article content from HTML bytes.
	Extract(ctx context.Context, htmlBytes []byte) ([]byte, error)

	// GenerateEPUB generates an EPUB document from an article.
	GenerateEPUB(article *model.Article) ([]byte, error)

	// SendArticle sends an EPUB document as attechment to a given email address.
	SendArticle(
		ctx context.Context,
		destEmail string,
		epubBytes []byte,
	) (*email.SendEmailResponse, error)

	// CreateArticle stores a new article from a given URL and account ID in the database.
	CreateArticle(ctx context.Context, rawURL, accountID string) (*model.Article, error)

	// GetArticle retrieves an article by account ID and article ID from the database.
	GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error)

	// GetArticlesMetadata retrieves all articles metadata by account ID and pagination parameters.
	GetArticlesMetadata(
		ctx context.Context, accountID string, page, pageSize int, favoriteFilter *bool,
	) (*servicetypes.GetArticlesResult, error)

	// DeleteArticle deletes an article by account ID and article ID from the database.
	DeleteArticle(ctx context.Context, accountID, articleID string) (*servicetypes.DeleteArticleResult, error)

	// DeleteAllArticles deletes all articles by account ID from the database.
	DeleteAllArticles(ctx context.Context, accountID string) (*servicetypes.DeleteArticleResult, error)

	// GetDBError returns the database error.
	GetDBError() error

	// GetUserDeviceEmail retrieves the device email and auto-send preference for a given account.
	GetUserDeviceEmail(ctx context.Context, accountID string) (string, bool, error)

	// SetUserDeviceEmailWithAutoSend sets the device email and auto-send preference for a given account.
	SetUserDeviceEmailWithAutoSend(ctx context.Context, accountID, deviceEmail string, autoSend bool) error

	// DeleteUserDeviceEmail deletes the device email for a given account.
	DeleteUserDeviceEmail(ctx context.Context, accountID string) error

	// GetUserProfile retrieves the user profile for a given account.
	GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error)

	// SetUserEmail sets the user email for a given account.
	SetUserEmail(ctx context.Context, accountID, email string) error

	// DeleteUserProfile deletes the user profile for a given account.
	DeleteUserProfile(ctx context.Context, accountID string) error

	// ToggleFavorite toggles the favorite status of an article.
	ToggleFavorite(ctx context.Context, accountID string, articleID string) (bool, error)

	// CountSendsByAccountDateRange counts the number of sends for a given account within a date range.
	CountSendsByAccountDateRange(ctx context.Context, accountID string, startDate, endDate time.Time) (int, error)

	// HandleBounce handles a bounce notification for a given device email.
	HandleBounce(ctx context.Context, deviceEmail, errorMessage string) error

	// IsEmailBouncing checks if an email address is currently bouncing.
	IsEmailBouncing(ctx context.Context, accountID, deviceEmail string) (bool, error)

	// GetAccountIDByDeviceEmail retrieves the account ID associated with a device email.
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
