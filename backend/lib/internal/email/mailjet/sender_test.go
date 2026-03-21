package mailjet

import (
	"context"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	mailjetLib "github.com/mailjet/mailjet-apiv3-go/v4"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
)

const (
	attachmentContentType = "application/epub+zip"
	defaultFilename       = "article.epub"
	testSenderEmail       = "test@example.com"
	testSubject           = "Test Subject"
	testKindleEmail       = "kindle@kindle.com"
	testEmailBody         = "email body"
	testUserKindleEmail   = "user@kindle.com"
	testArticle           = "Test Article"
	testMsgID             = "msg-12345"
)

func TestNewSender(t *testing.T) {
	sender := NewSender("test-key", "test-secret", testSenderEmail)
	if sender == nil {
		t.Fatal("NewSender returned nil")
	}

	if sender.apiKey != "test-key" {
		t.Error("Sender API key not set correctly")
	}
	if sender.apiSecret != "test-secret" {
		t.Error("Sender API secret not set correctly")
	}
	if sender.senderEmail != testSenderEmail {
		t.Error("Sender email not set correctly")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		apiSecret   string
		senderEmail string
		wantErr     bool
	}{
		{
			name:        "valid config",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: testSenderEmail,
			wantErr:     false,
		},
		{
			name:        "missing api key",
			apiKey:      "",
			apiSecret:   "secret",
			senderEmail: testSenderEmail,
			wantErr:     true,
		},
		{
			name:        "missing api secret",
			apiKey:      "key",
			apiSecret:   "",
			senderEmail: testSenderEmail,
			wantErr:     true,
		},
		{
			name:        "missing sender email",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewSender(tt.apiKey, tt.apiSecret, tt.senderEmail)
			err := sender.validateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *email.Request
		wantErr bool
	}{
		{
			name: "valid request",
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("test epub data")),
				DestEmail: testKindleEmail,
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr: false,
		},
		{
			name: "missing device email",
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: "",
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr: true,
		},
		{
			name: "missing epub data",
			req: &email.Request{
				EPUBData:  nil,
				DestEmail: testKindleEmail,
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr: true,
		},
		{
			name: "missing body",
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: testKindleEmail,
				Body:      "",
				Subject:   testSubject,
			},
			wantErr: true,
		},
		{
			name: "empty epub data",
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("")),
				DestEmail: testKindleEmail,
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := email.ValidateRequest(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendEmailValidation(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		apiSecret   string
		senderEmail string
		req         *email.Request
		wantErr     bool
		expectResp  *email.SendEmailResponse
	}{
		{
			name:        "missing api key in config",
			apiKey:      "",
			apiSecret:   "secret",
			senderEmail: testSenderEmail,
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: testKindleEmail,
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing api secret in config",
			apiKey:      "key",
			apiSecret:   "",
			senderEmail: testSenderEmail,
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: testKindleEmail,
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing sender email in config",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: "",
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: testKindleEmail,
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing device email in request",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: testSenderEmail,
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: "",
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing epub data in request",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: testSenderEmail,
			req: &email.Request{
				EPUBData:  nil,
				DestEmail: testKindleEmail,
				Body:      testEmailBody,
				Subject:   testSubject,
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing body in request",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: testSenderEmail,
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: testKindleEmail,
				Body:      "",
				Subject:   testSubject,
			},
			wantErr:    true,
			expectResp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sender := NewSender(tt.apiKey, tt.apiSecret, tt.senderEmail)
			resp, err := sender.SendEmail(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SendEmail() expected error, got nil")
				}
				if resp != nil {
					t.Errorf("SendEmail() expected nil response on error, got %v", resp)
				}
				return
			}

			if err != nil {
				t.Errorf("SendEmail() unexpected error = %v", err)
				return
			}

			if tt.expectResp != nil {
				if resp.Status != tt.expectResp.Status {
					t.Errorf("SendEmail() Status = %v, want %v", resp.Status, tt.expectResp.Status)
				}
				if resp.Message != tt.expectResp.Message {
					t.Errorf("SendEmail() Message = %v, want %v", resp.Message, tt.expectResp.Message)
				}
			}
		})
	}
}

