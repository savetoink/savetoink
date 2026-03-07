// Package articles provides article management and processing services.
package articles

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shaftoe/savetoink/backend/internal/config"
	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/shaftoe/savetoink/backend/internal/email"
	"github.com/shaftoe/savetoink/backend/internal/logging"
	"github.com/shaftoe/savetoink/backend/internal/model"
	"github.com/shaftoe/savetoink/backend/internal/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/internal/service/content"
	"github.com/shaftoe/savetoink/backend/internal/service/processing"
	"github.com/shaftoe/savetoink/backend/internal/service/profile"
	"github.com/shaftoe/savetoink/backend/internal/service/servicetypes"
	"golang.org/x/sync/errgroup"
)

// ArticleService handles article CRUD operations and sending.
type ArticleService struct {
	repo        repository.ArticlesRepository
	sendsRepo   repository.SendsRepository
	processor   *processing.ArticleProcessingService
	userProfile *profile.UserProfileService
	cfg         *config.Config
	sender      email.Sender
	dbErrors    error
}

// New creates a new ArticleService instance.
func New(
	repo repository.ArticlesRepository,
	sendsRepo repository.SendsRepository,
	processor *processing.ArticleProcessingService,
	userProfile *profile.UserProfileService,
	cfg *config.Config,
	sender email.Sender,
) *ArticleService {
	return &ArticleService{
		repo:        repo,
		sendsRepo:   sendsRepo,
		processor:   processor,
		userProfile: userProfile,
		cfg:         cfg,
		sender:      sender,
	}
}

// CreateArticle processes a URL and creates an article entry.
func (s *ArticleService) CreateArticle(ctx context.Context, rawURL, accountID string) (*model.Article, error) {
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

	result, err := s.processor.Process(ctx, cleanURL)
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

	processedArticle := result.Article()
	processedArticle.Account = accountID
	processedArticle.ID = articleID
	articlesChan <- processedArticle

	return processedArticle, nil
}

// GetArticle retrieves an article by account ID and article ID.
func (s *ArticleService) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	if articleID == "" {
		return nil, errors.New(consts.ErrInvalidArticleID)
	}

	if s.repo == nil {
		return nil, errors.New("repository not configured")
	}

	article, err := s.repo.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, repoimpl.ErrNotFound) {
			return nil, errors.New("article not found")
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
		return nil, errors.New(consts.ErrInvalidArticleID)
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
			return false, errors.New("article not found")
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

// SendArticle sends an article to the user's device email.
// Routes to mode-specific implementation based on cliMode flag.
func (s *ArticleService) SendArticle(
	ctx context.Context,
	article *model.Article,
	accountID string,
	destEmail string,
	cliMode bool,
) (*email.SendEmailResponse, error) {
	if err := s.validateArticleForSend(article); err != nil {
		return nil, err
	}

	if cliMode {
		return s.sendArticleCLIMode(ctx, article, destEmail)
	}
	return s.sendArticleServerMode(ctx, article, accountID)
}

func (s *ArticleService) sendArticleCLIMode(
	ctx context.Context,
	article *model.Article,
	destEmail string,
) (*email.SendEmailResponse, error) {
	if destEmail == "" {
		return nil, errors.New("destination email is required in CLI mode")
	}

	result, processErr := s.processArticleToResult(article)
	if processErr != nil {
		return nil, processErr
	}

	emailReq := &email.Request{
		Article:   result.Article(),
		EPUBData:  result.EPUBData(),
		DestEmail: destEmail,
		Body:      consts.BuildCLIEmailBody(),
	}

	emailResp, sendErr := s.sender.SendEmail(ctx, emailReq)
	if sendErr != nil {
		return nil, fmt.Errorf("failed to send email: %w", sendErr)
	}

	return emailResp, nil
}

func (s *ArticleService) sendArticleServerMode(
	ctx context.Context,
	article *model.Article,
	accountID string,
) (*email.SendEmailResponse, error) {
	destEmail, _, getErr := s.userProfile.GetUserDeviceEmail(ctx, accountID)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get user device email: %w", getErr)
	}
	if destEmail == "" {
		return nil, errors.New("user email not configured")
	}

	result, processErr := s.processArticleToResult(article)
	if processErr != nil {
		return nil, processErr
	}

	emailReq := &email.Request{
		Article:   result.Article(),
		EPUBData:  result.EPUBData(),
		DestEmail: destEmail,
		Body:      consts.BuildEmailBody(s.cfg.AppURL),
	}

	emailResp, sendErr, dbError := s.sendEmailAndCreateRecord(ctx, emailReq, accountID, article, destEmail)
	if dbError != nil {
		s.dbErrors = errors.Join(s.dbErrors, dbError)
	}

	if sendErr != nil {
		dbError = s.updateSendRecordOnFailure(ctx, accountID, article.ID, sendErr)
		if dbError != nil {
			s.dbErrors = errors.Join(s.dbErrors, dbError)
		}
		return nil, sendErr
	}

	dbError = s.updateSendRecordAndArticleOnSuccess(ctx, accountID, article, emailResp.MessageID)
	if dbError != nil {
		s.dbErrors = errors.Join(s.dbErrors, dbError)
	}

	return emailResp, nil
}

