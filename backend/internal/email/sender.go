// Package email provides email sending functionality.
package email

import (
	"context"
	"regexp"
	"strings"

	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/shaftoe/savetoink/backend/internal/model"
)

// SendEmailResponse contains the response from sending an email.
type SendEmailResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	MessageID string `json:"message_id,omitempty"`
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

	// DestEmail is the email address of the recipient, typically a
	// Kindle Personal Document Service address like "abcd@kindle.com".
	DestEmail string

	// AppURL is the base URL for the application, used in email body.
	AppURL string
}

// GenerateFilename creates a sanitized filename from the article title.
func GenerateFilename(article *model.Article) string {
	if article.Title != "" {
		return sanitizeFilename(article.Title) + ".epub"
	}
	return "article.epub"
}

// BuildSubject creates an email subject in the format "[Save to Ink] <Title>".
func BuildSubject(articleTitle string) string {
	if articleTitle == "" {
		articleTitle = "Document"
	}
	articleTitle = strings.TrimSpace(articleTitle)
	maxTitleLength := max(0, consts.MaxSubjectLength-len(consts.MailSubjectPrefix))
	if len(articleTitle) > maxTitleLength {
		articleTitle = articleTitle[:maxTitleLength]
	}
	return consts.MailSubjectPrefix + articleTitle
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
