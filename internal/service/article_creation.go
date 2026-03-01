// Package service provides article creation functionality.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shaftoe/savetoink/internal/email"
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

// SendArticle sends an already-stored article to Kindle via email.
// Generates EPUB from the stored content and sends it to the user's Kindle email.
func (s *Service) SendArticle(
	ctx context.Context,
	article *model.Article,
	accountID string,
) (*email.SendEmailResponse, error) {
	if article == nil {
		return nil, errors.New("article is nil")
	}

	if article.Content == "" {
		return nil, errors.New("article has no content")
	}

	destEmail, _, err := s.GetUserDeviceEmail(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user device email: %w", err)
	}

	if destEmail == "" {
		return nil, errors.New("user email not configured")
	}

	epubData, err := s.generator.Generate(article)
	if err != nil {
		return nil, fmt.Errorf("failed to generate EPUB: %w", err)
	}

	result := NewProcessResult(article, epubData, article.URL)

	send := s.buildSendRecord(accountID, article.ID, article.Title, destEmail)
	if s.sendsRepo != nil {
		if storeErr := s.sendsRepo.CreateSend(ctx, send); storeErr != nil {
			return nil, fmt.Errorf("failed to store send record: %w", storeErr)
		}
	}

	emailResp, err := s.Send(ctx, result, destEmail)
	if err != nil {
		s.updateSendRecord(ctx, send, "failed", "", err.Error())
		return nil, err
	}

	s.updateSendRecord(ctx, send, "success", emailResp.MessageID, "")

	if s.repo != nil {
		if storeErr := s.repo.Store(ctx, article); storeErr != nil {
			slog.Warn("failed to update article in database", "error", storeErr)
		}
	}

	return emailResp, nil
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
