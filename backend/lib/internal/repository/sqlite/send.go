// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

// sendRow represents a row in the sends table.
type sendRow struct {
	Account       string
	ArticleID     string
	SentAt        string
	Title         sql.NullString
	DestEmail     sql.NullString
	Status        sql.NullString
	SenderEmail   sql.NullString
	MessageID     sql.NullString
	Provider      sql.NullString
	ErrorResponse sql.NullString
}

func (r *sendRow) toSend() (*model.Send, error) {
	sentAt, err := time.Parse(time.RFC3339, r.SentAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sent_at: %w", err)
	}

	return &model.Send{
		Account:       r.Account,
		ArticleID:     r.ArticleID,
		SentAt:        sentAt,
		Title:         r.Title.String,
		DestEmail:     r.DestEmail.String,
		Status:        r.Status.String,
		SenderEmail:   r.SenderEmail.String,
		MessageID:     r.MessageID.String,
		Provider:      r.Provider.String,
		ErrorResponse: r.ErrorResponse.String,
	}, nil
}

func (s *SQLite) querySends(ctx context.Context, query string, args ...any) ([]*model.Send, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sends: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("failed to close rows: %v", closeErr)
		}
	}()

	var sends []*model.Send
	for rows.Next() {
		var sr sendRow
		scanErr := rows.Scan(
			&sr.Account, &sr.ArticleID, &sr.SentAt,
			&sr.Title, &sr.DestEmail, &sr.Status,
			&sr.SenderEmail, &sr.MessageID, &sr.Provider, &sr.ErrorResponse,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan send row: %w", scanErr)
		}

		send, sendErr := sr.toSend()
		if sendErr != nil {
			return nil, sendErr
		}
		sends = append(sends, send)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", iterErr)
	}

	return sends, nil
}

// CreateSendRecord implements SendsRepository.CreateSendRecord.
func (s *SQLite) CreateSendRecord(ctx context.Context, send *model.Send) error {
	now := time.Now().UTC()
	sentAt := now.Format(time.RFC3339)

	query := `
		INSERT INTO sends (
			account, article_id, sent_at, title, dest_email, status,
			sender_email, message_id, provider, error_response
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		send.Account,
		send.ArticleID,
		sentAt,
		nullStringOrEmpty(sql.NullString{String: send.Title, Valid: send.Title != ""}),
		nullStringOrEmpty(sql.NullString{String: send.DestEmail, Valid: send.DestEmail != ""}),
		"pending",
		nullStringOrEmpty(sql.NullString{String: send.SenderEmail, Valid: send.SenderEmail != ""}),
		nullStringOrEmpty(sql.NullString{String: send.MessageID, Valid: send.MessageID != ""}),
		nullStringOrEmpty(sql.NullString{String: send.Provider, Valid: send.Provider != ""}),
		nullStringOrEmpty(sql.NullString{String: send.ErrorResponse, Valid: send.ErrorResponse != ""}),
	)
	if err != nil {
		return fmt.Errorf("failed to create send record: %w", err)
	}

	return nil
}

// UpdateSendRecord implements SendsRepository.UpdateSendRecord.
func (s *SQLite) UpdateSendRecord(ctx context.Context, send *model.Send) error {
	query := `
		UPDATE sends
		SET status = ?, message_id = ?, error_response = ?
		WHERE account = ? AND article_id = ? AND sent_at = (
			SELECT sent_at FROM sends WHERE account = ? AND article_id = ?
			ORDER BY sent_at DESC LIMIT 1
		)
	`

	result, err := s.db.ExecContext(ctx, query,
		nullStringOrEmpty(sql.NullString{String: send.Status, Valid: send.Status != ""}),
		nullStringOrEmpty(sql.NullString{String: send.MessageID, Valid: send.MessageID != ""}),
		nullStringOrEmpty(sql.NullString{String: send.ErrorResponse, Valid: send.ErrorResponse != ""}),
		send.Account,
		send.ArticleID,
		send.Account,
		send.ArticleID,
	)
	if err != nil {
		return fmt.Errorf("failed to update send record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("send record not found for article %s", send.ArticleID)
	}

	return nil
}

// GetSendsByArticleID implements SendsRepository.GetSendsByArticleID.
func (s *SQLite) GetSendsByArticleID(ctx context.Context, articleID string) ([]*model.Send, error) {
	query := `
		SELECT account, article_id, sent_at, title, dest_email, status,
			sender_email, message_id, provider, error_response
		FROM sends
		WHERE article_id = ?
		ORDER BY sent_at DESC
	`

	return s.querySends(ctx, query, articleID)
}

// GetSendsByAccountDateRange implements SendsRepository.GetSendsByAccountDateRange.
func (s *SQLite) GetSendsByAccountDateRange(
	ctx context.Context,
	account string,
	startDate, endDate time.Time,
) ([]*model.Send, error) {
	query := `
		SELECT account, article_id, sent_at, title, dest_email, status,
			sender_email, message_id, provider, error_response
		FROM sends
		WHERE account = ? AND sent_at BETWEEN ? AND ?
		ORDER BY sent_at DESC
	`

	startDateStr := startDate.UTC().Format(time.RFC3339)
	endDateStr := endDate.UTC().Format(time.RFC3339)

	return s.querySends(ctx, query, account, startDateStr, endDateStr)
}

// CountSendsByAccountDateRange implements SendsRepository.CountSendsByAccountDateRange.
func (s *SQLite) CountSendsByAccountDateRange(
	ctx context.Context,
	account string,
	startDate, endDate time.Time,
) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM sends
		WHERE account = ? AND sent_at BETWEEN ? AND ?
	`

	startDateStr := startDate.UTC().Format(time.RFC3339)
	endDateStr := endDate.UTC().Format(time.RFC3339)

	var count int
	err := s.db.QueryRowContext(ctx, query, account, startDateStr, endDateStr).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sends: %w", err)
	}

	return count, nil
}
