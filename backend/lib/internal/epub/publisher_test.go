package epub

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

const (
	testArticleTitle   = "Test Article"
	testArticleContent = "<p>This is test content</p>"
	testArticleAuthor  = "Test Author"
	testExampleSite    = "Example Site"
	testExampleDomain  = "example.com"
	testTitle          = "Test Title"
)

func TestNewPublisher(t *testing.T) {
	pub := NewPublisher()
	if pub == nil {
		t.Fatal("NewPublisher returned nil")
	}
}

func TestNewPublisher_WithMemoryStorage(t *testing.T) {
	pub := NewPublisher(WithMemoryStorage())
	if pub == nil {
		t.Fatal("NewPublisher returned nil")
		return
	}
	if !pub.useMemoryStorage {
		t.Error("NewPublisher WithMemoryStorage option did not set useMemoryStorage flag")
	}
}

func TestGenerate_WithMemoryStorage(t *testing.T) {
	pub := NewPublisher(WithMemoryStorage())
	article := &model.Article{
		Title:   testArticleTitle,
		Content: testArticleContent,
		Author:  testArticleAuthor,
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() with memory storage unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	data, err := io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}

	if len(data) == 0 {
		t.Error("Generate() expected non-empty data")
	}
}

func TestGenerate_Success(t *testing.T) {
	pub := NewPublisher()
	article := &model.Article{
		Title:   testArticleTitle,
		Content: testArticleContent,
		Author:  testArticleAuthor,
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	data, err := io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}

	if len(data) == 0 {
		t.Error("Generate() expected non-empty data")
	}
}

func TestGenerate_WithMetadata(t *testing.T) {
	pub := NewPublisher()
	publishedAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	article := &model.Article{
		Title:              testArticleTitle,
		Content:            testArticleContent,
		Author:             testArticleAuthor,
		SiteName:           testExampleSite,
		SourceDomain:       testExampleDomain,
		ReadingTimeMinutes: 5,
		PublishedAt:        &publishedAt,
		ContentType:        "article",
		CreatedAt:          time.Now().UTC(),
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	data, err := io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}

	if len(data) == 0 {
		t.Error("Generate() expected non-empty data")
	}
}

func TestGenerate_EmptyTitle(t *testing.T) {
	pub := NewPublisher()
	article := &model.Article{
		Title:   "",
		Content: testArticleContent,
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	_, err = io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}
}

func TestGenerate_EmptyContent(t *testing.T) {
	pub := NewPublisher()
	article := &model.Article{
		Title:   testArticleTitle,
		Content: "",
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	_, err = io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}
}

func TestGenerate_WithLanguage(t *testing.T) {
	pub := NewPublisher()
	article := &model.Article{
		Title:    testArticleTitle,
		Content:  testArticleContent,
		Language: "en",
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	_, err = io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}
}

func TestGenerate_WithExcerpt(t *testing.T) {
	pub := NewPublisher()
	article := &model.Article{
		Title:   testArticleTitle,
		Content: testArticleContent,
		Excerpt: "This is a test excerpt",
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	_, err = io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}
}

func TestGenerate_WithImage(t *testing.T) {
	pub := NewPublisher()
	article := &model.Article{
		Title:    testArticleTitle,
		Content:  testArticleContent,
		ImageURL: "https://example.com/image.jpg",
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	_, err = io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}
}

func TestGenerate_ZeroReadingTime(t *testing.T) {
	pub := NewPublisher()
	article := &model.Article{
		Title:              testArticleTitle,
		Content:            testArticleContent,
		ReadingTimeMinutes: 0,
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	_, err = io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}
}

func TestGenerate_NilPublishedAt(t *testing.T) {
	pub := NewPublisher()
	article := &model.Article{
		Title:       testArticleTitle,
		Content:     testArticleContent,
		PublishedAt: nil,
	}

	epubReader, err := pub.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	if epubReader == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	_, err = io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data = %v", err)
	}
}

func TestBuildMetadataHeader_EmptyArticle(t *testing.T) {
	article := &model.Article{}
	result := buildMetadataHeader(article)

	if result != "" {
		t.Errorf("buildMetadataHeader() expected empty string, got %q", result)
	}
}

func TestBuildMetadataHeader_OnlyTitle(t *testing.T) {
	article := &model.Article{Title: testTitle}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if !strings.Contains(result, "Title: Test Title") {
		t.Errorf("buildMetadataHeader() expected title in result, got %q", result)
	}

	if !strings.Contains(result, "<strong>") {
		t.Error("buildMetadataHeader() expected <strong> tag in result")
	}

	if !strings.Contains(result, "</p>") {
		t.Error("buildMetadataHeader() expected </p> tag in result")
	}
}

