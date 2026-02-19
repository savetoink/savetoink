// Package service provides the main orchestration logic for processing articles.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/content"
	"github.com/shaftoe/savetoink/internal/email"
	"github.com/shaftoe/savetoink/internal/email/mailjet"
	"github.com/shaftoe/savetoink/internal/epub"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/repository"
	"golang.org/x/sync/errgroup"
)

// Interface defines the contract for service operations.
type Interface interface {
	Process(ctx context.Context, url string) (*ProcessResult, error)
	Send(ctx context.Context, result *ProcessResult, subject, destEmail string) (*email.SendEmailResponse, error)
	WriteToFile(result *ProcessResult, outputPath string) error
	CreateArticle(ctx context.Context, rawURL, accountID string) (*CreateArticleResult, error)
	GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error)
	GetArticlesMetadata(ctx context.Context, accountID string, page, pageSize int) (*GetArticlesResult, error)
	DeleteArticle(ctx context.Context, accountID, articleID string) (*DeleteArticleResult, error)
	DeleteAllArticles(ctx context.Context, accountID string) (*DeleteArticleResult, error)
	GetDBError() error
	GetUserKindleEmail(ctx context.Context, accountID string) (string, error)
	SetUserKindleEmail(ctx context.Context, accountID, kindleEmail string) error
	DeleteUserProfile(ctx context.Context, accountID string) error
}

// Service holds the stateless dependencies and provides methods to process articles.
type Service struct {
	extractor       *content.Extractor
	generator       *epub.Generator
	sender          email.Sender
	repo            repository.ArticlesRepository
	userProfileRepo repository.UserProfileRepository
	cfg             *config.Config
	dbErrors        error
}

// New creates a new Service instance with the given config.
// All internal dependencies (extractor, generator, sender, repository) are created based on configuration.
// DynamoDB repository is wired only if both DynamoDBTable and AWSConfig are available.
func New(cfg *config.Config) *Service {
	var sender email.Sender
	if cfg.SendEnabled {
		switch cfg.EmailProvider {
		case consts.EmailBackendMailjet:
			sender = mailjet.NewSender(cfg.MailjetAPIKey, cfg.MailjetAPISecret, cfg.SenderEmail)
		default:
			sender = mailjet.NewSender(cfg.MailjetAPIKey, cfg.MailjetAPISecret, cfg.SenderEmail)
		}
	}

	var repo repository.ArticlesRepository
	var userProfileRepo repository.UserProfileRepository
	if cfg.ArticlesTable != "" && cfg.AWSConfig != nil {
		repo = repository.NewDynamoDB(cfg.AWSConfig, cfg.ArticlesTable, cfg.UserProfileTable)
		userProfileRepo = repo.(repository.UserProfileRepository)
	}

	return &Service{
		extractor:       content.NewExtractor(),
		generator:       epub.NewGenerator(),
		sender:          sender,
		repo:            repo,
		userProfileRepo: userProfileRepo,
		cfg:             cfg,
	}
}

// CreateArticleResult holds the result of creating an article.
type CreateArticleResult struct {
	Article   *model.Article
	Message   string
	EmailResp *email.SendEmailResponse
}

// GetArticlesResult holds the result of listing articles with pagination (without content).
type GetArticlesResult struct {
	Articles []*model.Article
	Page     int
	PageSize int
	Total    int
	HasMore  bool
}

// DeleteArticleResult holds the result of deleting an article.
type DeleteArticleResult struct {
	Deleted int
}

// ProcessResult holds the result of processing an article.
type ProcessResult struct {
	article  *model.Article
	epubData []byte
	url      string
}

// Article returns the extracted article.
func (r *ProcessResult) Article() *model.Article {
	return r.article
}

// EPUBData returns the generated EPUB data.
func (r *ProcessResult) EPUBData() []byte {
	return r.epubData
}

// URL returns the URL that was processed.
func (r *ProcessResult) URL() string {
	return r.url
}

// NewProcessResult creates a new ProcessResult for testing purposes.
// This is primarily used in tests to create mock results.
func NewProcessResult(article *model.Article, epubData []byte, url string) *ProcessResult {
	return &ProcessResult{
		article:  article,
		epubData: epubData,
		url:      url,
	}
}

// Process extracts content from a URL and generates EPUB data.
// Can be called multiple times to re-fetch fresh content.
func (s *Service) Process(ctx context.Context, url string) (*ProcessResult, error) {
	article, err := s.extractor.ExtractFromURL(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to extract article: %w", err)
	}

	if article.Title == "" {
		article.Title = "Untitled"
	}

	epubData, err := s.generator.Generate(article)
	if err != nil {
		return nil, fmt.Errorf("failed to generate EPUB: %w", err)
	}

	return &ProcessResult{
		article:  article,
		epubData: epubData,
		url:      url,
	}, nil
}

