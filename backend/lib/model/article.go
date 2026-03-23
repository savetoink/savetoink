// Package model provides data models used throughout the application.
package model

import "time"

// Article represents all article data including content and metadata.
type Article struct {
	Account   string    `json:"account" dynamodbav:"account"`
	ID        string    `json:"id" dynamodbav:"id"`
	URL       string    `json:"url" dynamodbav:"url"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`

	// optional metadata
	Title              string     `json:"title,omitempty" dynamodbav:"title,omitempty"`
	Content            string     `json:"content,omitempty" dynamodbav:"content,omitempty"`
	Author             string     `json:"author,omitempty" dynamodbav:"author,omitempty"`
	SiteName           string     `json:"siteName,omitempty" dynamodbav:"siteName,omitempty"`
	SourceDomain       string     `json:"sourceDomain,omitempty" dynamodbav:"sourceDomain,omitempty"`
	Excerpt            string     `json:"excerpt,omitempty" dynamodbav:"excerpt,omitempty"`
	ImageURL           string     `json:"imageUrl,omitempty" dynamodbav:"imageUrl,omitempty"`
	ContentType        string     `json:"contentType,omitempty" dynamodbav:"contentType,omitempty"`
	Language           string     `json:"language,omitempty" dynamodbav:"language,omitempty"`
	Error              string     `json:"error,omitempty" dynamodbav:"error,omitempty"`
	WordCount          int        `json:"wordCount,omitempty" dynamodbav:"wordCount,omitempty"`
	ReadingTimeMinutes int        `json:"readingTimeMinutes,omitempty" dynamodbav:"readingTimeMinutes,omitempty"`
	PublishedAt        *time.Time `json:"publishedAt,omitempty" dynamodbav:"publishedAt,omitempty"`
	Favorite           bool       `json:"favorite,omitempty" dynamodbav:"favorite,omitempty"`
	AccountFavorite    string     `json:"-" dynamodbav:"accountFavorite,omitempty"`
}

// ErrorResponse represents the unified error response.
type ErrorResponse struct {
	Error string `json:"error"`
}