func (s *ArticleService) sendEmailAndCreateRecord(
	ctx context.Context,
	emailReq *email.Request,
	accountID string,
	article *model.Article,
	destEmail string,
) (*email.SendEmailResponse, error, error) {
	eg, _ := errgroup.WithContext(ctx)
	var emailResp *email.SendEmailResponse
	var sendErr error
	var dbError error

	eg.Go(func() error {
		var err error
		emailResp, err = s.sender.SendEmail(ctx, emailReq)
		if err != nil {
			sendErr = fmt.Errorf("failed to send email: %w", err)
		} else {
			logging.AddLogAttr(ctx, slog.String("email_message_id", emailResp.MessageID))
			logging.AddLogAttr(ctx, slog.String("email_destination", destEmail))
		}
		return nil
	})

	eg.Go(func() error {
		if s.sendsRepo == nil {
			return nil
		}
		createErr := s.sendsRepo.CreateSendRecord(
			ctx,
			accountID,
			article.ID,
			article.Title,
			destEmail,
		)
		if createErr != nil {
			dbError = fmt.Errorf("failed to create send record: %w", createErr)
		}
		return nil
	})

	_ = eg.Wait()
	return emailResp, sendErr, dbError
}

func (s *ArticleService) updateSendRecordOnFailure(
	ctx context.Context,
	accountID string,
	articleID string,
	sendErr error,
) error {
	if s.sendsRepo == nil {
		return nil
	}
	updateErr := s.sendsRepo.UpdateSendRecord(
		ctx,
		accountID,
		articleID,
		"failed",
		"",
		sendErr.Error(),
	)
	if updateErr != nil {
		return fmt.Errorf("failed to update send record: %w", updateErr)
	}
	return nil
}

func (s *ArticleService) updateSendRecordAndArticleOnSuccess(
	ctx context.Context,
	accountID string,
	article *model.Article,
	messageID string,
) error {
	eg, _ := errgroup.WithContext(ctx)
	var dbError error

	eg.Go(func() error {
		if s.sendsRepo == nil {
			return nil
		}
		updateErr := s.sendsRepo.UpdateSendRecord(
			ctx,
			accountID,
			article.ID,
			"success",
			messageID,
			"",
		)
		if updateErr != nil {
			dbError = fmt.Errorf("failed to update send record: %w", updateErr)
		}
		return nil
	})

	eg.Go(func() error {
		if s.repo == nil {
			return nil
		}
		storeErr := s.repo.Store(ctx, article)
		if storeErr != nil {
			dbError = errors.Join(dbError, fmt.Errorf("failed to store article: %w", storeErr))
		}
		return nil
	})

	_ = eg.Wait()
	return dbError
}

func (s *ArticleService) processArticleToResult(article *model.Article) (*servicetypes.ProcessResult, error) {
	epubData, err := s.processor.Generator.Generate(article)
	if err != nil {
		return nil, fmt.Errorf("failed to generate EPUB: %w", err)
	}
	return servicetypes.NewProcessResult(article, epubData, article.URL), nil
}

// CountSendsByAccountDateRange counts the number of sends within a date range.
func (s *ArticleService) CountSendsByAccountDateRange(
	ctx context.Context,
	accountID string,
	startDate, endDate time.Time,
) (int, error) {
	if s.sendsRepo == nil {
		return 0, nil
	}

	count, err := s.sendsRepo.CountSendsByAccountDateRange(ctx, accountID, startDate, endDate)
	if err != nil {
		return 0, errors.New("failed to count sends")
	}

	return count, nil
}

// GetDBError returns any database errors that occurred during background operations.
func (s *ArticleService) GetDBError() error {
	return s.dbErrors
}

func (s *ArticleService) startBackgroundDBStore(
	ctx context.Context,
) (eg *errgroup.Group, articles chan *model.Article) {
	eg, groupCtx := errgroup.WithContext(ctx)
	articles = make(chan *model.Article)
	var dbErrors error

	eg.Go(func() error {
		for article := range articles {
			if s.repo != nil {
				if storeErr := s.repo.Store(groupCtx, article); storeErr != nil {
					dbErrors = errors.Join(dbErrors, storeErr)
				}
			}
		}

		if dbErrors != nil {
			s.dbErrors = errors.Join(s.dbErrors, dbErrors)
		}

		return nil
	})

	return
}

func (s *ArticleService) validateArticleForSend(article *model.Article) error {
	if article == nil {
		return errors.New("article is nil")
	}
	if article.Content == "" {
		return errors.New("article has no content")
	}
	return nil
}