// Send sends an email with the processed article and EPUB.
// Returns an error if the result is nil or if sending fails.
// Can be called multiple times with the same result.
func (s *Service) Send(
	ctx context.Context,
	result *ProcessResult,
	subject, destEmail string,
) (*email.SendEmailResponse, error) {
	if result == nil {
		return nil, errors.New("result is nil, must call Process first")
	}

	if result.article == nil {
		return nil, errors.New("article is nil, must call Process first")
	}

	if s.sender == nil {
		return nil, errors.New("email sender is not configured")
	}

	if destEmail == "" {
		return nil, errors.New("destination email is required")
	}

	emailReq := &email.Request{
		Article:   result.article,
		EPUBData:  result.epubData,
		DestEmail: destEmail,
		Subject:   email.GenerateSubject(result.article.Title, subject),
	}

	resp, err := s.sender.SendEmail(ctx, emailReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return resp, nil
}

// WriteToFile writes the EPUB data to a file.
// Returns an error if the result is nil or if writing fails.
func (s *Service) WriteToFile(result *ProcessResult, outputPath string) error {
	if result == nil {
		return errors.New("result is nil, must call Process first")
	}

	if result.article == nil {
		return errors.New("article is nil, must call Process first")
	}

	if outputPath == "" {
		return errors.New("output path is empty")
	}

	err := s.generator.GenerateAndWrite(result.article, outputPath)
	if err != nil {
		return fmt.Errorf("failed to write EPUB document: %w", err)
	}

	return nil
}

// CreateArticle orchestrates the entire article creation flow:
// - cleans the URL and generates an article ID
// - processes the article (extracts content and generates EPUB)
// - optionally sends the article to Kindle via email
// - enriches the article with delivery metadata
// - stores the article to the database in the background (if repository is configured)
// Returns CreateArticleResult with the article and status information.
func (s *Service) CreateArticle(ctx context.Context, rawURL, accountID string) (*CreateArticleResult, error) {
	cleanURL, err := content.CleanURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to clean url: %w", err)
	}

	articleID, err := content.ArticleIDFromURL(cleanURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate article id: %w", err)
	}

	eg, articlesChan := s.startBackgroundDBStore(ctx)
	defer func() {
		close(articlesChan)
		_ = eg.Wait()
	}()

	article := &model.Article{
		Account:   accountID,
		ID:        articleID,
		URL:       cleanURL,
		CreatedAt: time.Now().UTC(),
	}
	articlesChan <- article

	result, err := s.Process(ctx, cleanURL)
	if err != nil {
		article.Error = err.Error()
		articlesChan <- article
		return nil, fmt.Errorf("failed to process article: %w", err)
	}

	if result.Article() == nil {
		articleErr := errors.New("failed to process article: article is nil")
		article.Error = articleErr.Error()
		articlesChan <- article
		return nil, articleErr
	}

	emailResp, destEmail, err := s.sendArticle(ctx, result, accountID)
	if err != nil {
		article.Error = err.Error()
		articlesChan <- article
		return nil, err
	}

	s.enrichArticle(result.Article(), &articleID, emailResp, accountID, destEmail)
	articlesChan <- result.Article()

	return &CreateArticleResult{
		Article:   result.Article(),
		Message:   s.getMessage(result.Article(), emailResp),
		EmailResp: emailResp,
	}, nil
}

func (s *Service) sendArticle(
	ctx context.Context,
	result *ProcessResult,
	accountID string,
) (*email.SendEmailResponse, string, error) {
	if !s.cfg.SendEnabled {
		return nil, "", nil
	}

	destEmail, err := s.GetUserKindleEmail(ctx, accountID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user kindle email: %w", err)
	}
	if destEmail == "" {
		return nil, "", errors.New("kindle email not configured for user")
	}

	emailResp, err := s.Send(ctx, result, "", destEmail)
	if err != nil {
		return nil, "", err
	}

	return emailResp, destEmail, nil
}

// GetDBError returns any accumulated database errors from background operations.
func (s *Service) GetDBError() error {
	return s.dbErrors
}

func (s *Service) startBackgroundDBStore(ctx context.Context) (eg *errgroup.Group, articles chan *model.Article) {
	eg, ctx = errgroup.WithContext(ctx)
	articles = make(chan *model.Article)
	var dbErrors error

	eg.Go(func() error {
		for article := range articles {
			if s.repo != nil {
				if storeErr := s.repo.Store(ctx, article); storeErr != nil {
					dbErrors = errors.Join(dbErrors, storeErr)
				}
			}
		}

		if dbErrors != nil {
			s.dbErrors = errors.Join(s.dbErrors, dbErrors)
		}

		return nil
	})

	return eg, articles
}

