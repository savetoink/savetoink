package validation

import (
	"strings"
	"testing"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/stretchr/testify/assert"
)

const (
	errCannotBeEmpty          = "cannot be empty"
	errInvalidEmailFormat     = "invalid email format"
	errPrivateInternalNetwork = "private/internal network"
	errValidEmailEndingWith   = "must be a valid email ending with"

	testLocalhost        = "localhost"
	testExceedsMaxLen    = "exceeds maximum length"
	testTagTech          = "tech"
	testTagEmpty         = "tag cannot be empty"
	testTagInvalidChars  = "can only contain letters, numbers, spaces, hyphens, and underscores"
	testTagProgramming   = "programming"
	testErrMsgHTTPScheme = "must use http or https scheme"
	testEmailUser        = "user@example.com"
	testNameEmptyEmail   = "empty email"
	testNameInvalidFmt   = "invalid format"
	testTagTechUpper     = "TECH"
	testTagTechMixed     = "Tech"
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
			errMsg:  errCannotBeEmpty,
		},
		{
			name:    "invalid scheme",
			url:     "ftp://example.com",
			wantErr: true,
			errMsg:  testErrMsgHTTPScheme,
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
			errMsg:  errPrivateInternalNetwork,
		},
		{
			name:    testLocalhost,
			url:     "http://localhost",
			wantErr: true,
			errMsg:  errPrivateInternalNetwork,
		},
		{
			name:    "loopback IP",
			url:     "http://127.0.0.1",
			wantErr: true,
			errMsg:  errPrivateInternalNetwork,
		},
		{
			name:    "URL too long",
			url:     "https://example.com/" + strings.Repeat("a", maxURLLength+100),
			wantErr: true,
			errMsg:  testExceedsMaxLen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateURL(tt.url)
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
			email:   testEmailUser,
			wantErr: false,
		},
		{
			name:    testNameEmptyEmail,
			email:   "",
			wantErr: true,
			errMsg:  errCannotBeEmpty,
		},
		{
			name:    testNameInvalidFmt,
			email:   testNotAnEmail,
			wantErr: true,
			errMsg:  errInvalidEmailFormat,
		},
		{
			name:    "email too long",
			email:   "user@" + strings.Repeat("a", 312) + ".com",
			wantErr: true,
			errMsg:  testExceedsMaxLen,
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
			email:   testEmailUser,
			wantErr: true,
			errMsg:  errValidEmailEndingWith,
		},
		{
			name:    testNameEmptyEmail,
			email:   "",
			wantErr: true,
			errMsg:  errCannotBeEmpty,
		},
		{
			name:    testNameInvalidFmt,
			email:   testNotAnEmail,
			wantErr: true,
			errMsg:  errInvalidEmailFormat,
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
			errMsg:  errPrivateInternalNetwork,
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
			errMsg:  testErrMsgHTTPScheme,
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
			_, err := ValidateURL(tt.url)
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
			errMsg:  errInvalidEmailFormat,
		},
		{
			name:    "email with angle brackets only",
			email:   "<user@example.com>",
			wantErr: true,
			errMsg:  errInvalidEmailFormat,
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
			errMsg:  errInvalidEmailFormat,
		},
		{
			name:    "email ending with dot",
			email:   "user.@example.com",
			wantErr: true,
			errMsg:  errInvalidEmailFormat,
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
			errMsg:  errValidEmailEndingWith,
		},
		{
			name:    "case insensitive domain match",
			email:   "user@KINDLE.COM",
			wantErr: true,
			errMsg:  errValidEmailEndingWith,
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
			name: testLocalhost,
			host: testLocalhost,
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
		{
			name:     "ErrInvalidTag",
			err:      ErrInvalidTag,
			expected: "invalid tag",
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

func TestValidateTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid single word tag",
			tag:     testTagTech,
			wantErr: false,
		},
		{
			name:    "valid tag with hyphen",
			tag:     "tech-news",
			wantErr: false,
		},
		{
			name:    "valid tag with underscore",
			tag:     "tech_news",
			wantErr: false,
		},
		{
			name:    "valid tag with space",
			tag:     "tech news",
			wantErr: false,
		},
		{
			name:    "valid tag with numbers",
			tag:     "python3",
			wantErr: false,
		},
		{
			name:    "valid tag mixed case (gets lowercased)",
			tag:     "TechNews",
			wantErr: false,
		},
		{
			name:    "valid tag with leading/trailing spaces",
			tag:     "  tech  ",
			wantErr: false,
		},
		{
			name:    "empty tag after trim",
			tag:     "   ",
			wantErr: true,
			errMsg:  testTagEmpty,
		},
		{
			name:    "empty tag",
			tag:     "",
			wantErr: true,
			errMsg:  testTagEmpty,
		},
		{
			name:    "tag exceeds max length",
			tag:     strings.Repeat("a", consts.MaxTagLength+1),
			wantErr: true,
			errMsg:  testExceedsMaxLen,
		},
		{
			name:    "tag at max length",
			tag:     strings.Repeat("a", consts.MaxTagLength),
			wantErr: false,
		},
		{
			name:    "tag with special characters",
			tag:     "tech@news",
			wantErr: true,
			errMsg:  testTagInvalidChars,
		},
		{
			name:    "tag with parentheses",
			tag:     "tech(news)",
			wantErr: true,
			errMsg:  testTagInvalidChars,
		},
		{
			name:    "tag with slash",
			tag:     "tech/news",
			wantErr: true,
			errMsg:  testTagInvalidChars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateTag(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateTag() error = %v, expected to contain %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr {
				// Check that result is lowercased
				if result != strings.ToLower(result) {
					t.Errorf("ValidateTag() result = %v, expected to be lowercased", result)
				}
			}
		})
	}
}

