package email

import (
	"context"
	"testing"
)

func TestBuildSubject(t *testing.T) {
	tests := []struct {
		name         string
		articleTitle string
		expected     string
	}{
		{
			name:         "article title",
			articleTitle: "Article Title",
			expected:     "[Save to Ink] Article Title",
		},
		{
			name:         "title with leading/trailing spaces",
			articleTitle: "  Article Title  ",
			expected:     "[Save to Ink] Article Title",
		},
		{
			name: "long title",
			articleTitle: "This is a very long article title that definitely exceeds " +
				"the maximum length limit and should be properly truncated to fit",
			expected: "[Save to Ink] This is a very long article title that definitely exceeds the maximum length limit and",
		},
		{
			name:     "empty title",
			expected: "[Save to Ink] Document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSubject(tt.articleTitle)
			if got != tt.expected {
				t.Errorf("BuildSubject() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name      string
		epubData  []byte
		destEmail string
		appURL    string
	}{
		{
			name:      "valid request with all fields",
			epubData:  []byte("epub content"),
			destEmail: "user@kindle.com",
			appURL:    "https://saveto.ink",
		},
		{
			name:      "request with empty epub data",
			epubData:  []byte{},
			destEmail: "user@free.kindle.com",
			appURL:    "https://example.com",
		},
		{
			name:      "request with large epub data",
			epubData:  make([]byte, 1024*1024),
			destEmail: "user@kindle.com",
			appURL:    "https://saveto.ink",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRequest(tt.epubData, tt.destEmail, tt.appURL)

			if got.EPUBData == nil {
				t.Errorf("NewRequest().EPUBData should not be nil")
			} else if len(got.EPUBData) != len(tt.epubData) {
				t.Errorf("NewRequest().EPUBData length = %d, want %d", len(got.EPUBData), len(tt.epubData))
			}
			if got.DestEmail != tt.destEmail {
				t.Errorf("NewRequest().DestEmail = %v, want %v", got.DestEmail, tt.destEmail)
			}
			if got.Body == "" {
				t.Errorf("NewRequest().Body should not be empty")
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *Request
		wantErr bool
	}{
		{
			name: "valid request",
			req: &Request{
				EPUBData:  []byte("test epub data"),
				DestEmail: "kindle@kindle.com",
				Body:      "email body",
				Subject:   "Test Subject",
			},
			wantErr: false,
		},
		{
			name: "missing device email",
			req: &Request{
				EPUBData:  []byte("data"),
				DestEmail: "",
				Body:      "email body",
				Subject:   "Test Subject",
			},
			wantErr: true,
		},
		{
			name: "missing epub data",
			req: &Request{
				EPUBData:  nil,
				DestEmail: "kindle@kindle.com",
				Body:      "email body",
				Subject:   "Test Subject",
			},
			wantErr: true,
		},
		{
			name: "missing body",
			req: &Request{
				EPUBData:  []byte("data"),
				DestEmail: "kindle@kindle.com",
				Body:      "",
				Subject:   "Test Subject",
			},
			wantErr: true,
		},
		{
			name: "empty epub data",
			req: &Request{
				EPUBData:  []byte{},
				DestEmail: "kindle@kindle.com",
				Body:      "email body",
				Subject:   "Test Subject",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := ValidateRequest(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "simple title",
			title:    "Test Article",
			expected: "Test Article.epub",
		},
		{
			name:     "title with colon",
			title:    "Test: Article",
			expected: "Test Article.epub",
		},
		{
			name:     "title with slash",
			title:    "Test/Article",
			expected: "Test Article.epub",
		},
		{
			name:     "title with backslash",
			title:    "Test\\Article",
			expected: "Test Article.epub",
		},
		{
			name:     "title with leading spaces",
			title:    "   Test Article",
			expected: "Test Article.epub",
		},
		{
			name:     "title with trailing spaces",
			title:    "Test Article   ",
			expected: "Test Article.epub",
		},
		{
			name:     "title with quotes",
			title:    "\"Test Article\"",
			expected: "Test Article.epub",
		},
		{
			name:     "title with multiple dots",
			title:    "Test...Article",
			expected: "Test Article.epub",
		},
		{
			name:     "title with unicode and emoji",
			title:    "Test 中文 😊",
			expected: "Test.epub",
		},
		{
			name:     "title with mixed case",
			title:    "TeSt ArTiClE",
			expected: "TeSt ArTiClE.epub",
		},
		{
			name:     "title with numbers",
			title:    "Test Article 2024",
			expected: "Test Article 2024.epub",
		},
		{
			name:     "title with brackets",
			title:    "Test (Article) [2024]",
			expected: "Test (Article) [2024].epub",
		},
		{
			name:     "title with quote",
			title:    "Test Article's title",
			expected: "Test Article's title.epub",
		},
		{
			name:     "empty title",
			title:    "",
			expected: "article.epub",
		},
		{
			name: "title longer than 90 chars",
			title: "This is a very long article title that definitely exceeds " +
				"maximum length limit and should be properly truncated to fit",
			expected: "This is a very long article title that definitely exceeds maximum length limit and should.epub",
		},
		{
			name:     "title with only special characters",
			title:    "!!!***###",
			expected: "article.epub",
		},
		{
			name:     "title with consecutive special chars",
			title:    "Test---Article",
			expected: "Test Article.epub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.title)
			if got != tt.expected {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.expected)
			}
		})
	}
}
