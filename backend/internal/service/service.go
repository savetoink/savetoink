// Package service provides main orchestration logic for processing articles.
package service

import (
	"context"
	"time"

	"github.com/shaftoe/savetoink/backend/internal/config"
	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/shaftoe/savetoink/backend/internal/email"
	"github.com/shaftoe/savetoink/backend/internal/email/mailjet"
	"github.com/shaftoe/savetoink/backend/internal/model"
	"github.com/shaftoe/savetoink/backend/internal/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/internal/service/content"
	"github.com/shaftoe/savetoink/backend/internal/service/epub"
)

// Interface defines the contract for service operations.
type Interface interface {
	Process(ctx context.Context, url string) (*ProcessResult, error)
	Send(
		ctx context.Context,
		result *ProcessResult,
		destEmail string,
	) (*email.SendEmailResponse, error)
	SendArticle(
		ctx context.Context,
		article *model.Article,
		accountID string,
	) (*email.SendEmailResponse, error)
	WriteToFile(result *ProcessResult, outputPath string) error
	CreateArticle(ctx context.Context, rawURL, accountID string) (*model.Article, error)
	GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error)
	GetArticlesMetadata(
		ctx context.Context, accountID string, page, pageSize int, favoriteFilter *bool,
	) (*GetArticlesResult, error)
	DeleteArticle(ctx context.Context, accountID, articleID string) (*DeleteArticleResult, error)
	DeleteAllArticles(ctx context.Context, accountID string) (*DeleteArticleResult, error)
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

// Service holds the stateless dependencies and provides methods to process articles.
type Service struct {
	extractor       *content.Extractor
	generator       *epub.Generator
	sender          email.Sender
	repo            repository.ArticlesRepository
	userProfileRepo repository.UserProfileRepository
	sendsRepo       repository.SendsRepository
	cfg             *config.Config
	dbErrors        error
}

// New creates a new Service instance with the provided dependencies.
func New(deps *Dependencies) *Service {
	return &Service{
		extractor:       deps.Extractor,
		generator:       deps.Generator,
		sender:          deps.Sender,
		repo:            deps.ArticlesRepo,
		userProfileRepo: deps.UserProfileRepo,
		sendsRepo:       deps.SendsRepo,
		cfg:             deps.Config,
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