func TestBuildMetadataHeader_OnlyAuthor(t *testing.T) {
	article := &model.Article{Author: testArticleAuthor}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if !strings.Contains(result, "Author: Test Author") {
		t.Errorf("buildMetadataHeader() expected author in result, got %q", result)
	}

	if !strings.Contains(result, "<strong>") {
		t.Error("buildMetadataHeader() expected <strong> tag in result")
	}
}

func TestBuildMetadataHeader_OnlySiteName(t *testing.T) {
	article := &model.Article{SiteName: testExampleSite}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if !strings.Contains(result, "Source: Example Site") {
		t.Errorf("buildMetadataHeader() expected site name in result, got %q", result)
	}
}

func TestBuildMetadataHeader_OnlySourceDomain(t *testing.T) {
	article := &model.Article{SourceDomain: testExampleDomain}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if !strings.Contains(result, "Source: example.com") {
		t.Errorf("buildMetadataHeader() expected source domain in result, got %q", result)
	}
}

func TestBuildMetadataHeader_SourceWithBothSiteNameAndDomain(t *testing.T) {
	article := &model.Article{SiteName: testExampleSite, SourceDomain: testExampleDomain}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if !strings.Contains(result, "Source: Example Site (example.com)") {
		t.Errorf("buildMetadataHeader() expected combined source in result, got %q", result)
	}
}

func TestBuildMetadataHeader_SourceWithBothEmpty(t *testing.T) {
	article := &model.Article{SiteName: "", SourceDomain: ""}
	result := buildMetadataHeader(article)

	if result != "" {
		t.Errorf("buildMetadataHeader() expected empty string, got %q", result)
	}
}

func TestBuildMetadataHeader_OnlyReadingTime(t *testing.T) {
	article := &model.Article{ReadingTimeMinutes: 5}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if !strings.Contains(result, "Reading time: 5 min") {
		t.Errorf("buildMetadataHeader() expected reading time in result, got %q", result)
	}
}

func TestBuildMetadataHeader_OnlyPublishedAt(t *testing.T) {
	publishedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	article := &model.Article{PublishedAt: &publishedAt}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if !strings.Contains(result, "Published: 2024-01-15T10:30:00Z") {
		t.Errorf("buildMetadataHeader() expected published date in result, got %q", result)
	}
}

func TestBuildMetadataHeader_OnlyCreatedAt(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	article := &model.Article{CreatedAt: createdAt}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if !strings.Contains(result, "Added: 2024-01-15T10:30:00Z") {
		t.Errorf("buildMetadataHeader() expected created date in result, got %q", result)
	}
}

func TestBuildMetadataHeader_AllFieldsPopulated(t *testing.T) {
	publishedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC)
	article := &model.Article{
		Title:              testTitle,
		Author:             testArticleAuthor,
		SiteName:           testExampleSite,
		SourceDomain:       testExampleDomain,
		ReadingTimeMinutes: 5,
		PublishedAt:        &publishedAt,
		CreatedAt:          createdAt,
	}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	expectedFields := []string{
		"Title: Test Title",
		"Author: Test Author",
		"Source: Example Site (example.com)",
		"Reading time: 5 min",
		"Published: 2024-01-15T10:30:00Z",
		"Added: 2024-01-16T11:00:00Z",
	}

	for _, field := range expectedFields {
		if !strings.Contains(result, field) {
			t.Errorf("buildMetadataHeader() expected %q in result, got %q", field, result)
		}
	}

	if !strings.Contains(result, `<div style="font-size: 0.85em; color: #666;`) {
		t.Error("buildMetadataHeader() expected div with style in result")
	}

	if !strings.Contains(result, `background-color: #f9f9f9;`) {
		t.Error("buildMetadataHeader() expected background color in result")
	}
}

func TestBuildMetadataHeader_ZeroReadingTimeNotShown(t *testing.T) {
	article := &model.Article{
		Title:              testTitle,
		ReadingTimeMinutes: 0,
	}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if strings.Contains(result, "Reading time") {
		t.Errorf("buildMetadataHeader() should not show reading time when zero, got %q", result)
	}

	if !strings.Contains(result, "Title: Test Title") {
		t.Errorf("buildMetadataHeader() expected title in result, got %q", result)
	}
}

