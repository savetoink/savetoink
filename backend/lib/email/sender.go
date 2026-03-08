// Package email provides email sending functionality.
package email

import (
	"context"
	"errors"
	"regexp"
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

// ValidateRequest validates an email request.
func ValidateRequest(_ context.Context, req *Request) error {
	if req.DestEmail == "" {
		return errors.New("device email is required")
	}
	if req.EPUBData == nil {
		return errors.New("epub data is required")
	}
	if len(req.EPUBData) == 0 {
		return errors.New("epub data is empty")
	}
	if req.Body == "" {
		return errors.New("email body is required")
	}
	return nil
}

const (
	maxFilenameLength = 90
)

var (
	reUnsafe   = regexp.MustCompile(`[^a-zA-Z0-9._ '\-\[\]\(\)]`)
	reCollapse = regexp.MustCompile(`[._-]{2,}| {2,}`)
)

// SanitizeFilename sanitizes a filename by removing unsafe characters.
func SanitizeFilename(title string) string {
	name := reUnsafe.ReplaceAllString(title, " ")
	name = strings.TrimSpace(name)
	name = reCollapse.ReplaceAllString(name, " ")
	name = strings.Trim(name, "._-")

	if name == "" {
		name = "article"
	}
	if len(name) > maxFilenameLength {
		name = name[:maxFilenameLength]
		name = strings.TrimRight(name, "._- ")
	}

	return name + ".epub"
}
