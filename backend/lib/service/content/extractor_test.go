package content

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewExtractor(t *testing.T) {
	extractor := NewExtractor()
	if extractor == nil {
		t.Fatal("NewExtractor returned nil")
	}
}

func TestNewFetcher(t *testing.T) {
	fetcher := NewFetcher("")
	if fetcher == nil {
		t.Fatal("NewFetcher returned nil")
	}
	if fetcher.client == nil {
		t.Error("Fetcher client is nil")
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

func TestTitleExtractionFromTitleTag(t *testing.T) {
	ctx := context.Background()
	html := `<!DOCTYPE html>
<html>
<head>
	<title>First run the tests - Agentic Engineering Patterns - Simon Willison's Weblog</title>
	<meta name="author" content="Simon Willison">
	<meta property="og:site_name" content="Simon Willison's Weblog">
</head>
<body>
	<h1><a href="/">Simon Willison's Weblog</a></h1>
	<h2 class="archive-h2">First run the tests</h2>
	<article>
		<p>This is test content.</p>
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

	if article.Title == "" {
		t.Error("Expected title to be set")
	}

	if article.Title == "Simon Willison's Weblog" {
		t.Errorf("Title should not be equal to sitename. Got: %q", article.Title)
	}

	expectedTitle := "First run the tests - Agentic Engineering Patterns"
	if article.Title != expectedTitle {
		t.Errorf("Expected title %q, got %q", expectedTitle, article.Title)
	}
}

func TestTitleExtractionFromH2(t *testing.T) {
	ctx := context.Background()
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Simon Willison's Weblog</title>
	<meta name="author" content="Simon Willison">
	<meta property="og:site_name" content="Simon Willison's Weblog">
</head>
<body>
	<h1><a href="/">Simon Willison's Weblog</a></h1>
	<h2>First run the tests</h2>
	<article>
		<p>This is test content.</p>
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

	expectedTitle := "First run the tests"
	if article.Title != expectedTitle {
		t.Errorf("Expected title %q, got %q", expectedTitle, article.Title)
	}
}

func TestUserAgentHeader(t *testing.T) {
	ctx := context.Background()
	html := "<!DOCTYPE html><html><head><title>Test</title></head>" +
		"<body><article><h1>Test</h1><p>Content</p></article></body></html>"

	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	extractor := NewExtractor()
	_, err := extractor.ExtractFromURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("ExtractFromURL() error = %v", err)
	}

	if receivedUA == "" {
		t.Error("Expected User-Agent header to be set")
	}

	if !strings.HasPrefix(receivedUA, "Mozilla/5.0") {
		t.Errorf("Expected User-Agent to start with 'Mozilla/5.0', got: %s", receivedUA)
	}
}

func TestValidateHTMLContent_ValidHTML(t *testing.T) {
	validHTML := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><p>Content</p></body>
</html>`

	err := validateHTMLContent([]byte(validHTML))
	assert.NoError(t, err)
}

func TestValidateHTMLContent_HTML5NoDoctype(t *testing.T) {
	html5NoDoctype := `<html>
<head><title>Test</title></head>
<body><p>Content</p></body>
</html>`

	err := validateHTMLContent([]byte(html5NoDoctype))
	assert.NoError(t, err)
}

func TestValidateHTMLContent_XHTML(t *testing.T) {
	xhtml := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN"
"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Test</title></head>
<body><p>Content</p></body>
</html>`

	err := validateHTMLContent([]byte(xhtml))
	assert.NoError(t, err)
}

func TestValidateHTMLContent_ErrorPattern(t *testing.T) {
	pattern := "This website is using a security service to protect itself " +
		"from online attacks. The action you just performed triggered the " +
		"security solution."
	errorPatternHTML := `<html>
<head><title>Error</title></head>
<body>
<p>` + pattern + `</p>
</body>
</html>`

	err := validateHTMLContent([]byte(errorPatternHTML))
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "error page")
	}
}

func TestValidateHTMLContent_MissingHTMLTag(t *testing.T) {
	noHTMLTag := `<head><title>Test</title></head>
<body><p>Content</p></body>`

	err := validateHTMLContent([]byte(noHTMLTag))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not appear to be valid HTML")
}

func TestValidateHTMLContent_EmptyContent(t *testing.T) {
	emptyContent := []byte("")

	err := validateHTMLContent(emptyContent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not appear to be valid HTML")
}

func TestValidateHTMLContent_PlainText(t *testing.T) {
	plainText := []byte("This is just plain text without any HTML tags.")

	err := validateHTMLContent(plainText)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not appear to be valid HTML")
}

func TestValidateHTMLContent_JSONContent(t *testing.T) {
	jsonContent := []byte(`{"title": "Test Article", "content": "Some content"}`)

	err := validateHTMLContent(jsonContent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not appear to be valid HTML")
}

func TestValidateHTMLContent_MinimalHTML(t *testing.T) {
	minimalHTML := `<!DOCTYPE html><html><body>Minimal</body></html>`

	err := validateHTMLContent([]byte(minimalHTML))
	assert.NoError(t, err)
}

func TestValidateHTMLContent_HTMLWithScript(t *testing.T) {
	htmlWithScript := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<script>alert('test');</script>
<p>Content</p>
</body>
</html>`

	err := validateHTMLContent([]byte(htmlWithScript))
	assert.NoError(t, err)
}

func TestValidateHTMLContent_HTMLWithStyle(t *testing.T) {
	htmlWithStyle := `<!DOCTYPE html>
<html>
<head>
<style>body { color: red; }</style>
</head>
<body><p>Content</p></body>
</html>`

	err := validateHTMLContent([]byte(htmlWithStyle))
	assert.NoError(t, err)
}