func TestBuildMetadataHeader_NilPublishedAtNotShown(t *testing.T) {
	article := &model.Article{
		Title:       testTitle,
		PublishedAt: nil,
	}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if strings.Contains(result, "Published:") {
		t.Errorf("buildMetadataHeader() should not show published date when nil, got %q", result)
	}
}

func TestBuildMetadataHeader_ZeroPublishedAtNotShown(t *testing.T) {
	zeroTime := time.Time{}
	article := &model.Article{
		Title:       testTitle,
		PublishedAt: &zeroTime,
	}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if strings.Contains(result, "Published:") {
		t.Errorf("buildMetadataHeader() should not show published date when zero, got %q", result)
	}
}

func TestBuildMetadataHeader_ZeroCreatedAtNotShown(t *testing.T) {
	article := &model.Article{
		Title:     testTitle,
		CreatedAt: time.Time{},
	}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if strings.Contains(result, "Added:") {
		t.Errorf("buildMetadataHeader() should not show added date when zero, got %q", result)
	}
}

func TestBuildMetadataHeader_EmptyTitleNotShown(t *testing.T) {
	article := &model.Article{
		Title:  "",
		Author: testArticleAuthor,
	}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if strings.Contains(result, "Title:") {
		t.Errorf("buildMetadataHeader() should not show title when empty, got %q", result)
	}

	if !strings.Contains(result, "Author: Test Author") {
		t.Errorf("buildMetadataHeader() expected author in result, got %q", result)
	}
}

func TestBuildMetadataHeader_EmptyAuthorNotShown(t *testing.T) {
	article := &model.Article{
		Title:  testTitle,
		Author: "",
	}
	result := buildMetadataHeader(article)

	if result == "" {
		t.Fatal("buildMetadataHeader() expected non-empty result")
	}

	if strings.Contains(result, "Author:") {
		t.Errorf("buildMetadataHeader() should not show author when empty, got %q", result)
	}

	if !strings.Contains(result, "Title: Test Title") {
		t.Errorf("buildMetadataHeader() expected title in result, got %q", result)
	}
}

func TestBuildMetadataHeader_HtmlStructure(t *testing.T) {
	article := &model.Article{Title: testTitle}
	result := buildMetadataHeader(article)

	expectedTags := []string{
		`<div`,
		`style="font-size: 0.85em;`,
		`color: #666;`,
		`margin-bottom: 2em;`,
		`padding: 1em;`,
		`border-left: 3px solid #ccc;`,
		`background-color: #f9f9f9;"`,
		`</div>`,
	}

	for _, tag := range expectedTags {
		if !strings.Contains(result, tag) {
			t.Errorf("buildMetadataHeader() expected %q in HTML structure, got %q", tag, result)
		}
	}
}

// extractXHTMLContent reads an EPUB from raw bytes and returns the content
// of the chapter XHTML file as a string.
func extractXHTMLContent(t *testing.T, epubData []byte) string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(epubData), int64(len(epubData)))
	if err != nil {
		t.Fatalf("failed to open epub as zip: %v", err)
	}
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "chapter1.xhtml") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open chapter: %v", err)
			}
			defer func() { _ = rc.Close() }()
			data, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("failed to read chapter: %v", err)
			}
			return string(data)
		}
	}
	t.Fatal("chapter1.xhtml not found in epub")
	return ""
}

// extractCSSContent reads an EPUB from raw bytes and returns the content
// of the CSS stylesheet file as a string.
func extractCSSContent(t *testing.T, epubData []byte) string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(epubData), int64(len(epubData)))
	if err != nil {
		t.Fatalf("failed to open epub as zip: %v", err)
	}
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".css") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open css: %v", err)
			}
			defer func() { _ = rc.Close() }()
			data, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("failed to read css: %v", err)
			}
			return string(data)
		}
	}
	t.Fatal("css file not found in epub")
	return ""
}

func TestGenerate_EPUBContainsStylesheet(t *testing.T) {
	pub := NewPublisher(WithMemoryStorage())
	article := &model.Article{
		Title:   "Test Article",
		Content: "<p>This is test content</p>",
	}

	epubReader, err := pub.GenerateEPUB(article)
	if err != nil {
		t.Fatalf("GenerateEPUB() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	data, err := io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data: %v", err)
	}

	css := extractCSSContent(t, data)

	if !strings.Contains(css, "white-space: pre-wrap") {
		t.Error("EPUB CSS should contain white-space: pre-wrap for code blocks")
	}
	if !strings.Contains(css, "font-family: monospace") {
		t.Error("EPUB CSS should contain font-family: monospace for code blocks")
	}
	if !strings.Contains(css, "pre {") {
		t.Error("EPUB CSS should contain pre selector")
	}
	if !strings.Contains(css, "code {") {
		t.Error("EPUB CSS should contain code selector")
	}
}

