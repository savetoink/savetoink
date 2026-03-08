package content

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetcherTypeString(t *testing.T) {
	tests := []struct {
		name     string
		ft       FetcherType
		expected string
	}{
		{
			name:     "go fetcher type",
			ft:       FetcherTypeGo,
			expected: "go",
		},
		{
			name:     "browserless fetcher type",
			ft:       FetcherTypeBrowserless,
			expected: "browserless",
		},
		{
			name:     "unknown fetcher type",
			ft:       FetcherType(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ft.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFetchReturnsCorrectFetcherType(t *testing.T) {
	ctx := context.Background()

	t.Run("go fetcher type on success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>Test</body></html>"))
		}))
		defer server.Close()

		fetcher := NewFetcher("")
		result, err := fetcher.Fetch(ctx, server.URL)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, FetcherTypeGo, result.Type)
		assert.Equal(t, "go", result.Type.String())
	})

	t.Run("browserless fallback fetcher type", func(t *testing.T) {
		goServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer goServer.Close()

		browserlessServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>Browserless Test</body></html>"))
		}))
		defer browserlessServer.Close()

		fetcher := NewFetcher("test-key")
		result, err := fetcher.Fetch(ctx, goServer.URL)

		if err == nil {
			assert.NotNil(t, result)
			assert.Equal(t, FetcherTypeBrowserless, result.Type)
			assert.Equal(t, "browserless", result.Type.String())
		}
	})
}

func TestFetchContentTypeValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("non-html content type returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"title": "test"}`))
		}))
		defer server.Close()

		fetcher := NewFetcher("")
		result, err := fetcher.Fetch(ctx, server.URL)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestFetchResult(t *testing.T) {
	html := []byte("<html><body>Test Content</body></html>")

	result := &FetchResult{
		HTML: html,
		Type: FetcherTypeGo,
	}

	assert.Equal(t, html, result.HTML)
	assert.Equal(t, FetcherTypeGo, result.Type)
	assert.Equal(t, "go", result.Type.String())
}

func TestFetchWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	fetcher := NewFetcher("")
	result, err := fetcher.Fetch(ctx, server.URL)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFetchInvalidURL(t *testing.T) {
	ctx := context.Background()
	fetcher := NewFetcher("")

	tests := []struct {
		name string
		url  string
	}{
		{"empty url", ""},
		{"invalid scheme", "ftp://example.com"},
		{"no host", "https://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fetcher.Fetch(ctx, tt.url)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}
