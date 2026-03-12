package service

import (
	"context"
	"fmt"
	"io"
	"time"

	apperrors "github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
)

// SendArticle sends an EPUB document via email.
// For CLI use.
func (s *Service) SendArticle(
	ctx context.Context,
	destEmail string,
	epubData io.ReadCloser,
	title string,
) (*email.SendEmailResponse, error) {
	req := &email.Request{
		EPUBData:  epubData,
		DestEmail: destEmail,
		Body:      consts.BuildCLIEmailBody(),
		Subject:   email.BuildSubject(title),
	}
	resp, err := s.sender.SendEmail(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return resp, nil
}

// SendArticleByID retrieves an article, generates an EPUB, and sends it to the user's device email.
func (s *Service) SendArticleByID(
	ctx context.Context,
	accountID,
	articleID string,
) (*servicetypes.SendArticleResult, error) {
	deviceEmail, _, err := s.GetUserDeviceEmail(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if deviceEmail == "" {
		return nil, fmt.Errorf("%w: user device email not configured", apperrors.ErrInvalid)
	}

	article, err := s.GetArticle(ctx, accountID, articleID)
	if err != nil {
		return nil, err
	}

	epubReader, err := s.GenerateEPUB(article)
	if err != nil {
		return nil, err
	}

	if createErr := s.createSendRecord(ctx, accountID, articleID, article.Title, deviceEmail); createErr != nil {
		_ = epubReader.Close()
		return nil, createErr
	}

	emailResp, err := s.SendArticle(ctx, deviceEmail, epubReader, article.Title)
	if err != nil {
		if updateErr := s.updateSendRecordOnFailure(ctx, accountID, articleID, err); updateErr != nil {
			return nil, fmt.Errorf("%w and failed to update send record: %w", err, updateErr)
		}
		return nil, err
	}

	if updateErr := s.updateSendRecordOnSuccess(ctx, accountID, articleID, emailResp.MessageID); updateErr != nil {
		return nil, fmt.Errorf("failed to update send record: %w", updateErr)
	}

	return &servicetypes.SendArticleResult{
		Article:     article,
		DeviceEmail: deviceEmail,
		EmailResp:   emailResp,
	}, nil
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
		return 0, fmt.Errorf("failed to count sends: %w", err)
	}

	return count, nil
}

func (s *Service) createSendRecord(
	ctx context.Context,
	accountID, articleID, title, deviceEmail string,
) error {
	if s.sendsRepo == nil {
		return nil
	}
	if err := s.sendsRepo.CreateSendRecord(ctx, &model.Send{
		Account:     accountID,
		ArticleID:   articleID,
		Title:       title,
		DestEmail:   deviceEmail,
		SenderEmail: s.cfg.SenderEmail,
		Provider:    string(s.cfg.EmailProvider),
	}); err != nil {
		return fmt.Errorf("failed to create send record: %w", err)
	}

	return nil
}

func (s *Service) updateSendRecordOnFailure(
	ctx context.Context,
	accountID, articleID string,
	sendErr error,
) error {
	if s.sendsRepo == nil {
		return nil
	}
	if err := s.sendsRepo.UpdateSendRecord(ctx, &model.Send{
		Account:       accountID,
		ArticleID:     articleID,
		Status:        "failed",
		ErrorResponse: sendErr.Error(),
	}); err != nil {
		return fmt.Errorf("failed to update send record: %w", err)
	}
	return nil
}

func (s *Service) updateSendRecordOnSuccess(
	ctx context.Context,
	accountID, articleID, messageID string,
) error {
	if s.sendsRepo == nil {
		return nil
	}
	if err := s.sendsRepo.UpdateSendRecord(ctx, &model.Send{
		Account:   accountID,
		ArticleID: articleID,
		Status:    "success",
		MessageID: messageID,
	}); err != nil {
		return fmt.Errorf("failed to update send record: %w", err)
	}
	return nil
}
