// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"fmt"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

func (s *SQLite) queryArticlesByAccount(
	ctx context.Context,
	account string,
	pageSize, offset int,
	favoriteFilter *bool,
) ([]*model.Article, error) {
	query := `
		SELECT account, id, url, created_at, title, content, author, site_name,
			source_domain, excerpt, image_url, content_type, language, error,
			word_count, reading_time_minutes, published_at, favorite
		FROM articles
		WHERE account = ?
	`

	args := []any{account}
	if favoriteFilter != nil {
		query += queryAndFavoriteFilter
		args = append(args, boolToInt(*favoriteFilter))
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
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

	var articles []*model.Article
	for rows.Next() {
		var a articleRow
		scanErr := rows.Scan(
			&a.Account, &a.ID, &a.URL, &a.CreatedAt,
			&a.Title, &a.Content, &a.Author, &a.SiteName,
			&a.SourceDomain, &a.Excerpt, &a.ImageURL, &a.ContentType,
			&a.Language, &a.Error, &a.WordCount, &a.ReadingTimeMinutes,
			&a.PublishedAt, &a.Favorite,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan article row: %w", scanErr)
		}

		article, articleErr := a.toArticle()
		if articleErr != nil {
			return nil, articleErr
		}
		articles = append(articles, article)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", iterErr)
	}

	return articles, nil
}