func TestBuildMessageInfo(t *testing.T) {
	tests := []struct {
		name           string
		req            *email.Request
		senderEmail    string
		expectFilename string
	}{
		{
			name: "valid request builds correct message info",
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("test epub data")),
				DestEmail: testKindleEmail,
				Body:      testEmailBody,
				Subject:   testArticle,
			},
			senderEmail:    testSenderEmail,
			expectFilename: "Test Article.epub",
		},
		{
			name: "request with empty subject",
			req: &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: testUserKindleEmail,
				Body:      "body",
				Subject:   "",
			},
			senderEmail:    testSenderEmail,
			expectFilename: "article.epub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewSender("key", "secret", tt.senderEmail)
			messages, err := sender.buildMessageInfo(tt.req)
			if err != nil {
				t.Fatalf("buildMessageInfo() unexpected error = %v", err)
			}

			if len(messages) != 1 {
				t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
			}

			msg := messages[0]

			if msg.From == nil {
				t.Error("buildMessageInfo() From should not be nil")
			} else if msg.From.Email != tt.senderEmail {
				t.Errorf("buildMessageInfo() From.Email = %v, want %v", msg.From.Email, tt.senderEmail)
			}

			switch {
			case msg.To == nil:
				t.Error("buildMessageInfo() To should not be nil")
			case len(*msg.To) != 1:
				t.Errorf("buildMessageInfo() To length = %d, want 1", len(*msg.To))
			case (*msg.To)[0].Email != tt.req.DestEmail:
				t.Errorf("buildMessageInfo() To[0].Email = %v, want %v", (*msg.To)[0].Email, tt.req.DestEmail)
			}

			expectedSubject := tt.req.Subject
			if expectedSubject == "" {
				expectedSubject = consts.DefaultEmailSubject
			}
			if msg.Subject != expectedSubject {
				t.Errorf("buildMessageInfo() Subject = %v, want %v", msg.Subject, expectedSubject)
			}

			if msg.TextPart != tt.req.Body {
				t.Errorf("buildMessageInfo() TextPart = %v, want %v", msg.TextPart, tt.req.Body)
			}

			switch {
			case msg.Attachments == nil:
				t.Error("buildMessageInfo() Attachments should not be nil")
			case len(*msg.Attachments) != 1:
				t.Errorf("buildMessageInfo() Attachments length = %d, want 1", len(*msg.Attachments))
			default:
				att := (*msg.Attachments)[0]
				if att.ContentType != attachmentContentType {
					t.Errorf("buildMessageInfo() Attachment.ContentType = %v, want %v", att.ContentType, attachmentContentType)
				}

				if att.Filename != tt.expectFilename {
					t.Errorf("buildMessageInfo() Attachment.Filename = %v, want %v", att.Filename, tt.expectFilename)
				}

				if att.Base64Content == "" {
					t.Error("buildMessageInfo() Attachment.Base64Content should not be empty")
				}
			}
		})
	}
}

func TestBuildMessageInfo_LongTitle(t *testing.T) {
	longTitle := strings.Repeat("a", 150)
	req := &email.Request{
		EPUBData:  io.NopCloser(strings.NewReader("data")),
		DestEmail: testUserKindleEmail,
		Body:      "body",
		Subject:   longTitle,
	}
	sender := NewSender("key", "secret", testSenderEmail)
	messages, err := sender.buildMessageInfo(req)
	if err != nil {
		t.Fatalf("buildMessageInfo() unexpected error = %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
	}

	msg := messages[0]
	if msg.Subject != longTitle {
		t.Errorf("buildMessageInfo() Subject = %v, want %v", msg.Subject, longTitle)
	}
}

func TestBuildMessageInfo_UnicodeTitle(t *testing.T) {
	title := "Test article"
	req := &email.Request{
		EPUBData:  io.NopCloser(strings.NewReader("data")),
		DestEmail: testUserKindleEmail,
		Body:      "body",
		Subject:   title,
	}
	sender := NewSender("key", "secret", testSenderEmail)
	messages, err := sender.buildMessageInfo(req)
	if err != nil {
		t.Fatalf("buildMessageInfo() unexpected error = %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
	}

	msg := messages[0]
	if msg.Subject != title {
		t.Errorf("buildMessageInfo() Subject = %v, want %v", msg.Subject, title)
	}

	switch {
	case msg.Attachments == nil:
		t.Fatal("buildMessageInfo() Attachments should not be nil")
	case len(*msg.Attachments) != 1:
		t.Fatal("buildMessageInfo() should have exactly one attachment")
	default:
		att := (*msg.Attachments)[0]
		expectedFilename := "Test article.epub"
		if att.Filename != expectedFilename {
			t.Errorf("buildMessageInfo() Attachment.Filename = %v, want %v", att.Filename, expectedFilename)
		}
	}
}

func TestBuildMessageInfo_VaryingEPUBSizes(t *testing.T) {
	tests := []struct {
		name     string
		epubSize int
	}{
		{"1 byte", 1},
		{"1KB", 1024},
		{"1MB", 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epubData := make([]byte, tt.epubSize)
			for i := range epubData {
				epubData[i] = byte(i % 256)
			}

			req := &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader(string(epubData))),
				DestEmail: testUserKindleEmail,
				Body:      "body",
				Subject:   testArticle,
			}
			sender := NewSender("key", "secret", testSenderEmail)
			messages, err := sender.buildMessageInfo(req)
			if err != nil {
				t.Fatalf("buildMessageInfo() unexpected error = %v", err)
			}

			if len(messages) != 1 {
				t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
			}

			msg := messages[0]
			switch {
			case msg.Attachments == nil:
				t.Fatal("buildMessageInfo() Attachments should not be nil")
			case len(*msg.Attachments) != 1:
				t.Fatal("buildMessageInfo() should have exactly one attachment")
			default:
				att := (*msg.Attachments)[0]
				if att.ContentType != attachmentContentType {
					t.Errorf("buildMessageInfo() Attachment.ContentType = %v, want %v", att.ContentType, attachmentContentType)
				}

				decoded, decodeErr := base64.StdEncoding.DecodeString(att.Base64Content)
				if decodeErr != nil {
					t.Fatalf("buildMessageInfo() failed to decode base64 content: %v", decodeErr)
				}
				if len(decoded) != tt.epubSize {
					t.Errorf("buildMessageInfo() decoded content size = %d, want %d", len(decoded), tt.epubSize)
				}
			}
		})
	}
}