func TestGenerate_ChapterReferencesStylesheet(t *testing.T) {
	pub := NewPublisher(WithMemoryStorage())
	article := &model.Article{
		Title:   "Test Article",
		Content: "<p>This is test content</p>",
	}

	epubReader, err := pub.GenerateEPUB(article)
	if err != nil {
		t.Fatalf("GenerateEPUB() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	data, err := io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data: %v", err)
	}

	xhtml := extractXHTMLContent(t, data)

	if !strings.Contains(xhtml, `rel="stylesheet"`) {
		t.Error("Chapter XHTML should reference a stylesheet")
	}
	if !strings.Contains(xhtml, `type="text/css"`) {
		t.Error("Chapter XHTML stylesheet link should specify text/css type")
	}
	if !strings.Contains(xhtml, "styles.css") {
		t.Error("Chapter XHTML should reference styles.css")
	}
}

func TestGenerate_CodeBlockPreservesIndentation(t *testing.T) {
	pub := NewPublisher(WithMemoryStorage())
	codeContent := `<pre><code>func main() {
    fmt.Println("Hello, World!")
    if true {
        for i := 0; i < 10; i++ {
            fmt.Println(i)
        }
    }
}</code></pre>`
	article := &model.Article{
		Title:   "Test Article",
		Content: codeContent,
	}

	epubReader, err := pub.GenerateEPUB(article)
	if err != nil {
		t.Fatalf("GenerateEPUB() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	data, err := io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data: %v", err)
	}

	xhtml := extractXHTMLContent(t, data)

	// Verify the indented code is preserved in the XHTML
	if !strings.Contains(xhtml, "    fmt.Println") {
		t.Errorf("Code block should preserve indentation, got XHTML:\n%s", xhtml)
	}
	if !strings.Contains(xhtml, "        for i := 0") {
		t.Errorf("Code block should preserve nested indentation, got XHTML:\n%s", xhtml)
	}
}

func TestGenerate_InlineCodeStyled(t *testing.T) {
	pub := NewPublisher(WithMemoryStorage())
	article := &model.Article{
		Title:   "Test Article",
		Content: "<p>Use the <code>fmt.Println()</code> function.</p>",
	}

	epubReader, err := pub.GenerateEPUB(article)
	if err != nil {
		t.Fatalf("GenerateEPUB() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	data, err := io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data: %v", err)
	}

	css := extractCSSContent(t, data)

	if !strings.Contains(css, "p code, li code") {
		t.Error("EPUB CSS should contain inline code selector (p code, li code)")
	}
	if !strings.Contains(css, "background-color") {
		t.Error("EPUB CSS inline code style should specify background-color")
	}
}

func TestGenerate_MixedContentWithCode(t *testing.T) {
	pub := NewPublisher(WithMemoryStorage())
	article := &model.Article{
		Title:   "Code Article",
		Author:  "Dev Author",
		Content: "<p>Here is a code example:</p>\n<pre><code>func hello() {\n    return \"world\"\n}</code></pre>\n<p>End.</p>",
	}

	epubReader, err := pub.GenerateEPUB(article)
	if err != nil {
		t.Fatalf("GenerateEPUB() unexpected error = %v", err)
	}
	defer func() { _ = epubReader.Close() }()

	data, err := io.ReadAll(epubReader)
	if err != nil {
		t.Fatalf("Failed to read epub data: %v", err)
	}

	xhtml := extractXHTMLContent(t, data)
	css := extractCSSContent(t, data)

	// Verify content is present
	if !strings.Contains(xhtml, "Here is a code example") {
		t.Error("XHTML should contain article text")
	}
	if !strings.Contains(xhtml, "    return") {
		t.Error("XHTML should preserve code indentation")
	}

	// Verify CSS is embedded
	if !strings.Contains(css, "pre {") {
		t.Error("CSS should style pre elements")
	}
	if !strings.Contains(css, "code {") {
		t.Error("CSS should style code elements")
	}

	// Verify stylesheet is referenced from chapter
	if !strings.Contains(xhtml, `rel="stylesheet"`) {
		t.Error("Chapter should reference the stylesheet")
	}
}
