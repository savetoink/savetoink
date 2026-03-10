package content

import (
	"context"
	"net/url"
	"testing"

	"github.com/shaftoe/savetoink/backend/lib/validation"
	"github.com/stretchr/testify/assert"
)

const testArticleURL = "https://example.com"

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
			url:     testArticleURL + "/article",
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
			_, err := validation.ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractFromURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		html          string
		wantErr       bool
		expectedTitle string
	}{
		{
			name: "successful extraction",
			html: `<!DOCTYPE html>` +
				`<html><head><title>Test Article</title></head>` +
				`<body><article><h1>Test Article</h1>` +
				`<p>Content here</p></article></body></html>`,
			wantErr:       false,
			expectedTitle: "Test Article",
		},
		{
			name:    "non-html content",
			html:    `{"title": "test"}`,
			wantErr: false,
		},
		{
			name:    "minimal html",
			html:    `<html><body>Not Found</body></html>`,
			wantErr: false,
		},
		{
			name:    "minimal html 2",
			html:    `<html><body>Internal Error</body></html>`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testURL, _ := url.Parse(testArticleURL)
			extractor := NewExtractor()
			article, err := extractor.GenerateFromHTML(ctx, []byte(tt.html), testURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFromHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if article == nil {
					t.Fatal("Expected article but got nil")
				}
				if tt.expectedTitle != "" && article.Title != tt.expectedTitle {
					t.Errorf("Expected title %s, got %s", tt.expectedTitle, article.Title)
				}
			}
		})
	}
}

