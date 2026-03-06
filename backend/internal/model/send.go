package model

import "time"

// Send represents a single email delivery event for an article.
type Send struct {
	Account       string
	ArticleID     string
	SentAt        time.Time
	Title         string
	DestEmail     string
	Status        string
	SenderEmail   string
	MessageID     string
	Provider      string
	ErrorResponse string
}
