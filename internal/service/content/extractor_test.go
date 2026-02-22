package content

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewExtractor(t *testing.T) {
	extractor := NewExtractor()
	if extractor == nil {
		t.Fatal("NewExtractor returned nil")
	}
	if extractor.client == nil {
		t.Error("Extractor client is nil")
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https url",
			url:     "https://example.com/article",
			wantErr: false,
		},
		{
			name:    "valid http url",
			url:     "http://example.com/article",
			wantErr: false,
		},
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			url:     "ftp://example.com/article",
			wantErr: true,
		},
		{
			name:    "no scheme",
			url:     "example.com/article",
			wantErr: true,
		},
		{
			name:    "no host",
			url:     "https://",
			wantErr: true,
		},
		{
			name:    "malformed url",
			url:     "://example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractFromURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		responseCode  int
		contentType   string
		html          string
		wantErr       bool
		expectedTitle string
	}{
		{
			name:         "successful extraction",
			responseCode: http.StatusOK,
			contentType:  "text/html",
			html: `<!DOCTYPE html>` +
				`<html><head><title>Test Article</title></head>` +
				`<body><article><h1>Test Article</h1>` +
				`<p>Content here</p></article></body></html>`,
			wantErr:       false,
			expectedTitle: "Test Article",
		},
		{
			name:         "non-html content type",
			responseCode: http.StatusOK,
			contentType:  "application/json",
			html:         `{"title": "test"}`,
			wantErr:      true,
		},
		{
			name:         "404 response",
			responseCode: http.StatusNotFound,
			contentType:  "text/html",
			html:         `<html><body>Not Found</body></html>`,
			wantErr:      true,
		},
		{
			name:         "500 response",
			responseCode: http.StatusInternalServerError,
			contentType:  "text/html",
			html:         `<html><body>Internal Error</body></html>`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.responseCode)
				_, _ = w.Write([]byte(tt.html))
			}))
			defer server.Close()

			extractor := NewExtractor()
			article, err := extractor.ExtractFromURL(ctx, server.URL)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if article == nil {
					t.Fatal("Expected article but got nil")
				}
				if tt.expectedTitle != "" && article.Title != tt.expectedTitle {
					t.Errorf("Expected title %s, got %s", tt.expectedTitle, article.Title)
				}
				if article.URL != server.URL {
					t.Errorf("Expected URL %s, got %s", server.URL, article.URL)
				}
			}
		})
	}
}

func TestExtractFromURLWithContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Test</body></html>"))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	extractor := NewExtractor()
	_, err := extractor.ExtractFromURL(ctx, server.URL)
	if err == nil {
		t.Error("Expected error due to canceled context, got nil")
	}
}

func TestExtractFromURLInvalidURL(t *testing.T) {
	ctx := context.Background()
	extractor := NewExtractor()

	tests := []struct {
		name string
		url  string
	}{
		{"empty url", ""},
		{"invalid scheme", "ftp://example.com"},
		{"no host", "https://"},
		{"malformed", "://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractor.ExtractFromURL(ctx, tt.url)
			if err == nil {
				t.Errorf("Expected error for URL %s, got nil", tt.url)
			}
		})
	}
}

func TestArticleFields(t *testing.T) {
	ctx := context.Background()
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Test Article Title</title>
	<meta name="description" content="This is a test excerpt">
	<meta name="author" content="John Doe">
	<meta property="og:image" content="https://example.com/image.jpg">
	<meta name="date" content="2024-01-15">
</head>
<body>
	<article>
		<h1>Test Article Title</h1>
		<p>This is the main content of the article with multiple paragraphs.</p>
		<p>Second paragraph content.</p>
	</article>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	extractor := NewExtractor()
	article, err := extractor.ExtractFromURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("ExtractFromURL() error = %v", err)
	}

	if article == nil {
		t.Fatal("Expected article but got nil")
	}

	if article.Title == "" {
		t.Error("Expected title to be set")
	}

	if article.Content == "" {
		t.Error("Expected content to be set")
	}

	if article.URL != server.URL {
		t.Errorf("Expected URL to be %s, got %s", server.URL, article.URL)
	}

	if article.Excerpt == "" {
		t.Error("Expected excerpt to be set")
	}

	if article.Author == "" {
		t.Error("Expected author to be set")
	}

	if article.WordCount == 0 {
		t.Error("Expected word count to be set")
	}

	if article.ReadingTimeMinutes == 0 {
		t.Error("Expected reading time to be set")
	}

	id, err := ArticleIDFromURL(server.URL)
	if err != nil {
		t.Fatalf("ArticleIDFromURL() error = %v", err)
	}
	article.ID = id
	if article.ID == "" {
		t.Error("Expected ID to be set")
	}
}