func TestBuildMessageInfo_BodyWithNewlines(t *testing.T) {
	req := &email.Request{
		EPUBData:  io.NopCloser(strings.NewReader("data")),
		DestEmail: testUserKindleEmail,
		Body:      "Line 1\nLine 2\nLine 3",
		Subject:   testArticle,
	}
	sender := NewSender("key", "secret", testSenderEmail)
	messages, err := sender.buildMessageInfo(req)
	if err != nil {
		t.Fatalf("buildMessageInfo() unexpected error = %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
	}

	msg := messages[0]
	if msg.TextPart != req.Body {
		t.Errorf("buildMessageInfo() TextPart = %v, want %v", msg.TextPart, req.Body)
	}
}

func TestBuildMessageInfo_SenderEmailVariations(t *testing.T) {
	tests := []struct {
		name        string
		senderEmail string
	}{
		{"simple email", "sender@example.com"},
		{"email with subdomain", "sender@mail.example.com"},
		{"email with dots", "first.last@example.com"},
		{"email with plus", "sender+tag@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: testUserKindleEmail,
				Body:      "body",
				Subject:   testArticle,
			}
			sender := NewSender("key", "secret", tt.senderEmail)
			messages, err := sender.buildMessageInfo(req)
			if err != nil {
				t.Fatalf("buildMessageInfo() unexpected error = %v", err)
			}

			if len(messages) != 1 {
				t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
			}

			msg := messages[0]
			if msg.From == nil || msg.From.Email != tt.senderEmail {
				t.Errorf("buildMessageInfo() From.Email = %v, want %v", msg.From.Email, tt.senderEmail)
			}
		})
	}
}

func TestBuildMessageInfo_DestEmailVariations(t *testing.T) {
	tests := []struct {
		name      string
		destEmail string
	}{
		{"kindle email", "user@kindle.com"},
		{"free kindle email", "user@free.kindle.com"},
		{"kobo email", "user@send.kobo.com"},
		{"tolino email", "user@mytolino.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: tt.destEmail,
				Body:      "body",
				Subject:   testArticle,
			}
			sender := NewSender("key", "secret", testSenderEmail)
			messages, err := sender.buildMessageInfo(req)
			if err != nil {
				t.Fatalf("buildMessageInfo() unexpected error = %v", err)
			}

			if len(messages) != 1 {
				t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
			}

			msg := messages[0]
			switch {
			case msg.To == nil:
				t.Fatalf("buildMessageInfo() should have exactly one recipient")
			case len(*msg.To) != 1:
				t.Fatalf("buildMessageInfo() should have exactly one recipient")
			case (*msg.To)[0].Email != tt.destEmail:
				t.Errorf("buildMessageInfo() To[0].Email = %v, want %v", (*msg.To)[0].Email, tt.destEmail)
			}
		})
	}
}

