// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	internaltypes "github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

// articleRow represents a row in the articles table.
type articleRow struct {
	Account            string
	ID                 string
	URL                string
	CreatedAt          string
	Title              sql.NullString
	Content            sql.NullString
	Author             sql.NullString
	SiteName           sql.NullString
	SourceDomain       sql.NullString
	Excerpt            sql.NullString
	ImageURL           sql.NullString
	ContentType        sql.NullString
	Language           sql.NullString
	Error              sql.NullString
	WordCount          sql.NullInt64
	ReadingTimeMinutes sql.NullInt64
	PublishedAt        sql.NullString
	Favorite           sql.NullInt64
}

func (r *articleRow) toArticle() (*model.Article, error) {
	createdAt, err := time.Parse(time.RFC3339, r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	article := &model.Article{
		Account:      r.Account,
		ID:           r.ID,
		URL:          r.URL,
		CreatedAt:    createdAt,
		Title:        r.Title.String,
		Content:      r.Content.String,
		Author:       r.Author.String,
		SiteName:     r.SiteName.String,
		SourceDomain: r.SourceDomain.String,
		Excerpt:      r.Excerpt.String,
		ImageURL:     r.ImageURL.String,
		ContentType:  r.ContentType.String,
		Language:     r.Language.String,
		Error:        r.Error.String,
		Favorite:     r.Favorite.Int64 == 1,
	}

	if r.WordCount.Valid {
		article.WordCount = int(r.WordCount.Int64)
	}
	if r.ReadingTimeMinutes.Valid {
		article.ReadingTimeMinutes = int(r.ReadingTimeMinutes.Int64)
	}
	if r.PublishedAt.Valid {
		publishedAt, parseErr := time.Parse(time.RFC3339, r.PublishedAt.String)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse published_at: %w", parseErr)
		}
		article.PublishedAt = &publishedAt
	}

	return article, nil
}

func articleToRow(article *model.Article) *articleRow {
	row := &articleRow{
		Account:   article.Account,
		ID:        article.ID,
		URL:       article.URL,
		CreatedAt: article.CreatedAt.UTC().Format(time.RFC3339),
		Favorite:  sql.NullInt64{Int64: boolToInt(article.Favorite), Valid: true},
	}

	if article.Title != "" {
		row.Title = sql.NullString{String: article.Title, Valid: true}
	}
	if article.Content != "" {
		row.Content = sql.NullString{String: article.Content, Valid: true}
	}
	if article.Author != "" {
		row.Author = sql.NullString{String: article.Author, Valid: true}
	}
	if article.SiteName != "" {
		row.SiteName = sql.NullString{String: article.SiteName, Valid: true}
	}
	if article.SourceDomain != "" {
		row.SourceDomain = sql.NullString{String: article.SourceDomain, Valid: true}
	}
	if article.Excerpt != "" {
		row.Excerpt = sql.NullString{String: article.Excerpt, Valid: true}
	}
	if article.ImageURL != "" {
		row.ImageURL = sql.NullString{String: article.ImageURL, Valid: true}
	}
	if article.ContentType != "" {
		row.ContentType = sql.NullString{String: article.ContentType, Valid: true}
	}
	if article.Language != "" {
		row.Language = sql.NullString{String: article.Language, Valid: true}
	}
	if article.Error != "" {
		row.Error = sql.NullString{String: article.Error, Valid: true}
	}
	if article.WordCount > 0 {
		row.WordCount = sql.NullInt64{Int64: int64(article.WordCount), Valid: true}
	}
	if article.ReadingTimeMinutes > 0 {
		row.ReadingTimeMinutes = sql.NullInt64{Int64: int64(article.ReadingTimeMinutes), Valid: true}
	}
	if article.PublishedAt != nil {
		row.PublishedAt = sql.NullString{String: article.PublishedAt.UTC().Format(time.RFC3339), Valid: true}
	}

	return row
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int64) bool {
	return i == 1
}

