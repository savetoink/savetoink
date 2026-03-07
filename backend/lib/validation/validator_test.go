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

func TestValidateURLOnlyFormat(t *testing.T) {
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
			name:    "valid URL with port",
			url:     "https://example.com:8080/path",
			wantErr: false,
		},
		{
			name:    "private IP is allowed",
			url:     "http://192.168.1.1/admin",
			wantErr: false,
		},
		{
			name:    "localhost is allowed",
			url:     "http://localhost:8080",
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
			name:    "invalid URL format",
			url:     "://example.com",
			wantErr: true,
			errMsg:  "failed to parse URL",
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
			err := ValidateURLOnlyFormat(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURLOnlyFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateURLOnlyFormat() error = %v, expected to contain %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateURL_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "URL with IPv6 public address",
			url:     "https://[2001:db8::1]/path",
			wantErr: false,
		},
		{
			name:    "URL with IPv6 loopback",
			url:     "http://[::1]/path",
			wantErr: true,
			errMsg:  "private/internal network",
		},
		{
			name:    "URL with port",
			url:     "https://example.com:443/path",
			wantErr: false,
		},
		{
			name:    "URL with port 80",
			url:     "http://example.com:80/path",
			wantErr: false,
		},
		{
			name:    "URL with custom port",
			url:     "http://example.com:8080/path",
			wantErr: false,
		},

		{
			name:    "file scheme",
			url:     "file:///path/to/file",
			wantErr: true,
			errMsg:  "must use http or https scheme",
		},
		{
			name:    "invalid URL format - malformed",
			url:     "://example.com",
			wantErr: true,
			errMsg:  "failed to parse URL",
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

func TestValidateEmail_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "email with display name",
			email:   "John Doe <john@example.com>",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "email with angle brackets only",
			email:   "<user@example.com>",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "email with plus sign",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "email with hyphen in local part",
			email:   "user-name@example.com",
			wantErr: false,
		},
		{
			name:    "email with dot in local part",
			email:   "user.name@example.com",
			wantErr: false,
		},
		{
			name:    "email starting with dot",
			email:   ".user@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "email ending with dot",
			email:   "user.@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "email at minimum length",
			email:   "a@b.c",
			wantErr: false,
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

func TestValidateDeviceEmail_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "device email with plus sign",
			email:   "user+tag@kindle.com",
			wantErr: false,
		},
		{
			name:    "device email with subdomain",
			email:   "user@sub.kindle.com",
			wantErr: true,
			errMsg:  "must be a valid email ending with",
		},
		{
			name:    "case insensitive domain match",
			email:   "user@KINDLE.COM",
			wantErr: true,
			errMsg:  "must be a valid email ending with",
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

func TestIsPrivateHost(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		{
			name: "private IPv4 192.168.1.1",
			host: "192.168.1.1",
			want: true,
		},
		{
			name: "private IPv4 10.0.0.1",
			host: "10.0.0.1",
			want: true,
		},
		{
			name: "private IPv4 172.16.0.1",
			host: "172.16.0.1",
			want: true,
		},
		{
			name: "loopback IPv4 127.0.0.1",
			host: "127.0.0.1",
			want: true,
		},

		{
			name: "link-local IPv4 169.254.1.1",
			host: "169.254.1.1",
			want: true,
		},
		{
			name: "localhost",
			host: "localhost",
			want: true,
		},
		{
			name: "localhost with port 80",
			host: "localhost:80",
			want: true,
		},
		{
			name: "localhost with port 443",
			host: "localhost:443",
			want: true,
		},
		{
			name: "private IPv4 with port",
			host: "192.168.1.1:8080",
			want: true,
		},
		{
			name: "IPv6 loopback with brackets",
			host: "[::1]",
			want: true,
		},
		{
			name: "IPv6 loopback with brackets and port 80",
			host: "[::1]:80",
			want: true,
		},
		{
			name: "IPv6 loopback with brackets and port 443",
			host: "[::1]:443",
			want: true,
		},
		{
			name: "public IPv4 8.8.8.8",
			host: "8.8.8.8",
			want: false,
		},
		{
			name: "public IPv4 with port",
			host: "8.8.8.8:443",
			want: false,
		},
		{
			name: "public IPv6 2001:db8::1",
			host: "2001:db8::1",
			want: false,
		},
		{
			name: "public IPv6 with brackets",
			host: "[2001:db8::1]",
			want: false,
		},
		{
			name: "domain name",
			host: "example.com",
			want: false,
		},
		{
			name: "domain name with port",
			host: "example.com:8080",
			want: false,
		},
		{
			name: "invalid IP",
			host: "256.256.256.256",
			want: false,
		},
		{
			name: "empty host",
			host: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrivateHost(tt.host)
			if got != tt.want {
				t.Errorf("isPrivateHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrInvalidURL",
			err:      ErrInvalidURL,
			expected: "invalid URL",
		},
		{
			name:     "ErrInvalidEmail",
			err:      ErrInvalidEmail,
			expected: "invalid email address",
		},
		{
			name:     "ErrPrivateIPAddress",
			err:      ErrPrivateIPAddress,
			expected: "URL points to private/internal network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("%s.Error() = %v, want %v", tt.name, tt.err.Error(), tt.expected)
			}
		})
	}
}
