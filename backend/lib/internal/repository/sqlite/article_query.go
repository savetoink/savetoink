// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

// scanArticleRows scans sql.Rows into a slice of Article models.
func scanArticleRows(rows *sql.Rows) ([]*model.Article, error) {
	var articles []*model.Article
	for rows.Next() {
		var a articleRow
		err := rows.Scan(
			&a.Account, &a.ID, &a.URL, &a.CreatedAt,
			&a.Title, &a.Content, &a.Author, &a.SiteName,
			&a.SourceDomain, &a.Excerpt, &a.ImageURL, &a.ContentType,
			&a.Language, &a.Error, &a.WordCount, &a.ReadingTimeMinutes,
			&a.PublishedAt, &a.Favorite,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article row: %w", err)
		}

		article, articleErr := a.toArticle()
		if articleErr != nil {
			return nil, articleErr
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return articles, nil
}

func (s *SQLite) queryArticlesByAccount(
	ctx context.Context,
	account string,
	pageSize, offset int,
	favoriteFilter *bool,
	tagFilter *string,
) ([]*model.Article, error) {
	var query string
	var args []any

	if tagFilter != nil {
		// Use JOIN for tag filtering
		query = `
			SELECT a.account, a.id, a.url, a.created_at, a.title, a.content, a.author,
				a.site_name, a.source_domain, a.excerpt, a.image_url, a.content_type,
				a.language, a.error, a.word_count, a.reading_time_minutes, a.published_at, a.favorite
			FROM articles a
			LEFT JOIN article_tags at ON a.account = at.account AND a.id = at.article_id
			WHERE a.account = ? AND at.tag = ?
		`
		args = []any{account, *tagFilter}
	} else {
		// No tag filter, simple query
		query = `
			SELECT account, id, url, created_at, title, content, author, site_name,
				source_domain, excerpt, image_url, content_type, language, error,
				word_count, reading_time_minutes, published_at, favorite
			FROM articles
			WHERE account = ?
		`
		args = []any{account}
	}

	if favoriteFilter != nil {
		query += queryAndFavoriteFilter
		args = append(args, boolToInt(*favoriteFilter))
	}

	if tagFilter != nil {
		query += ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	} else {
		query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	}
	args = append(args, pageSize, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("failed to close rows: %v", closeErr)
		}
	}()

	return scanArticleRows(rows)
}
