// Package service provides article creation functionality.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/service/content"
	"golang.org/x/sync/errgroup"
)

// CreateArticle orchestrates the article creation flow:
// - cleans the URL and generates an article ID
// - processes the article (extracts content and generates EPUB)
// - stores the article to the database in the background (if repository is configured).
func (s *Service) CreateArticle(ctx context.Context, rawURL, accountID string) (*model.Article, error) {
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

	processedArticle := result.Article()
	processedArticle.Account = accountID
	processedArticle.ID = articleID
	articlesChan <- processedArticle

	return processedArticle, nil
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

func (s *Service) buildSendRecord(accountID, articleID, title, destEmail string) *model.Send {
	now := time.Now().UTC()
	return &model.Send{
		PK:          "USER#" + accountID,
		SK:          "SEND#" + now.Format(time.RFC3339) + "#" + articleID,
		Account:     accountID,
		ArticleID:   articleID,
		SentAt:      now,
		Title:       title,
		DestEmail:   destEmail,
		Status:      "pending",
		SenderEmail: s.cfg.SenderEmail,
		Provider:    string(s.cfg.EmailProvider),
	}
}

func (s *Service) updateSendRecord(ctx context.Context, send *model.Send, status, messageID, errorResponse string) {
	send.Status = status
	send.MessageID = messageID
	send.ErrorResponse = errorResponse
	if s.sendsRepo != nil {
		if err := s.sendsRepo.CreateSend(ctx, send); err != nil {
			slog.Warn("failed to update send record", "error", err)
		}
	}
}
