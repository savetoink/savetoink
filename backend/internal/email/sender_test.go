package email

import (
	"testing"

	"github.com/shaftoe/savetoink/backend/internal/model"
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
			expected: "[Save to Ink] This is a very long article title that definitely " +
				"exceeds the maximum length limit and",
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
