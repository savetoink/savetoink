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
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/internal/email/mailjet"
	"github.com/shaftoe/savetoink/backend/lib/internal/epub"
	"github.com/shaftoe/savetoink/backend/lib/internal/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/internal/repository/dynamodb"
	repoimplsqlite "github.com/shaftoe/savetoink/backend/lib/internal/repository/sqlite"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/articles"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/profile"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	"github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"golang.org/x/net/html"
)

// Interface defines the contract for service operations.
type Interface interface {
	// Fetch fetches the HTML content of a given URL and returns the fetched content with metadata.
	Fetch(ctx context.Context, u *url.URL) (*content.FetchedContent, error)

	// ParseHTML parses HTML content from fetched content into a DOM node.
	ParseHTML(ctx context.Context, fetched *content.FetchedContent) (*html.Node, error)

	// Clean extracts article content from a DOM node.
	Clean(ctx context.Context, doc *html.Node, u *url.URL) (*model.Article, error)

	// GenerateEPUB generates an EPUB document from an article.
	GenerateEPUB(article *model.Article) (io.ReadCloser, error)

	// ReadEPUB reads an EPUB file from a URL and returns the file reader and title.
	ReadEPUB(ctx context.Context, u *url.URL) (io.ReadCloser, string, error)

	///////////
	// Articles
	///////////

	// CreateArticle stores a new partial metadata article from a given URL and account ID in the database.
	CreateArticle(ctx context.Context, u *url.URL, accountID string) (*model.Article, error)

	// UpdateArticle updates an existing article with full content and metadata.
	UpdateArticle(ctx context.Context, article *model.Article) error

	// GetArticle retrieves an existing article by account ID and article ID from the database.
	GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error)

	// GetArticlesMetadata retrieves all articles metadata by account ID and pagination parameters.
	GetArticlesMetadata(
		ctx context.Context, accountID string, page, pageSize int, filter *types.ArticleFilter,
	) (*servicetypes.GetArticlesResult, error)

	// DeleteArticle deletes an existing article by account ID and article ID from the database.
	DeleteArticle(ctx context.Context, accountID, articleID string) (*servicetypes.DeleteArticleResult, error)

	// ToggleFavorite toggles the favorite status of an article.
	ToggleFavorite(ctx context.Context, accountID string, articleID string) (bool, error)

	// SendArticle sends an EPUB document as attachment to a given email address.
	SendArticle(
		ctx context.Context,
		destEmail string,
		epubData io.ReadCloser,
		title string,
	) (*email.SendEmailResponse, error)

	// SendArticleByID retrieves an article by account ID and article ID,
	// generates an EPUB, and sends it to the user's device email.
	SendArticleByID(ctx context.Context, accountID, articleID string) (*servicetypes.SendArticleResult, error)

	///////////////
	// User profile
	///////////////

	// GetUserProfile retrieves the user profile for a given account.
	GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error)

	// SetUserEmail sets the user email for a given account.
	SetUserEmail(ctx context.Context, accountID, email string) error

	// DeleteUserProfile deletes the user profile for a given account.
	DeleteUserProfile(ctx context.Context, accountID string) error

	// GetUserDeviceEmailAndAutoSend retrieves the device email and auto-send preference for a given account.
	GetUserDeviceEmailAndAutoSend(ctx context.Context, accountID string) (string, bool, error)

	// SetUserDeviceEmailWithAutoSend sets the device email and auto-send preference for a given account.
	SetUserDeviceEmailWithAutoSend(ctx context.Context, accountID, deviceEmail string, autoSend bool) error

	// DeleteUserDeviceEmail deletes the device email for a given account.
	DeleteUserDeviceEmail(ctx context.Context, accountID string) error

	//////////
	// Mailing
	//////////

	// HandleBounce handles a bounce notification for a given device email.
	HandleBounce(ctx context.Context, deviceEmail, errorMessage string) error

	// IsEmailBouncing checks if an email address is currently bouncing.
	IsEmailBouncing(ctx context.Context, accountID, deviceEmail string) (bool, error)

	// GetAccountIDByDeviceEmail retrieves the account ID associated with a device email.
	GetAccountIDByDeviceEmail(ctx context.Context, deviceEmail string) (string, error)

	////////
	// Sends
	////////

	// CountSendsByAccountDateRange counts the number of sends for a given account within a date range.
	CountSendsByAccountDateRange(ctx context.Context, accountID string, startDate, endDate time.Time) (int, error)
}

// Dependencies holds all external dependencies required by Service.
type Dependencies struct {
	Fetcher         *content.Fetcher
	Extractor       content.Extractor
	Cleaner         content.Cleaner
	Publisher       *epub.Publisher
	Reader          *epub.Reader
	Sender          email.Sender
	ArticlesRepo    repository.ArticlesRepository
	UserProfileRepo repository.UserProfileRepository
	SendsRepo       repository.SendsRepository
	Config          *config.Config
}

// Service orchestrator composes sub-services and implements the Interface.
type Service struct {
	fetcher   *content.Fetcher
	extractor content.Extractor
	cleaner   content.Cleaner
	publisher *epub.Publisher
	reader    *epub.Reader
	articles  *articles.ArticleService
	profile   *profile.UserProfileService
	sender    email.Sender
	sendsRepo repository.SendsRepository
	cfg       *config.Config
}

// New creates a Service instance with the provided dependencies.
func New(deps *Dependencies) *Service {
	fetcher := deps.Fetcher
	publisher := deps.Publisher
	userProfile := profile.New(deps.UserProfileRepo)
	articleSvc := articles.New(
		deps.ArticlesRepo,
		publisher,
		userProfile,
	)

	return &Service{
		fetcher:   fetcher,
		extractor: deps.Extractor,
		cleaner:   deps.Cleaner,
		publisher: publisher,
		reader:    deps.Reader,
		articles:  articleSvc,
		profile:   userProfile,
		sender:    deps.Sender,
		sendsRepo: deps.SendsRepo,
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

	// disable repositories in CLI mode
	if cfg.Mode == consts.ModeCLI {
		return Dependencies{
			Fetcher:         content.NewFetcher(cfg.BrowserlessKey),
			Extractor:       content.NewDOMExtractor(),
			Cleaner:         content.NewTrafilaturaCleaner(),
			Publisher:       epub.NewPublisher(epub.WithMemoryStorage()),
			Reader:          epub.NewReader(),
			Sender:          sender,
			ArticlesRepo:    articlesRepo,
			UserProfileRepo: userProfileRepo,
			SendsRepo:       sendsRepo,
			Config:          cfg,
		}
	}

	switch cfg.StorageBackend {
	case consts.StorageBackendDynamoDB, "":
		if cfg.AWSConfig != nil {
			dynamoDB := repoimpl.NewDynamoDB(cfg.AWSConfig, cfg.ArticlesTable, cfg.UserProfileTable, cfg.SendsTable)
			articlesRepo = dynamoDB
			userProfileRepo = dynamoDB
			sendsRepo = dynamoDB
		}
	case consts.StorageBackendSQLite:
		ctx, cancel := context.WithTimeout(context.Background(), consts.SqliteInitTimeout)
		defer cancel()
		sqlite, err := repoimplsqlite.NewSQLite(ctx, cfg.SQLitePath)
		if err != nil {
			panic(fmt.Sprintf("failed to create SQLite repository: %v", err))
		}
		articlesRepo = sqlite
		userProfileRepo = sqlite
		sendsRepo = sqlite
	}

	return Dependencies{
		Fetcher:         content.NewFetcher(cfg.BrowserlessKey),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Reader:          epub.NewReader(),
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
