// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite" //nolint:blankimports // required for SQLite driver
)

const (
	queryAndFavoriteFilter = " AND favorite = ?"
)

// SQLite implements Repository interface using SQLite.
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite repository instance.
func NewSQLite(ctx context.Context, dbPath string) (*SQLite, error) {
	if dbPath == "" {
		return nil, errors.New("database path is required")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if errPing := db.PingContext(ctx); errPing != nil {
		return nil, fmt.Errorf("failed to ping database: %w", errPing)
	}

	sqlite := &SQLite{db: db}
	if errCreate := sqlite.CreateTables(ctx); errCreate != nil {
		return nil, fmt.Errorf("failed to create tables: %w", errCreate)
	}

	return sqlite, nil
}

// Close closes the database connection.
func (s *SQLite) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

// CreateTables creates all required tables if they don't exist.
func (s *SQLite) CreateTables(ctx context.Context) error {
	if err := s.createArticleTable(ctx); err != nil {
		return fmt.Errorf("failed to create article table: %w", err)
	}
	if err := s.createUserProfileTable(ctx); err != nil {
		return fmt.Errorf("failed to create user profile table: %w", err)
	}
	if err := s.createSendsTable(ctx); err != nil {
		return fmt.Errorf("failed to create sends table: %w", err)
	}
	return nil
}

func (s *SQLite) createArticleTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS articles (
			account TEXT NOT NULL,
			id TEXT NOT NULL,
			url TEXT NOT NULL,
			created_at TEXT NOT NULL,
			title TEXT,
			content TEXT,
			author TEXT,
			site_name TEXT,
			source_domain TEXT,
			excerpt TEXT,
			image_url TEXT,
			content_type TEXT,
			language TEXT,
			error TEXT,
			word_count INTEGER,
			reading_time_minutes INTEGER,
			published_at TEXT,
			favorite INTEGER DEFAULT 0,
			PRIMARY KEY (account, id)
		);
		CREATE INDEX IF NOT EXISTS idx_articles_account_created ON articles(account, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_articles_account_favorite ON articles(account, favorite) WHERE favorite = 1;
	`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create article table: %w", err)
	}
	return nil
}

func (s *SQLite) createUserProfileTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_profiles (
			account TEXT PRIMARY KEY,
			email TEXT,
			device_email TEXT,
			auto_send INTEGER DEFAULT 0,
			bounced_emails TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_user_profiles_device_email ON user_profiles(device_email);
	`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create user profile table: %w", err)
	}
	return nil
}

func (s *SQLite) createSendsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS sends (
			account TEXT NOT NULL,
			article_id TEXT NOT NULL,
			sent_at TEXT NOT NULL,
			title TEXT,
			dest_email TEXT,
			status TEXT,
			sender_email TEXT,
			message_id TEXT,
			provider TEXT,
			error_response TEXT,
			PRIMARY KEY (account, sent_at, article_id)
		);
		CREATE INDEX IF NOT EXISTS idx_sends_article_id ON sends(article_id);
		CREATE INDEX IF NOT EXISTS idx_sends_account_sent_at ON sends(account, sent_at);
	`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create sends table: %w", err)
	}
	return nil
}