func TestValidateTags(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		wantErr bool
		errMsg  string
		wantLen int
	}{
		{
			name:    "valid single tag",
			tags:    []string{testTagTech},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "valid multiple tags",
			tags:    []string{testTagTech, testTagProgramming, "golang"},
			wantErr: false,
			wantLen: 3,
		},
		{
			name:    "empty slice is valid",
			tags:    []string{},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "nil slice is valid",
			tags:    nil,
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "duplicate tags are removed",
			tags:    []string{testTagTech, testTagTech, testTagProgramming},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "case duplicates are removed",
			tags:    []string{testTagTechMixed, testTagTech, testTagTechUpper},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "tags are normalized (trimmed and lowercased)",
			tags:    []string{"  Tech  ", "PROGRAMMING", "  GoLang  "},
			wantErr: false,
			wantLen: 3,
		},
		{
			name:    "too many tags",
			tags:    make([]string, consts.MaxTagsPerArticle+1),
			wantErr: true,
			errMsg:  "invalid tag: maximum 10 tags allowed per article",
			wantLen: 0,
		},
		{
			name:    "max allowed tags",
			tags:    []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7", "tag8", "tag9", "tag10"},
			wantErr: false,
			wantLen: consts.MaxTagsPerArticle,
		},
		{
			name:    "one invalid tag in list",
			tags:    []string{testTagTech, "invalid@tag", testTagProgramming},
			wantErr: true,
			errMsg:  testTagInvalidChars,
		},
		{
			name:    "empty tag in list",
			tags:    []string{testTagTech, "", testTagProgramming},
			wantErr: true,
			errMsg:  testTagEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateTags(tt.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateTags() error = %v, expected to contain %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr {
				if len(result) != tt.wantLen {
					t.Errorf("ValidateTags() result length = %v, want %v", len(result), tt.wantLen)
				}
				// Check that all tags are lowercased
				for _, tag := range result {
					if tag != strings.ToLower(tag) {
						t.Errorf("ValidateTags() result contains non-lowercased tag: %v", tag)
					}
				}
			}
		})
	}
}

func TestValidateTags_Deduplication(t *testing.T) {
	tags := []string{testTagTechMixed, testTagTech, testTagTechUpper, "Programming", testTagProgramming, "Go"}
	result, err := ValidateTags(tags)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.ElementsMatch(t, []string{testTagTech, testTagProgramming, "go"}, result)
}

func TestValidateTags_MaxTagLength(t *testing.T) {
	// Create a tag that's exactly at max length
	validTag := strings.Repeat("a", consts.MaxTagLength)
	tags := []string{validTag}
	result, err := ValidateTags(tags)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, validTag, result[0])

	// Create a tag that's one character over max length
	invalidTag := strings.Repeat("b", consts.MaxTagLength+1)
	tags = []string{invalidTag}
	_, err = ValidateTags(tags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), testExceedsMaxLen)
}
