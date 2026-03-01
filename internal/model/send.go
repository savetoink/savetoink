package model

import "time"

// Send represents a single email delivery event for an article.
type Send struct {
	PK            string    `dynamodbav:"pk"`
	SK            string    `dynamodbav:"sk"`
	Account       string    `dynamodbav:"account"`
	ArticleID     string    `dynamodbav:"articleId"`
	SentAt        time.Time `dynamodbav:"sentAt"`
	Title         string    `dynamodbav:"title"`
	DestEmail     string    `dynamodbav:"destEmail"`
	Status        string    `dynamodbav:"status"`
	SenderEmail   string    `dynamodbav:"senderEmail"`
	MessageID     string    `dynamodbav:"messageID,omitempty"`
	Provider      string    `dynamodbav:"provider"`
	ErrorResponse string    `dynamodbav:"errorResponse,omitempty"`
}
