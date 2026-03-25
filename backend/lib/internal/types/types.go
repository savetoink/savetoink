// Package types provides shared internal types used across the application.
package types

// ArticleFilter holds filtering options for getting articles.
type ArticleFilter struct {
	Favorite *bool
	Tag      *string
}