// Store saves an article to SQLite.
func (s *SQLite) Store(ctx context.Context, article *model.Article) error {
	if article.Account == "" {
		return errors.New("account field is required")
	}

	row := articleToRow(article)

	query := `
		INSERT INTO articles (
			account, id, url, created_at, title, content, author, site_name,
			source_domain, excerpt, image_url, content_type, language, error,
			word_count, reading_time_minutes, published_at, favorite
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(account, id) DO UPDATE SET
			url = excluded.url,
			title = excluded.title,
			content = excluded.content,
			author = excluded.author,
			site_name = excluded.site_name,
			source_domain = excluded.source_domain,
			excerpt = excluded.excerpt,
			image_url = excluded.image_url,
			content_type = excluded.content_type,
			language = excluded.language,
			error = excluded.error,
			word_count = excluded.word_count,
			reading_time_minutes = excluded.reading_time_minutes,
			published_at = excluded.published_at,
			favorite = excluded.favorite
	`

	_, err := s.db.ExecContext(ctx, query,
		row.Account, row.ID, row.URL, row.CreatedAt,
		nullStringOrEmpty(row.Title),
		nullStringOrEmpty(row.Content),
		nullStringOrEmpty(row.Author),
		nullStringOrEmpty(row.SiteName),
		nullStringOrEmpty(row.SourceDomain),
		nullStringOrEmpty(row.Excerpt),
		nullStringOrEmpty(row.ImageURL),
		nullStringOrEmpty(row.ContentType),
		nullStringOrEmpty(row.Language),
		nullStringOrEmpty(row.Error),
		nullInt64OrZero(row.WordCount),
		nullInt64OrZero(row.ReadingTimeMinutes),
		nullStringOrEmpty(row.PublishedAt),
		row.Favorite,
	)
	if err != nil {
		return fmt.Errorf("failed to store article: %w", err)
	}

	return nil
}

func nullStringOrEmpty(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func nullInt64OrZero(ni sql.NullInt64) any {
	if ni.Valid {
		return ni.Int64
	}
	return nil
}

// GetByAccountAndID implements ArticlesRepository.GetByAccountAndID.
func (s *SQLite) GetByAccountAndID(ctx context.Context, account, id string) (*model.Article, error) {
	query := `
		SELECT account, id, url, created_at, title, content, author, site_name,
			source_domain, excerpt, image_url, content_type, language, error,
			word_count, reading_time_minutes, published_at, favorite
		FROM articles
		WHERE account = ? AND id = ?
	`

	row := s.db.QueryRowContext(ctx, query, account, id)

	var a articleRow
	err := row.Scan(
		&a.Account, &a.ID, &a.URL, &a.CreatedAt,
		&a.Title, &a.Content, &a.Author, &a.SiteName,
		&a.SourceDomain, &a.Excerpt, &a.ImageURL, &a.ContentType,
		&a.Language, &a.Error, &a.WordCount, &a.ReadingTimeMinutes,
		&a.PublishedAt, &a.Favorite,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return a.toArticle()
}

// DeleteByAccountAndID implements ArticlesRepository.DeleteByAccountAndID.
func (s *SQLite) DeleteByAccountAndID(ctx context.Context, account, id string) error {
	query := `DELETE FROM articles WHERE account = ? AND id = ?`

	result, err := s.db.ExecContext(ctx, query, account, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
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

// UpdateFavorite implements ArticlesRepository.UpdateFavorite.
func (s *SQLite) UpdateFavorite(ctx context.Context, account, id string, favorite bool) error {
	query := `UPDATE articles SET favorite = ? WHERE account = ? AND id = ?`

	result, err := s.db.ExecContext(ctx, query, boolToInt(favorite), account, id)
	if err != nil {
		return fmt.Errorf("failed to update favorite: %w", err)
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

// GetMetadataByAccount implements ArticlesRepository.GetMetadataByAccount.
func (s *SQLite) GetMetadataByAccount(
	ctx context.Context,
	account string,
	page, pageSize int,
	filter *internaltypes.ArticleFilter,
) (articles []*model.Article, total int, err error) {
	if page < 1 || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var favoriteFilter *bool
	if filter != nil {
		favoriteFilter = filter.Favorite
	}

	total, err = s.countArticlesByAccount(ctx, account, favoriteFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count articles: %w", err)
	}

	if total == 0 {
		return []*model.Article{}, 0, nil
	}

	offset := (page - 1) * pageSize
	if offset >= total {
		return []*model.Article{}, total, nil
	}

	articles, err = s.queryArticlesByAccount(ctx, account, pageSize, offset, favoriteFilter)
	if err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (s *SQLite) countArticlesByAccount(ctx context.Context, account string, favoriteFilter *bool) (int, error) {
	query := `SELECT COUNT(*) FROM articles WHERE account = ?`

	args := []any{account}
	if favoriteFilter != nil {
		query += queryAndFavoriteFilter
		args = append(args, boolToInt(*favoriteFilter))
	}

	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count articles: %w", err)
	}

	return count, nil
}