// DeleteArticle deletes a single article by account and ID.
func (s *Service) DeleteArticle(ctx context.Context, accountID, articleID string) (*DeleteArticleResult, error) {
	if articleID == "" {
		return nil, errors.New(consts.ErrInvalidArticleID)
	}

	if s.repo == nil {
		return &DeleteArticleResult{Deleted: 0}, nil
	}

	_, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return &DeleteArticleResult{Deleted: 0}, nil
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	err = s.repo.DeleteByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete article: %w", err)
	}

	return &DeleteArticleResult{Deleted: 1}, nil
}

// DeleteAllArticles deletes all articles for a given account.
func (s *Service) DeleteAllArticles(ctx context.Context, accountID string) (*DeleteArticleResult, error) {
	if s.repo == nil {
		return &DeleteArticleResult{Deleted: 0}, nil
	}

	deleted, err := s.repo.DeleteByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete all articles: %w", err)
	}

	return &DeleteArticleResult{Deleted: deleted}, nil
}

func (s *Service) enrichArticle(
	article *model.Article,
	id *string,
	emailResp *email.SendEmailResponse,
	accountID, destEmail string,
) {
	article.Account = accountID
	article.ID = *id

	if !s.cfg.SendEnabled {
		return
	}

	if emailResp == nil {
		article.DeliveryStatus = consts.StatusFailed
		return
	}

	article.DeliveryStatus = consts.StatusDelivered
	article.DeliveredFrom = &s.cfg.SenderEmail
	article.DeliveredTo = &destEmail
	article.DeliveredEmailUUID = &emailResp.EmailUUID
	article.DeliveredBy = s.cfg.EmailProvider
}

func (s *Service) getMessage(_ *model.Article, _ *email.SendEmailResponse) string {
	if !s.cfg.SendEnabled {
		return "article processed successfully (email sending disabled)"
	}
	return "article sent to Kindle successfully"
}

// GetArticle retrieves a single article by account ID and article ID.
// Returns the full article including all metadata and content.
func (s *Service) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	if articleID == "" {
		return nil, errors.New(consts.ErrInvalidArticleID)
	}

	if s.repo == nil {
		return nil, errors.New("repository not configured")
	}

	article, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("article not found")
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return article, nil
}

// GetArticlesMetadata retrieves article metadata for a given account with pagination.
// page starts at 1, pageSize limits the number of articles returned.
// Content field is excluded from returned articles.
func (s *Service) GetArticlesMetadata(
	ctx context.Context,
	accountID string,
	page, pageSize int,
) (*GetArticlesResult, error) {
	if s.repo == nil {
		return &GetArticlesResult{
			Articles: []*model.Article{},
			Page:     page,
			PageSize: pageSize,
			Total:    0,
			HasMore:  false,
		}, nil
	}

	articles, lastEvaluatedKey, total, err := s.repo.GetMetadataByAccount(ctx, accountID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}

	if articles == nil {
		articles = []*model.Article{}
	}

	return &GetArticlesResult{
		Articles: articles,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  lastEvaluatedKey != nil,
	}, nil
}

// GetUserKindleEmail retrieves the kindle email for a given account ID.
func (s *Service) GetUserKindleEmail(ctx context.Context, accountID string) (string, error) {
	if s.userProfileRepo == nil {
		return "", errors.New("user profile repository not configured")
	}

	profile, err := s.userProfileRepo.GetUserProfile(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to get user profile: %w", err)
	}

	if profile == nil {
		return "", nil
	}

	return profile.KindleEmail, nil
}

// SetUserKindleEmail sets the kindle email for a given account ID.
func (s *Service) SetUserKindleEmail(ctx context.Context, accountID, kindleEmail string) error {
	if s.userProfileRepo == nil {
		return errors.New("user profile repository not configured")
	}

	profile := &model.UserProfile{
		Account:     accountID,
		KindleEmail: kindleEmail,
	}

	if err := s.userProfileRepo.PutUserProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to set user profile: %w", err)
	}

	return nil
}

// DeleteUserProfile deletes the user profile for a given account ID.
func (s *Service) DeleteUserProfile(ctx context.Context, accountID string) error {
	if s.userProfileRepo == nil {
		return errors.New("user profile repository not configured")
	}

	if err := s.userProfileRepo.DeleteUserProfile(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	return nil
}
