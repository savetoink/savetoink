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
		if tagFilter != nil {
			query += ` AND a.favorite = ?`
		} else {
			query += ` AND favorite = ?`
		}
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
