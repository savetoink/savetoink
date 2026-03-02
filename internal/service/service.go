// Package service provides main orchestration logic for processing articles.
package service

import (
	"context"
	"time"

	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/email"
	"github.com/shaftoe/savetoink/internal/email/mailjet"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/repository"
	repoimpl "github.com/shaftoe/savetoink/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/internal/service/content"
	"github.com/shaftoe/savetoink/internal/service/epub"
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

// New creates a new Service instance with the given config.
// All internal dependencies (extractor, generator, sender, repository) are created based on configuration.
func New(cfg *config.Config) *Service {
	var sender email.Sender
	if cfg.MailjetAPIKey != "" && cfg.MailjetAPISecret != "" && cfg.SenderEmail != "" {
		cfg.EmailProvider = consts.EmailBackendMailjet
		sender = mailjet.NewSender(cfg.MailjetAPIKey, cfg.MailjetAPISecret, cfg.SenderEmail)
	}

	var repo repository.ArticlesRepository
	var userProfileRepo repository.UserProfileRepository
	var sendsRepo repository.SendsRepository
	if cfg.AWSConfig != nil {
		repo = repoimpl.NewDynamoDB(cfg.AWSConfig, cfg.ArticlesTable, cfg.UserProfileTable, cfg.SendsTable)
		userProfileRepo = repo.(repository.UserProfileRepository)
		sendsRepo = repo.(repository.SendsRepository)
	}

	return &Service{
		extractor:       content.NewExtractor(),
		generator:       epub.NewGenerator(),
		sender:          sender,
		repo:            repo,
		userProfileRepo: userProfileRepo,
		sendsRepo:       sendsRepo,
		cfg:             cfg,
	}
}
