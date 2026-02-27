// Package service provides article creation functionality.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/email"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/service/content"
	"golang.org/x/sync/errgroup"
)

// CreateArticleResult holds the result of creating an article.
type CreateArticleResult struct {
	Article   *model.Article
	Message   string
	EmailResp *email.SendEmailResponse
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
	destEmail, err := s.GetUserKindleEmail(ctx, accountID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user kindle email: %w", err)
	}
	if destEmail == "" {
		return nil, "", nil
	}

	emailResp, err := s.Send(ctx, result, destEmail)
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

func (s *Service) enrichArticle(
	article *model.Article,
	id *string,
	emailResp *email.SendEmailResponse,
	accountID, destEmail string,
) {
	article.Account = accountID
	article.ID = *id

	if emailResp == nil {
		if destEmail != "" {
			article.DeliveryStatus = consts.StatusFailed
		}
		return
	}

	article.DeliveryStatus = consts.StatusDelivered
	article.DeliveredFrom = &s.cfg.SenderEmail
	article.DeliveredTo = &destEmail
	article.DeliveredEmailUUID = &emailResp.EmailUUID
	article.DeliveredBy = s.cfg.EmailProvider
}

func (s *Service) getMessage(_ *model.Article, emailResp *email.SendEmailResponse) string {
	if emailResp == nil {
		return "article saved (kindle email not configured)"
	}
	return "article sent to Kindle successfully"
}