func TestBuildMessageInfo_SubjectEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		subject        string
		expectFilename string
	}{
		{"subject with colon", "Test: Article", "Test Article.epub"},
		{"subject with slash", "Test/Article", "Test Article.epub"},
		{"subject with backslash", "Test\\Article", "Test Article.epub"},
		{"subject with leading spaces", "   Test Article", "Test Article.epub"},
		{"subject with trailing spaces", "Test Article   ", "Test Article.epub"},
		{"subject with quotes", "\"Test Article\"", "Test Article.epub"},
		{"subject with multiple dots", "Test...Article", "Test Article.epub"},
		{"subject with unicode and emoji", "Test 中文 😊", "Test.epub"},
		{"subject with mixed case", "TeSt ArTiClE", "TeSt ArTiClE.epub"},
		{"subject with numbers", "Test Article 2024", "Test Article 2024.epub"},
		{"subject with brackets", "Test (Article) [2024]", "Test (Article) [2024].epub"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: "user@kindle.com",
				Body:      "body",
				Subject:   tt.subject,
			}
			sender := NewSender("key", "secret", "test@example.com")
			messages, err := sender.buildMessageInfo(req)
			if err != nil {
				t.Fatalf("buildMessageInfo() unexpected error = %v", err)
			}

			if len(messages) != 1 {
				t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
			}

			msg := messages[0]
			if msg.Subject != tt.subject {
				t.Errorf("buildMessageInfo() Subject = %v, want %v", msg.Subject, tt.subject)
			}

			switch {
			case msg.Attachments == nil:
				t.Fatal("buildMessageInfo() Attachments should not be nil")
			case len(*msg.Attachments) != 1:
				t.Fatal("buildMessageInfo() should have exactly one attachment")
			default:
				att := (*msg.Attachments)[0]
				if att.Filename != tt.expectFilename {
					t.Errorf("buildMessageInfo() Attachment.Filename = %v, want %v", att.Filename, tt.expectFilename)
				}
			}
		})
	}
}

func TestBuildMessageInfo_BodySpecialCharacters(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"body with unicode", "Hello 世界"},
		{"body with emoji", "Hello 😊"},
		{"body with multiple languages", "Hello 世界 Bonjour 😊"},
		{"body with tabs", "Line1\tLine2"},
		{"body with special symbols", "Hello @#$%^&*()"},
		{"body with quotes", `He said "Hello"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &email.Request{
				EPUBData:  io.NopCloser(strings.NewReader("data")),
				DestEmail: "user@kindle.com",
				Body:      tt.body,
				Subject:   "Test",
			}
			sender := NewSender("key", "secret", "test@example.com")
			messages, err := sender.buildMessageInfo(req)
			if err != nil {
				t.Fatalf("buildMessageInfo() unexpected error = %v", err)
			}

			if len(messages) != 1 {
				t.Fatalf("buildMessageInfo() returned %d messages, want 1", len(messages))
			}

			msg := messages[0]
			if msg.TextPart != tt.body {
				t.Errorf("buildMessageInfo() TextPart = %v, want %v", msg.TextPart, tt.body)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name      string
		resp      *mailjetLib.ResultsV31
		wantErr   bool
		errMsg    string
		wantMsgID string
	}{
		{
			name: "successful response with message ID",
			resp: &mailjetLib.ResultsV31{
				ResultsV31: []mailjetLib.ResultV31{
					{
						Status: "success",
						To: []mailjetLib.GeneratedMessageV31{
							{
								MessageUUID: "msg-12345",
							},
						},
					},
				},
			},
			wantErr:   false,
			wantMsgID: "msg-12345",
		},
		{
			name: "successful response without message ID",
			resp: &mailjetLib.ResultsV31{
				ResultsV31: []mailjetLib.ResultV31{
					{
						Status: "success",
						To:     []mailjetLib.GeneratedMessageV31{},
					},
				},
			},
			wantErr:   false,
			wantMsgID: "",
		},
		{
			name: "response with failure status",
			resp: &mailjetLib.ResultsV31{
				ResultsV31: []mailjetLib.ResultV31{
					{
						Status: "error",
						To: []mailjetLib.GeneratedMessageV31{
							{
								MessageUUID: "msg-12345",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "email send failed with status: error",
		},
		{
			name: "response with another failure status",
			resp: &mailjetLib.ResultsV31{
				ResultsV31: []mailjetLib.ResultV31{
					{
						Status: "failed",
						To: []mailjetLib.GeneratedMessageV31{
							{
								MessageUUID: "msg-12345",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "email send failed with status: failed",
		},
		{
			name:    "empty response",
			resp:    &mailjetLib.ResultsV31{ResultsV31: []mailjetLib.ResultV31{}},
			wantErr: true,
			errMsg:  "no messages in response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewSender("key", "secret", "test@example.com")

			resp, err := sender.parseResponse(tt.resp)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				if err.Error() != tt.errMsg {
					t.Errorf("parseResponse() error message = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if resp == nil {
				t.Fatal("parseResponse() returned nil response")
			}

			if resp.Status != "success" {
				t.Errorf("parseResponse() Status = %v, want 'success'", resp.Status)
			}
			if resp.Message != "Email sent successfully" {
				t.Errorf("parseResponse() Message = %v, want 'Email sent successfully'", resp.Message)
			}

			if resp.MessageID != tt.wantMsgID {
				t.Errorf("parseResponse() MessageID = %v, want %v", resp.MessageID, tt.wantMsgID)
			}
		})
	}
}
