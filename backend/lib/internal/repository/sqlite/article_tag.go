// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	maxPageSize           = 100
	numFixedFieldsForArgs = 2
)

// AddTagsToArticle adds tags to an article. If createdAt is nil, it fetches the article
// from the database to get the creation time. If createdAt is provided, it uses it directly,
// avoiding the extra database query.
func (s *SQLite) AddTagsToArticle(ctx context.Context, accountID, articleID string, tags []string,
	createdAt *time.Time) error {
	if len(tags) == 0 {
		return nil
	}

	// Deduplicate tags to prevent duplicate entries in the database
	seen := make(map[string]bool)
	uniqueTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}
	tags = uniqueTags

	var articleCreatedAt time.Time
	if createdAt != nil {
		articleCreatedAt = *createdAt
	} else {
		// Get the article's creation time for consistency with DynamoDB
		article, err := s.GetByAccountAndID(ctx, accountID, articleID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("article not found: %s", articleID)
			}
			return fmt.Errorf("failed to get article: %w", err)
		}
		articleCreatedAt = article.CreatedAt
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `
		INSERT INTO article_tags (account, tag, account_tag, article_id, created_at, created_at_article_id)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(account_tag, created_at_article_id) DO NOTHING
	`

	createdAtUnix := articleCreatedAt.Unix()
	for _, tag := range tags {
		accountTag := buildAccountTagKeySQLite(accountID, tag)
		createdAtArticleID := buildCreatedAtArticleIDKeySQLite(articleCreatedAt, articleID)

		if _, execErr := tx.ExecContext(ctx, query,
			accountID, tag, accountTag, articleID, createdAtUnix, createdAtArticleID); execErr != nil {
			return fmt.Errorf("failed to insert article tag: %w", execErr)
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

// RemoveTagsFromArticle removes specific tags from an article.
func (s *SQLite) RemoveTagsFromArticle(ctx context.Context, accountID, articleID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	// Get the article's creation time for consistency
	article, err := s.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("article not found: %s", articleID)
		}
		return fmt.Errorf("failed to get article: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	createdAtArticleID := buildCreatedAtArticleIDKeySQLite(article.CreatedAt, articleID)
	query := `
		DELETE FROM article_tags
		WHERE account = ? AND created_at_article_id = ? AND tag IN (%s)
	`

	// Build IN clause placeholders
	placeholders := make([]string, len(tags))
	args := make([]any, 0, len(tags)+numFixedFieldsForArgs)
	args = append(args, accountID, createdAtArticleID)
	for i, tag := range tags {
		placeholders[i] = "?"
		args = append(args, tag)
	}

	query = fmt.Sprintf(query, sqlJoin(placeholders, ","))
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete article tags: %w", err)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no tags found to delete for article %s", articleID)
	}

	return nil
}

// SetArticleTags replaces all tags for an article with the provided tags.
func (s *SQLite) SetArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	// Get the article's creation time for consistency
	article, err := s.GetByAccountAndID(ctx, accountID, articleID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("article not found: %s", articleID)
		}
		return fmt.Errorf("failed to get article: %w", err)
	}

	// First, delete all existing tags for this article
	err = s.DeleteTagsForArticle(ctx, accountID, articleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Then add the new tags
	return s.AddTagsToArticle(ctx, accountID, articleID, tags, &article.CreatedAt)
}

// GetArticleTags retrieves all tags for a specific article.
func (s *SQLite) GetArticleTags(ctx context.Context, accountID, articleID string) ([]string, error) {
	query := `
		SELECT DISTINCT tag
		FROM article_tags
		WHERE account = ? AND article_id = ?
		ORDER BY tag ASC
	`

	rows, queryErr := s.db.QueryContext(ctx, query, accountID, articleID)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to query article tags: %w", queryErr)
	}
	defer func() {
		_ = rows.Close()
	}()

	var tags []string
	for rows.Next() {
		var tag string
		if scanErr := rows.Scan(&tag); scanErr != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", scanErr)
		}
		tags = append(tags, tag)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil, fmt.Errorf("error iterating article tags: %w", iterErr)
	}

	return tags, nil
}

// GetArticlesByTag retrieves article IDs for articles with a specific tag for a given account.
// Supports pagination with page (1-indexed) and pageSize parameters.
func (s *SQLite) GetArticlesByTag(ctx context.Context, accountID, tag string, page, pageSize int) (
	articleIDs []string, total int, err error,
) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT article_id)
		FROM article_tags
		WHERE account = ? AND tag = ?
	`
	if scanErr := s.db.QueryRowContext(ctx, countQuery, accountID, tag).Scan(&total); scanErr != nil {
		return nil, 0, fmt.Errorf("failed to get count for tag: %w", scanErr)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	query := `
		SELECT DISTINCT article_id
		FROM article_tags
		WHERE account = ? AND tag = ?
		ORDER BY created_at DESC, article_id ASC
		LIMIT ? OFFSET ?
	`

	rows, queryErr := s.db.QueryContext(ctx, query, accountID, tag, pageSize, offset)
	if queryErr != nil {
		return nil, 0, fmt.Errorf("failed to query articles by tag: %w", queryErr)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var articleID string
		if scanErr := rows.Scan(&articleID); scanErr != nil {
			return nil, 0, fmt.Errorf("failed to scan article id: %w", scanErr)
		}
		articleIDs = append(articleIDs, articleID)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil, 0, fmt.Errorf("error iterating articles: %w", iterErr)
	}

	return articleIDs, total, nil
}

// GetAllTagsForAccount retrieves all unique tags for an account.
func (s *SQLite) GetAllTagsForAccount(ctx context.Context, accountID string) ([]string, error) {
	query := `
		SELECT DISTINCT tag
		FROM article_tags
		WHERE account = ?
		ORDER BY tag ASC
	`

	rows, queryErr := s.db.QueryContext(ctx, query, accountID)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to query all tags for account: %w", queryErr)
	}
	defer func() {
		_ = rows.Close()
	}()

	var tags []string
	for rows.Next() {
		var tag string
		if scanErr := rows.Scan(&tag); scanErr != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", scanErr)
		}
		tags = append(tags, tag)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil, fmt.Errorf("error iterating tags: %w", iterErr)
	}

	return tags, nil
}

// DeleteTagsForArticle deletes all tags for a specific article.
func (s *SQLite) DeleteTagsForArticle(ctx context.Context, accountID, articleID string) error {
	query := `
		DELETE FROM article_tags
		WHERE account = ? AND article_id = ?
	`

	_, err := s.db.ExecContext(ctx, query, accountID, articleID)
	if err != nil {
		return fmt.Errorf("failed to delete article tags: %w", err)
	}

	// Don't return error if no rows were affected - article might not have any tags
	return nil
}

// Helper functions for building SQLite keys

func buildAccountTagKeySQLite(accountID, tag string) string {
	return fmt.Sprintf("%s:%s", accountID, tag)
}

func buildCreatedAtArticleIDKeySQLite(createdAt time.Time, articleID string) string {
	return fmt.Sprintf("%d:%s", createdAt.Unix(), articleID)
}

func sqlJoin(items []string, sep string) string {
	result := ""
	var resultBuild strings.Builder
	for i, item := range items {
		if i > 0 {
			resultBuild.WriteString(sep)
		}
		resultBuild.WriteString(item)
	}
	result += resultBuild.String()
	return result
}
