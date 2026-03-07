package email

import (
	"testing"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		article  *model.Article
		expected string
	}{
		{
			name: "article with title",
			article: &model.Article{
				Title: "Test Article",
			},
			expected: "Test Article.epub",
		},
		{
			name: "article with special characters in title",
			article: &model.Article{
				Title: "Test Article: What's New?",
			},
			expected: "Test Article Whats New.epub",
		},
		{
			name:     "article without title",
			article:  &model.Article{},
			expected: "article.epub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilename(tt.article)
			if got != tt.expected {
				t.Errorf("GenerateFilename() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple filename",
			input:    "test-file",
			expected: "test-file",
		},
		{
			name:     "filename with special chars",
			input:    "test: file? what!",
			expected: "test file what",
		},
		{
			name:     "empty filename",
			input:    "",
			expected: "article",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "article",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.expected)
			}
		})
	}
}

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
		article   *model.Article
		epubData  []byte
		destEmail string
		appURL    string
	}{
		{
			name: "valid request with all fields",
			article: &model.Article{
				ID:    "article-123",
				Title: "Test Article",
			},
			epubData:  []byte("epub content"),
			destEmail: "user@kindle.com",
			appURL:    "https://saveto.ink",
		},
		{
			name: "request with empty article",
			article: &model.Article{
				ID:    "",
				Title: "",
			},
			epubData:  []byte{},
			destEmail: "user@free.kindle.com",
			appURL:    "https://example.com",
		},
		{
			name:      "request with nil article",
			article:   nil,
			epubData:  []byte("data"),
			destEmail: "user@kindle.com",
			appURL:    "http://localhost:3000",
		},
		{
			name: "request with large epub data",
			article: &model.Article{
				ID:    "article-456",
				Title: "Large Article",
			},
			epubData:  make([]byte, 1024*1024),
			destEmail: "user@kindle.com",
			appURL:    "https://saveto.ink",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRequest(tt.article, tt.epubData, tt.destEmail, tt.appURL)

			if got.Article != tt.article {
				t.Errorf("NewRequest().Article = %v, want %v", got.Article, tt.article)
			}
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
