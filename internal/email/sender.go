// Package email provides email sending functionality.
package email

import (
	"context"
	"regexp"
	"strings"

	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
)

// SendEmailResponse contains the response from sending an email.
type SendEmailResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	EmailUUID string `json:"email_uuid,omitempty"`
}

// Sender defines the interface for sending emails.
type Sender interface {
	SendEmail(ctx context.Context, req *Request) (*SendEmailResponse, error)
}

// Request contains the data required to send an email.
type Request struct {
	// Article is the article to be sent.
	Article *model.Article

	// EPUBData is the EPUB data to be sent as attachment.
	EPUBData []byte

	// Subject is the email subject. nil means use default (article title or consts.DefaultSubject).
	Subject *string

	// DestEmail is the email address of the recipient, typically a
	// Kindle Personal Document Service address like "abcd@kindle.com".
	DestEmail string

	// BodyText is the email body text. nil means use consts.DefaultBodyText.
	BodyText *string
}

// GenerateFilename creates a sanitized filename from the article title.
func GenerateFilename(article *model.Article) string {
	if article.Title != "" {
		return sanitizeFilename(article.Title) + ".epub"
	}
	return "article.epub"
}

// GenerateSubject creates an email subject from the article title or custom subject.
func GenerateSubject(articleTitle, customSubject string) string {
	if customSubject != "" {
		return sanitizeSubject(customSubject)
	}
	if articleTitle != "" {
		return sanitizeSubject(articleTitle)
	}
	return consts.DefaultSubject
}

func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[^\w\s-]`)
	sanitized := re.ReplaceAllString(name, "")
	sanitized = strings.TrimSpace(sanitized)
	if sanitized == "" {
		return "article"
	}
	return sanitized
}

func sanitizeSubject(subject string) string {
	if subject == "" {
		return consts.DefaultSubject
	}
	subject = strings.TrimSpace(subject)
	if len(subject) > consts.MaxSubjectLength {
		subject = subject[:consts.MaxSubjectLength]
	}
	return subject
}
