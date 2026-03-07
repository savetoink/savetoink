package email

import (
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
