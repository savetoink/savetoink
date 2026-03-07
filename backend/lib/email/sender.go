// Package email provides email sending functionality.
package email

import (
	"context"
	"strings"

	"github.com/shaftoe/savetoink/backend/lib/consts"
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
	EPUBData  []byte
	DestEmail string
	Body      string
	Subject   string
}

// NewRequest creates a new email request with the provided data.
func NewRequest(epubData []byte, destEmail, appURL string) *Request {
	return &Request{
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