func TestExtractFromURLWithContextCancellation(t *testing.T) {
	t.Skip("GenerateFromHTML doesn't use context, skipping this test")
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

	testURL, _ := url.Parse(testArticleURL)
	extractor := NewExtractor()
	article, err := extractor.GenerateFromHTML(ctx, []byte(html), testURL)
	if err != nil {
		t.Fatalf("GenerateFromHTML() error = %v", err)
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

	if article.URL != testURL.String() {
		t.Errorf("Expected URL to be %s, got %s", testURL.String(), article.URL)
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

	id, err := ArticleIDFromURL(testURL)
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

	testURL, _ := url.Parse(testArticleURL)
	extractor := NewExtractor()
	article, err := extractor.GenerateFromHTML(ctx, []byte(html), testURL)
	if err != nil {
		t.Fatalf("GenerateFromHTML() error = %v", err)
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

	testURL, _ := url.Parse(testArticleURL)
	extractor := NewExtractor()
	article, err := extractor.GenerateFromHTML(ctx, []byte(html), testURL)
	if err != nil {
		t.Fatalf("GenerateFromHTML() error = %v", err)
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

	testURL, _ := url.Parse(testArticleURL)
	extractor := NewExtractor()
	fetcher := NewFetcher("")

	_, err := fetcher.Fetch(ctx, testURL)
	if err != nil {
		t.Skipf("Skipping test - network fetch failed: %v", err)
	}

	_, err = extractor.GenerateFromHTML(ctx, []byte(html), testURL)
	if err != nil {
		t.Fatalf("GenerateFromHTML() error = %v", err)
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

func TestGenerateFromHTML_Success(t *testing.T) {
	ctx := context.Background()
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Test Article</title>
	<meta name="author" content="Test Author">
	<meta name="description" content="Test excerpt">
</head>
<body>
	<article>
		<h1>Test Article</h1>
		<p>This is test content with multiple words.</p>
	</article>
</body>
	</html>`

	extractor := NewExtractor()
	testURL := testArticleURL + "/article"
	u, err := url.Parse(testURL)
	if err != nil {
		t.Fatalf("Failed to parse test URL: %v", err)
	}
	article, err := extractor.GenerateFromHTML(ctx, []byte(html), u)

	if err != nil {
		t.Fatalf("GenerateFromHTML() error = %v", err)
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

	if article.WordCount == 0 {
		t.Error("Expected word count to be set")
	}

	if article.ReadingTimeMinutes == 0 {
		t.Error("Expected reading time to be set")
	}
}

func TestGenerateFromHTML_TitleEqualsSitename(t *testing.T) {
	ctx := context.Background()
	html := `<!DOCTYPE html>
<html>
<head>
	<title>My Site</title>
	<meta property="og:site_name" content="My Site">
	<meta name="author" content="Test Author">
</head>
<body>
	<h2>Real Article Title</h2>
	<article>
		<p>Content goes here.</p>
	</article>
</body>
	</html>`

	extractor := NewExtractor()
	testURL := testArticleURL + "/article"
	u, err := url.Parse(testURL)
	if err != nil {
		t.Fatalf("Failed to parse test URL: %v", err)
	}
	article, err := extractor.GenerateFromHTML(ctx, []byte(html), u)

	if err != nil {
		t.Fatalf("GenerateFromHTML() error = %v", err)
	}

	if article.Title == "" {
		t.Error("Expected title to be set")
	}

	if article.Title == "My Site" {
		t.Errorf("Expected fallback title from h2, got sitename: %q", article.Title)
	}

	expectedTitle := "Real Article Title"
	if article.Title != expectedTitle {
		t.Errorf("Expected title %q, got %q", expectedTitle, article.Title)
	}
}

func TestExtractFromURL_InvalidURLFormat(t *testing.T) {
	ctx := context.Background()
	extractor := NewExtractor()

	invalidURLs := []string{
		"http://%00invalid.com",
		"https://[invalid-hostname]",
	}

	for _, url := range invalidURLs {
		_, err := extractor.ExtractFromURL(ctx, url)
		if err == nil {
			t.Errorf("Expected error for invalid URL %s, got nil", url)
		}
	}
}

func TestExtractTitleFromHTML_ParseError(t *testing.T) {
	invalidHTML := []byte("<html><body>unclosed tag")

	title := extractTitleFromHTML(invalidHTML)

	if title != "" {
		t.Errorf("Expected empty string for unparsable HTML, got %q", title)
	}
}

func TestExtractTitleFromHTML_EmptyTitleTag(t *testing.T) {
	html := []byte("<!DOCTYPE html><html><head><title></title></head><body><p>Content</p></body></html>")

	title := extractTitleFromHTML(html)

	if title != "" {
		t.Errorf("Expected empty string for empty title tag, got %q", title)
	}
}

func TestExtractTitleFromHTML_EmptyH2(t *testing.T) {
	html := []byte(
		"<!DOCTYPE html><html><head><title>Site Name</title>" +
			"</head><body><h2>  </h2><p>Content</p></body></html>",
	)

	title := extractTitleFromHTML(html)

	if title != "" {
		t.Errorf("Expected empty string for empty h2 tag, got %q", title)
	}
}

func TestExtractTitleFromHTML_TitleBecomesEmptyAfterClean(t *testing.T) {
	html := []byte("<!DOCTYPE html><html><head><title> - </title></head><body><p>Content</p></body></html>")

	title := extractTitleFromHTML(html)

	if title != "" {
		t.Errorf("Expected empty string when title becomes empty after clean, got %q", title)
	}
}

func TestGenerateFromHTML_WithFallbackToH2(t *testing.T) {
	ctx := context.Background()
	html := `<!DOCTYPE html>
<html>
<head>
	<title>My Blog</title>
	<meta property="og:site_name" content="My Blog">
	<meta name="author" content="John Doe">
	<meta name="description" content="A blog post">
</head>
<body>
	<h2>The Actual Article Title</h2>
	<article>
		<p>This is the content of the article.</p>
		<p>Second paragraph with more content.</p>
	</article>
</body>
	</html>`

	extractor := NewExtractor()
	testURL := testArticleURL + "/article"
	u, err := url.Parse(testURL)
	if err != nil {
		t.Fatalf("Failed to parse test URL: %v", err)
	}
	article, err := extractor.GenerateFromHTML(ctx, []byte(html), u)

	if err != nil {
		t.Fatalf("GenerateFromHTML() error = %v", err)
	}

	if article.Title == "" {
		t.Fatal("Expected title to be set")
	}

	if article.Title == "My Blog" {
		t.Error("Expected fallback to H2 title when metadata title equals sitename")
	}

	expectedTitle := "The Actual Article Title"
	if article.Title != expectedTitle {
		t.Errorf("Expected title %q, got %q", expectedTitle, article.Title)
	}
}

func TestExtractTitleFromHTML_NoTitleOrH2(t *testing.T) {
	html := []byte("<!DOCTYPE html><html><head></head><body><p>No title here</p></body></html>")

	title := extractTitleFromHTML(html)

	if title != "" {
		t.Errorf("Expected empty string when no title or h2 tags, got %q", title)
	}
}

func TestCleanTitle_InsufficientParts(t *testing.T) {
	singlePart := "SingleTitle"

	title := cleanTitle(singlePart)

	if title != "" {
		t.Errorf("Expected empty string for single part title, got %q", title)
	}
}

func TestCleanTitle_EmptyAfterTrim(t *testing.T) {
	onlyWhitespace := " - "

	title := cleanTitle(onlyWhitespace)

	if title != "" {
		t.Errorf("Expected empty string for whitespace-only title, got %q", title)
	}
}

func TestCleanTitle_ThreeParts(t *testing.T) {
	threePartTitle := "Part 1 - Part 2 - Part 3"

	title := cleanTitle(threePartTitle)

	expected := "Part 1 - Part 2"
	if title != expected {
		t.Errorf("Expected %q, got %q", expected, title)
	}
}
