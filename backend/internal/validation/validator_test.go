package validation

import (
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid HTTPS URL",
			url:     "https://example.com/article",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "invalid scheme",
			url:     "ftp://example.com",
			wantErr: true,
			errMsg:  "must use http or https scheme",
		},
		{
			name:    "missing host",
			url:     "https://",
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name:    "private IP",
			url:     "http://192.168.1.1",
			wantErr: true,
			errMsg:  "private/internal network",
		},
		{
			name:    "localhost",
			url:     "http://localhost",
			wantErr: true,
			errMsg:  "private/internal network",
		},
		{
			name:    "loopback IP",
			url:     "http://127.0.0.1",
			wantErr: true,
			errMsg:  "private/internal network",
		},
		{
			name:    "URL too long",
			url:     "https://example.com/" + strings.Repeat("a", maxURLLength+100),
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateURL() error = %v, expected to contain %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "invalid format",
			email:   "not-an-email",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "email too long",
			email:   "user@" + strings.Repeat("a", 312) + ".com",
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateEmail() error = %v, expected to contain %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateDeviceEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid kindle.com email",
			email:   "user@kindle.com",
			wantErr: false,
		},
		{
			name:    "valid free.kindle.com email",
			email:   "user@free.kindle.com",
			wantErr: false,
		},
		{
			name:    "invalid domain",
			email:   "user@example.com",
			wantErr: true,
			errMsg:  "must be a valid email ending with",
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "invalid format",
			email:   "not-an-email",
			wantErr: true,
			errMsg:  "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeviceEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeviceEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateDeviceEmail() error = %v, expected to contain %v", err, tt.errMsg)
				}
			}
		})
	}
}
