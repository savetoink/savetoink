// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

// profileRow represents a row in the user_profiles table.
type profileRow struct {
	Account       string
	Email         sql.NullString
	DeviceEmail   sql.NullString
	AutoSend      sql.NullInt64
	BouncedEmails sql.NullString
}

func (r *profileRow) toUserProfile() (*model.UserProfile, error) {
	profile := &model.UserProfile{
		Account:     r.Account,
		Email:       r.Email.String,
		DeviceEmail: r.DeviceEmail.String,
		AutoSend:    intToBool(r.AutoSend.Int64),
	}

	if r.BouncedEmails.Valid {
		bouncedEmails := make(map[string]model.BounceInfo)
		if err := json.Unmarshal([]byte(r.BouncedEmails.String), &bouncedEmails); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bounced emails: %w", err)
		}
		profile.BouncedEmails = bouncedEmails
	}

	return profile, nil
}

func profileToRow(profile *model.UserProfile) (*profileRow, error) {
	row := &profileRow{
		Account:  profile.Account,
		AutoSend: sql.NullInt64{Int64: boolToInt(profile.AutoSend), Valid: true},
	}

	if profile.Email != "" {
		row.Email = sql.NullString{String: profile.Email, Valid: true}
	}
	if profile.DeviceEmail != "" {
		row.DeviceEmail = sql.NullString{String: profile.DeviceEmail, Valid: true}
	}

	if len(profile.BouncedEmails) > 0 {
		bouncedEmailsJSON, err := json.Marshal(profile.BouncedEmails)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal bounced emails: %w", err)
		}
		row.BouncedEmails = sql.NullString{String: string(bouncedEmailsJSON), Valid: true}
	}

	return row, nil
}

// GetUserProfile implements UserProfileRepository.GetUserProfile.
func (s *SQLite) GetUserProfile(ctx context.Context, account string) (*model.UserProfile, error) {
	query := `
		SELECT account, email, device_email, auto_send, bounced_emails
		FROM user_profiles
		WHERE account = ?
	`

	row := s.db.QueryRowContext(ctx, query, account)

	var p profileRow
	err := row.Scan(&p.Account, &p.Email, &p.DeviceEmail, &p.AutoSend, &p.BouncedEmails)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return p.toUserProfile()
}

// GetAccountIDByDeviceEmail implements UserProfileRepository.GetAccountIDByDeviceEmail.
func (s *SQLite) GetAccountIDByDeviceEmail(ctx context.Context, deviceEmail string) (string, error) {
	query := `SELECT account FROM user_profiles WHERE device_email = ? LIMIT 1`

	var account string
	err := s.db.QueryRowContext(ctx, query, deviceEmail).Scan(&account)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to query by device email: %w", err)
	}

	return account, nil
}

// PutUserProfile implements UserProfileRepository.PutUserProfile.
func (s *SQLite) PutUserProfile(ctx context.Context, profile *model.UserProfile) error {
	if profile.Account == "" {
		return errors.New("account field is required")
	}

	row, err := profileToRow(profile)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO user_profiles (account, email, device_email, auto_send, bounced_emails)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(account) DO UPDATE SET
			email = excluded.email,
			device_email = excluded.device_email,
			auto_send = excluded.auto_send,
			bounced_emails = excluded.bounced_emails
	`

	_, err = s.db.ExecContext(ctx, query,
		row.Account,
		nullStringOrEmpty(row.Email),
		nullStringOrEmpty(row.DeviceEmail),
		nullInt64OrZero(row.AutoSend),
		nullStringOrEmpty(row.BouncedEmails),
	)
	if err != nil {
		return fmt.Errorf("failed to store user profile: %w", err)
	}

	return nil
}

// DeleteUserProfile implements UserProfileRepository.DeleteUserProfile.
func (s *SQLite) DeleteUserProfile(ctx context.Context, account string) error {
	query := `DELETE FROM user_profiles WHERE account = ?`

	result, err := s.db.ExecContext(ctx, query, account)
	if err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteUserDeviceEmail implements UserProfileRepository.DeleteUserDeviceEmail.
func (s *SQLite) DeleteUserDeviceEmail(ctx context.Context, account string) error {
	query := `UPDATE user_profiles SET device_email = NULL, auto_send = 0 WHERE account = ?`

	result, err := s.db.ExecContext(ctx, query, account)
	if err != nil {
		return fmt.Errorf("failed to delete user device email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
