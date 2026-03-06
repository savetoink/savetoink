// Package mailjet provides a Mailjet implementation of the email Sender interface.
package mailjet

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	mailjetLib "github.com/mailjet/mailjet-apiv3-go/v4"

	"github.com/shaftoe/savetoink/backend/internal/email"
)

// Sender implements the email.Sender interface using Mailjet.
type Sender struct {
	apiKey      string
	apiSecret   string
	senderEmail string
	client      *mailjetLib.Client
}

// NewSender creates a new Mailjet sender instance.
func NewSender(apiKey, apiSecret, senderEmail string) *Sender {
	mailjetClient := mailjetLib.NewMailjetClient(apiKey, apiSecret)
	return &Sender{
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		senderEmail: senderEmail,
		client:      mailjetClient,
	}
}

// SendEmail sends an email with the EPUB attachment via Mailjet.
func (s *Sender) SendEmail(ctx context.Context, req *email.Request) (*email.SendEmailResponse, error) {
	if err := s.validateConfig(); err != nil {
		return nil, fmt.Errorf("invalid sender config: %w", err)
	}

	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	messagesInfo := s.buildMessageInfo(req)
	messages := mailjetLib.MessagesV31{Info: messagesInfo}

	resp, err := s.client.SendMailV31(&messages, mailjetLib.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return s.parseResponse(resp)
}

func (s *Sender) buildMessageInfo(req *email.Request) []mailjetLib.InfoMessagesV31 {
	filename := email.GenerateFilename(req.Article)
	subject := email.BuildSubject(req.Article.Title)
	bodyText := req.Body

	base64Content := base64.StdEncoding.EncodeToString(req.EPUBData)

	return []mailjetLib.InfoMessagesV31{
		{
			From: &mailjetLib.RecipientV31{
				Email: s.senderEmail,
			},
			To: &mailjetLib.RecipientsV31{
				mailjetLib.RecipientV31{
					Email: req.DestEmail,
				},
			},
			Subject:  subject,
			TextPart: bodyText,
			Attachments: &mailjetLib.AttachmentsV31{
				mailjetLib.AttachmentV31{
					ContentType:   "application/epub+zip",
					Filename:      filename,
					Base64Content: base64Content,
				},
			},
		},
	}
}

func (s *Sender) parseResponse(resp *mailjetLib.ResultsV31) (*email.SendEmailResponse, error) {
	if len(resp.ResultsV31) == 0 {
		return nil, errors.New("no messages in response")
	}

	if resp.ResultsV31[0].Status != "success" {
		return nil, fmt.Errorf("email send failed with status: %s", resp.ResultsV31[0].Status)
	}

	result := &email.SendEmailResponse{
		Status:  "success",
		Message: "Email sent successfully",
	}

	if len(resp.ResultsV31[0].To) > 0 {
		result.MessageID = resp.ResultsV31[0].To[0].MessageUUID
	}

	return result, nil
}

func (s *Sender) validateConfig() error {
	if s.apiKey == "" {
		return errors.New("api key is required")
	}
	if s.apiSecret == "" {
		return errors.New("api secret is required")
	}
	if s.senderEmail == "" {
		return errors.New("sender email is required")
	}
	return nil
}

func (s *Sender) validateRequest(req *email.Request) error {
	if req.DestEmail == "" {
		return errors.New("device email is required")
	}
	if req.EPUBData == nil {
		return errors.New("epub data is required")
	}
	if len(req.EPUBData) == 0 {
		return errors.New("epub data is empty")
	}
	if req.Article == nil {
		return errors.New("article is required")
	}
	if req.Body == "" {
		return errors.New("email body is required")
	}
	return nil
}
