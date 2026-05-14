package content

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-shiori/dom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

const testHelloWorldContent = "Hello world"

const (
	testHelloWorld = " Hello  world "

	testBrowserlessKey = "test-key"
	testContentType    = "Content-Type"
	testTextHTML       = "text/html"
)

func TestTrafilaturaCleaner_Clean(t *testing.T) {
	cleaner := NewTrafilaturaCleaner()
	testURL, _ := url.Parse("https://example.com/article")

	t.Run("valid article content", func(t *testing.T) {
		html := `<!DOCTYPE html>
<html>
<head>
	<title>Article Title - Example Site</title>
	<meta name="author" content="John Doe">
	<meta name="description" content="Article description">
</head>
<body>
	<article>
		<h1>Article Title</h1>
		<p>This is the main content of the article.</p>
		<p>This is another paragraph.</p>
	</article>
</body>
</html>`

		doc, err := dom.Parse(strings.NewReader(html))
		require.NoError(t, err)

		article, err := cleaner.Clean(context.Background(), doc, testURL)
		require.NoError(t, err)
		assert.NotNil(t, article)
		assert.NotEmpty(t, article.Title)
		assert.NotEmpty(t, article.Content)
		assert.Greater(t, article.WordCount, 0)
		assert.Greater(t, article.ReadingTimeMinutes, 0)
	})

	const simpleHTML = `<html><body><p>Content</p></body></html>`

	t.Run("nil url", func(t *testing.T) {
		doc, err := dom.Parse(strings.NewReader(simpleHTML))
		require.NoError(t, err)

		_, err = cleaner.Clean(context.Background(), doc, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidURL)
	})

	t.Run("nil document", func(t *testing.T) {
		_, err := cleaner.Clean(context.Background(), nil, testURL)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNilContentNode)
	})
}

func TestTrafilaturaCleaner_extractTitleFromDocument(t *testing.T) {
	cleaner := NewTrafilaturaCleaner()

	t.Run("extract from title tag", func(t *testing.T) {
		html := `<html><head><title>Article Title - Site Name</title></head><body></body></html>`
		doc, _ := dom.Parse(strings.NewReader(html))

		title := cleaner.extractTitleFromDocument(doc)
		assert.Equal(t, "Article Title", title)
	})

	t.Run("no separator in title", func(t *testing.T) {
		html := `<html><head><title>Single Title</title></head><body></body></html>`
		doc, _ := dom.Parse(strings.NewReader(html))

		title := cleaner.extractTitleFromDocument(doc)
		assert.Equal(t, "", title)
	})

	t.Run("empty title", func(t *testing.T) {
		html := `<html><head><title></title></head><body></body></html>`
		doc, _ := dom.Parse(strings.NewReader(html))

		title := cleaner.extractTitleFromDocument(doc)
		assert.Equal(t, "", title)
	})

	t.Run("no title tag", func(t *testing.T) {
		html := `<html><head></head><body></body></html>`
		doc, _ := dom.Parse(strings.NewReader(html))

		title := cleaner.extractTitleFromDocument(doc)
		assert.Equal(t, "", title)
	})

	t.Run("fallback to h2", func(t *testing.T) {
		html := `<html><head></head><body><h2>Heading 2 Content</h2></body></html>`
		doc, _ := dom.Parse(strings.NewReader(html))

		title := cleaner.extractTitleFromDocument(doc)
		assert.Equal(t, "Heading 2 Content", title)
	})
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple paragraph",
			input:    "<p>Hello world</p>",
			expected: " Hello world ",
		},
		{
			name:     "multiple tags",
			input:    "<p>Hello</p><p>world</p>",
			expected: testHelloWorld,
		},
		{
			name:     "divs",
			input:    "<div>Hello</div><div>world</div>",
			expected: testHelloWorld,
		},
		{
			name:     "br tags",
			input:    "Hello<br>world",
			expected: testHelloWorldContent,
		},
		{
			name:     "self-closing br",
			input:    "Hello<br />world",
			expected: testHelloWorldContent,
		},
		{
			name:     "mixed tags",
			input:    "<p>Hello</p><div>world</div><br>test",
			expected: " Hello  world  test",
		},
		{
			name:     "nested tags",
			input:    "<div><p>nested</p></div>",
			expected: "  nested  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "simple sentence",
			input:    testHelloWorldContent,
			expected: 2,
		},
		{
			name:     "multiple spaces",
			input:    "Hello   world",
			expected: 2,
		},
		{
			name:     "tabs and newlines",
			input:    "Hello\tworld\n",
			expected: 2,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: 0,
		},
		{
			name:     "longer text",
			input:    "This is a longer text with many words",
			expected: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapStandaloneCodeBlocks(t *testing.T) {
	parseBody := func(t *testing.T, htmlContent string) *html.Node {
		t.Helper()
		doc, err := dom.Parse(strings.NewReader(htmlContent))
		require.NoError(t, err)
		body := dom.QuerySelector(doc, "body")
		require.NotNil(t, body)
		return body
	}

	t.Run("wraps standalone code in pre tags", func(t *testing.T) {
		body := parseBody(t, `<h1>Title</h1><p>Text</p><code>line 1
line 2</code><p>More text</p>`)

		wrapStandaloneCodeBlocks(body)

		result := dom.InnerHTML(body)
		assert.Contains(t, result, "<pre><code>")
		assert.Contains(t, result, "line 1\nline 2")
	})

	t.Run("does not wrap code inside paragraphs", func(t *testing.T) {
		body := parseBody(t, `<p>Use <code>fmt.Println()</code> here</p>`)

		wrapStandaloneCodeBlocks(body)

		result := dom.InnerHTML(body)
		assert.Contains(t, result, "<p>")
		assert.Contains(t, result, "<code>fmt.Println()</code>")
		assert.NotContains(t, result, "<pre><code>")
	})

	t.Run("handles nil content node", func(t *testing.T) {
		wrapStandaloneCodeBlocks(nil)
	})

	t.Run("wraps multiple standalone code blocks", func(t *testing.T) {
		body := parseBody(t, `<code>block1
line2</code><p>text</p><code>block2
line2</code>`)

		wrapStandaloneCodeBlocks(body)

		result := dom.InnerHTML(body)
		assert.Equal(t, 2, strings.Count(result, "<pre><code>"))
	})

	t.Run("preserves code block content with indentation", func(t *testing.T) {
		body := parseBody(t, `<code>func main() {
    fmt.Println("hello")
}</code>`)

		wrapStandaloneCodeBlocks(body)

		result := dom.InnerHTML(body)
		assert.Contains(t, result, "<pre><code>")
		assert.Contains(t, result, "    fmt.Println")
	})
}

func TestToTimePtr(t *testing.T) {
	t.Run("valid time", func(t *testing.T) {
		tm := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		result := toTimePtr(tm)
		assert.NotNil(t, result)
		assert.Equal(t, tm, *result)
	})

	t.Run("zero time", func(t *testing.T) {
		var tm time.Time
		result := toTimePtr(tm)
		assert.Nil(t, result)
	})
}
