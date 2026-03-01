package mailjet

import (
	"context"
	"testing"

	"github.com/shaftoe/savetoink/internal/email"
	"github.com/shaftoe/savetoink/internal/model"
)

func TestNewSender(t *testing.T) {
	sender := NewSender("test-key", "test-secret", "test@example.com")
	if sender == nil {
		t.Fatal("NewSender returned nil")
	}

	if sender.apiKey != "test-key" {
		t.Error("Sender API key not set correctly")
	}
	if sender.apiSecret != "test-secret" {
		t.Error("Sender API secret not set correctly")
	}
	if sender.senderEmail != "test@example.com" {
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
			senderEmail: "test@example.com",
			wantErr:     false,
		},
		{
			name:        "missing api key",
			apiKey:      "",
			apiSecret:   "secret",
			senderEmail: "test@example.com",
			wantErr:     true,
		},
		{
			name:        "missing api secret",
			apiKey:      "key",
			apiSecret:   "",
			senderEmail: "test@example.com",
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
				Article: &model.Article{
					Title: "Test Article",
				},
				EPUBData:  []byte("test epub data"),
				DestEmail: "kindle@kindle.com",
			},
			wantErr: false,
		},
		{
			name: "missing device email",
			req: &email.Request{
				Article:   &model.Article{Title: "Test"},
				EPUBData:  []byte("data"),
				DestEmail: "",
			},
			wantErr: true,
		},
		{
			name: "missing epub data",
			req: &email.Request{
				Article:   &model.Article{Title: "Test"},
				EPUBData:  nil,
				DestEmail: "kindle@kindle.com",
			},
			wantErr: true,
		},
		{
			name: "missing article",
			req: &email.Request{
				Article:   nil,
				EPUBData:  []byte("data"),
				DestEmail: "kindle@kindle.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewSender("key", "secret", "test@example.com")
			err := sender.validateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRequest_EmptyArticleTitle(t *testing.T) {
	sender := NewSender("key", "secret", "test@example.com")
	req := &email.Request{
		Article:   &model.Article{Title: ""},
		EPUBData:  []byte("data"),
		DestEmail: "kindle@kindle.com",
	}
	err := sender.validateRequest(req)
	if err != nil {
		t.Errorf("validateRequest() unexpected error = %v", err)
	}
}

func TestValidateRequest_NonEmptyArticleTitle(t *testing.T) {
	sender := NewSender("key", "secret", "test@example.com")
	req := &email.Request{
		Article:   &model.Article{Title: "Test Article"},
		EPUBData:  []byte("data"),
		DestEmail: "kindle@kindle.com",
	}
	err := sender.validateRequest(req)
	if err != nil {
		t.Errorf("validateRequest() unexpected error = %v", err)
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
			senderEmail: "test@example.com",
			req: &email.Request{
				Article:   &model.Article{Title: "Test"},
				EPUBData:  []byte("data"),
				DestEmail: "kindle@kindle.com",
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing api secret in config",
			apiKey:      "key",
			apiSecret:   "",
			senderEmail: "test@example.com",
			req: &email.Request{
				Article:   &model.Article{Title: "Test"},
				EPUBData:  []byte("data"),
				DestEmail: "kindle@kindle.com",
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
				Article:   &model.Article{Title: "Test"},
				EPUBData:  []byte("data"),
				DestEmail: "kindle@kindle.com",
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing device email in request",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: "test@example.com",
			req: &email.Request{
				Article:   &model.Article{Title: "Test"},
				EPUBData:  []byte("data"),
				DestEmail: "",
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing epub data in request",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: "test@example.com",
			req: &email.Request{
				Article:   &model.Article{Title: "Test"},
				EPUBData:  nil,
				DestEmail: "kindle@kindle.com",
			},
			wantErr:    true,
			expectResp: nil,
		},
		{
			name:        "missing article in request",
			apiKey:      "key",
			apiSecret:   "secret",
			senderEmail: "test@example.com",
			req: &email.Request{
				Article:   nil,
				EPUBData:  []byte("data"),
				DestEmail: "kindle@kindle.com",
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
