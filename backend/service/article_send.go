package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shaftoe/savetoink/backend/email"
	"github.com/shaftoe/savetoink/backend/model"
)

// SendArticle sends an already-stored article to devices via email.
// Generates EPUB from the stored content and sends it to the user's device email.
func (s *Service) SendArticle(
	ctx context.Context,
	article *model.Article,
	accountID string,
) (*email.SendEmailResponse, error) {
	if err := s.validateArticleForSend(article); err != nil {
		return nil, err
	}

	destEmail, _, getErr := s.GetUserDeviceEmail(ctx, accountID)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get user device email: %w", getErr)
	}

	if destEmail == "" {
		return nil, errors.New("user email not configured")
	}

	epubData, epubErr := s.generator.Generate(article)
	if epubErr != nil {
		return nil, fmt.Errorf("failed to generate EPUB: %w", epubErr)
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

func (s *Service) validateArticleForSend(article *model.Article) error {
	if article == nil {
		return errors.New("article is nil")
	}
	if article.Content == "" {
		return errors.New("article has no content")
	}
	return nil
}

// CountSendsByAccountDateRange counts the number of sends for a given account within a date range.
func (s *Service) CountSendsByAccountDateRange(
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
