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
	Process(ctx context.Context, url string) (*servicetypes.ProcessResult, error)
	SendArticle(
		ctx context.Context,
		article *model.Article,
		accountID string,
		destEmail string,
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
	SetUserEmail(ctx context.Context, accountID string, email string) error
	DeleteUserProfile(ctx context.Context, accountID string) error
	ToggleFavorite(ctx context.Context, accountID string, articleID string) (bool, error)
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
