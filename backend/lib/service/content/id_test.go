package content

import (
	"net/url"
	"testing"

	"github.com/shaftoe/savetoink/backend/lib/internal/validation"
	"github.com/stretchr/testify/assert"
)

const (
	baseArticleURL              = "https://example.com/article/123"
	articleURLWithQuery         = "https://example.com/article/123?source=twitter&utm=test"
	articleURLWithSource        = "https://example.com/article/123?source=twitter"
	articleURLWithTrailingSlash = "https://example.com/article/123/"
	httpArticleURL              = "http://example.com/article"
	baseURL                     = "https://example.com/"
	schemeError                 = "must use http or https scheme"
)

func TestArticleIDFromURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid URL",
			url:     baseArticleURL,
			wantErr: false,
		},
		{
			name:    "valid URL with query params",
			url:     articleURLWithQuery,
			wantErr: false,
		},
		{
			name:    "valid URL with fragment",
			url:     "https://example.com/article/123#section-1",
			wantErr: false,
		},
		{
			name:    "valid URL with both query and fragment",
			url:     "https://example.com/article/123?ref=news#intro",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			url:     "http://example.com/article/456",
			wantErr: false,
		},
		{
			name:    "URL with trailing slash",
			url:     "https://example.com/article/123/",
			wantErr: false,
		},
		{
			name:        "invalid URL",
			url:         "not-a-url",
			wantErr:     true,
			errContains: schemeError,
		},
		{
			name:        "URL without scheme",
			url:         "example.com/article",
			wantErr:     true,
			errContains: schemeError,
		},
		{
			name:        "empty URL",
			url:         "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := validation.ValidateURL(tt.url)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, u)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected parse error for %s: %v", tt.url, err)
			}

			id, err := ArticleIDFromURL(u)

			assert.NoError(t, err)
			assert.NotEmpty(t, id)
			assert.Len(t, id, 36)
		})
	}
}

func TestArticleIDFromURL_Deterministic(t *testing.T) {
	url1, _ := url.Parse(articleURLWithSource)
	url2, _ := url.Parse("https://example.com/article/123?utm_source=newsletter#intro")
	url3, _ := url.Parse("https://example.com/article/123/")

	id1, err1 := ArticleIDFromURL(url1)
	id2, err2 := ArticleIDFromURL(url2)
	id3, err3 := ArticleIDFromURL(url3)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	assert.NotEqual(t, id1, id2, "IDs should be different for different query params")
	assert.NotEqual(t, id1, id3, "IDs should be different for trailing slash")
}

func TestArticleIDFromURL_DifferentURLs(t *testing.T) {
	urls := []string{
		"https://example.com/article/1",
		"https://example.com/article/2",
		"https://other.com/article/1",
		"https://example.org/article/1",
	}

	ids := make(map[string]bool)
	for _, urlStr := range urls {
		u, err := url.Parse(urlStr)
		if err != nil {
			t.Fatalf("Failed to parse URL %s: %v", urlStr, err)
		}
		id, err := ArticleIDFromURL(u)
		assert.NoError(t, err)
		assert.False(t, ids[id], "ID should be unique for each URL")
		ids[id] = true
	}

	assert.Equal(t, len(urls), len(ids))
}

func TestArticleIDFromURL_HttpVsHttps(t *testing.T) {
	urlHTTP, _ := url.Parse(httpArticleURL)
	urlHTTPS, _ := url.Parse("https://example.com/article")

	idHTTP, err1 := ArticleIDFromURL(urlHTTP)
	idHTTPS, err2 := ArticleIDFromURL(urlHTTPS)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, idHTTP, idHTTPS, "HTTP and HTTPS should produce different IDs")
}

func TestCleanURL(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		expectedURL string
		wantErr     bool
		errContains string
	}{
		{
			name:        "preserves query parameters",
			inputURL:    articleURLWithQuery,
			expectedURL: articleURLWithQuery,
			wantErr:     false,
		},
		{
			name:        "strips fragment",
			inputURL:    baseArticleURL + "#section-1",
			expectedURL: baseArticleURL,
			wantErr:     false,
		},
		{
			name:        "strips fragment but preserves query",
			inputURL:    "https://example.com/article/123?ref=news#intro",
			expectedURL: "https://example.com/article/123?ref=news",
			wantErr:     false,
		},
		{
			name:        "preserves clean URL",
			inputURL:    baseArticleURL,
			expectedURL: baseArticleURL,
			wantErr:     false,
		},
		{
			name:        "removes trailing slash",
			inputURL:    articleURLWithTrailingSlash,
			expectedURL: baseArticleURL,
			wantErr:     false,
		},
		{
			name:        "handles root path",
			inputURL:    baseURL,
			expectedURL: baseURL,
			wantErr:     false,
		},
		{
			name:        "handles no path",
			inputURL:    "https://example.com",
			expectedURL: baseURL,
			wantErr:     false,
		},
		{
			name:        "handles HTTP",
			inputURL:    httpArticleURL,
			expectedURL: httpArticleURL,
			wantErr:     false,
		},
		{
			name:        "complex path with query",
			inputURL:    "https://example.com/blog/2023/12/post?id=456&category=tech",
			expectedURL: "https://example.com/blog/2023/12/post?id=456&category=tech",
			wantErr:     false,
		},
		{
			name:        "invalid URL",
			inputURL:    "not-a-url",
			expectedURL: "",
			wantErr:     true,
			errContains: schemeError,
		},
		{
			name:        "URL without scheme",
			inputURL:    "example.com/article",
			expectedURL: "",
			wantErr:     true,
			errContains: schemeError,
		},
		{
			name:        "empty URL",
			inputURL:    "",
			expectedURL: "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := validation.ValidateURL(tt.inputURL)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, u)
				return
			}
			if err != nil {
				t.Fatalf("Unexpected parse error for %s: %v", tt.inputURL, err)
			}

			cleanURL := CleanURL(u)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedURL, cleanURL)
		})
	}
}

func TestCleanURL_Deterministic(t *testing.T) {
	tests := []struct {
		inputURL         string
		expectedCleanURL string
	}{
		{
			inputURL:         articleURLWithSource,
			expectedCleanURL: articleURLWithSource,
		},
		{
			inputURL:         "https://example.com/article/123?utm_source=newsletter#intro",
			expectedCleanURL: "https://example.com/article/123?utm_source=newsletter",
		},
		{
			inputURL:         articleURLWithTrailingSlash,
			expectedCleanURL: baseArticleURL,
		},
		{
			inputURL:         baseArticleURL,
			expectedCleanURL: baseArticleURL,
		},
	}

	for _, tt := range tests {
		u, err := url.Parse(tt.inputURL)
		if err != nil {
			t.Fatalf("Failed to parse URL %s: %v", tt.inputURL, err)
		}
		cleanURL := CleanURL(u)
		assert.NoError(t, err)
		assert.Equal(t, tt.expectedCleanURL, cleanURL)
	}
}
