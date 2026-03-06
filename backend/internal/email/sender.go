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
	Article   *model.Article
	EPUBData  []byte
	DestEmail string
	Body      string
}

// GenerateFilename creates a sanitized filename from the article title.
func GenerateFilename(article *model.Article) string {
	if article.Title != "" {
		return sanitizeFilename(article.Title) + ".epub"
	}
	return "article.epub"
}

// NewRequest creates a new email request with the provided data.
func NewRequest(article *model.Article, epubData []byte, destEmail, appURL string) *Request {
	return &Request{
		Article:   article,
		EPUBData:  epubData,
		DestEmail: destEmail,
		Body:      consts.BuildEmailBody(appURL),
	}
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
