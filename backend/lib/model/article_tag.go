// Package model provides data models used throughout the application.
package model

import "time"

// ArticleTag represents a many-to-many relationship between articles and tags.
// Each record links a single article to a single tag, enabling efficient querying
// of articles by tag and managing tag collections per article.
type ArticleTag struct {
	Account            string    `json:"account" dynamodbav:"account"`
	Tag                string    `json:"tag" dynamodbav:"tag"`
	AccountTag         string    `json:"-" dynamodbav:"accountTag"` // composite key: "accountID:tag"
	ArticleID          string    `json:"articleId" dynamodbav:"articleId"`
	CreatedAt          time.Time `json:"createdAt" dynamodbav:"createdAt"`
	CreatedAtArticleID string    `json:"-" dynamodbav:"createdAtArticleId"` // composite key: "createdAt:articleID"
}
