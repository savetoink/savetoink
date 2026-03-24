package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testArticleURL = "https://example.com/article"
)

func TestArticle_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		article  Article
		wantJSON string
	}{
		{
			name: "full article",
			article: Article{
				Account:            "test-account",
				ID:                 "test-id",
				URL:                testArticleURL,
				CreatedAt:          time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Title:              "Test Article",
				Content:            "Test content",
				Author:             "Test Author",
				SiteName:           "Test Site",
				SourceDomain:       "example.com",
				Excerpt:            "Test excerpt",
				ImageURL:           "https://example.com/image.jpg",
				ContentType:        "text/html",
				Language:           "en",
				Error:              "",
				WordCount:          100,
				ReadingTimeMinutes: 1,
				PublishedAt:        func() *time.Time { t := time.Date(2024, 3, 14, 10, 0, 0, 0, time.UTC); return &t }(),
				Favorite:           true,
			},
			wantJSON: `{"account":"test-account","id":"test-id","url":"https://example.com/article",` +
				`"createdAt":"2024-03-15T10:30:00Z","title":"Test Article","content":"Test content",` +
				`"author":"Test Author","siteName":"Test Site","sourceDomain":"example.com",` +
				`"excerpt":"Test excerpt","imageUrl":"https://example.com/image.jpg","contentType":"text/html",` +
				`"language":"en","wordCount":100,"readingTimeMinutes":1,` +
				`"publishedAt":"2024-03-14T10:00:00Z","favorite":true}`,
		},
		{
			name: "minimal article",
			article: Article{
				Account:   "test-account",
				ID:        "test-id",
				URL:       testArticleURL,
				CreatedAt: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
			},
			wantJSON: `{"account":"test-account","id":"test-id","url":"https://example.com/article",` +
				`"createdAt":"2024-03-15T10:30:00Z"}`,
		},
		{
			name: "article with nil PublishedAt",
			article: Article{
				Account:   "test-account",
				ID:        "test-id",
				URL:       testArticleURL,
				CreatedAt: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Title:     "Test",
			},
			wantJSON: `{"account":"test-account","id":"test-id","url":"https://example.com/article",` +
				`"createdAt":"2024-03-15T10:30:00Z","title":"Test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.article)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled Article
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.article.Account, unmarshaled.Account)
			assert.Equal(t, tt.article.ID, unmarshaled.ID)
			assert.Equal(t, tt.article.URL, unmarshaled.URL)
			assert.True(t, tt.article.CreatedAt.Equal(unmarshaled.CreatedAt))
		})
	}
}

func TestArticle_DynamoDBAttributeMapping(t *testing.T) {
	tests := []struct {
		name    string
		article Article
	}{
		{
			name: "full article",
			article: Article{
				Account:            "test-account",
				ID:                 "test-id",
				URL:                testArticleURL,
				CreatedAt:          time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Title:              "Test Article",
				Content:            "Test content",
				Author:             "Test Author",
				SiteName:           "Test Site",
				SourceDomain:       "example.com",
				Excerpt:            "Test excerpt",
				ImageURL:           "https://example.com/image.jpg",
				ContentType:        "text/html",
				Language:           "en",
				Error:              "",
				WordCount:          100,
				ReadingTimeMinutes: 1,
				PublishedAt:        func() *time.Time { t := time.Date(2024, 3, 14, 10, 0, 0, 0, time.UTC); return &t }(),
				Favorite:           true,
			},
		},
		{
			name: "minimal article",
			article: Article{
				Account:   "test-account",
				ID:        "test-id",
				URL:       testArticleURL,
				CreatedAt: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "article with error",
			article: Article{
				Account:   "test-account",
				ID:        "test-id",
				URL:       testArticleURL,
				CreatedAt: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Error:     "failed to fetch content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := attributevalue.Marshal(tt.article)
			require.NoError(t, err)

			var unmarshaled Article
			err = attributevalue.Unmarshal(marshaled, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, tt.article.Account, unmarshaled.Account)
			assert.Equal(t, tt.article.ID, unmarshaled.ID)
			assert.Equal(t, tt.article.URL, unmarshaled.URL)
			assert.True(t, tt.article.CreatedAt.Equal(unmarshaled.CreatedAt))
			assert.Equal(t, tt.article.Title, unmarshaled.Title)
			assert.Equal(t, tt.article.Content, unmarshaled.Content)
			assert.Equal(t, tt.article.Author, unmarshaled.Author)
			assert.Equal(t, tt.article.SiteName, unmarshaled.SiteName)
			assert.Equal(t, tt.article.SourceDomain, unmarshaled.SourceDomain)
			assert.Equal(t, tt.article.Excerpt, unmarshaled.Excerpt)
			assert.Equal(t, tt.article.ImageURL, unmarshaled.ImageURL)
			assert.Equal(t, tt.article.ContentType, unmarshaled.ContentType)
			assert.Equal(t, tt.article.Language, unmarshaled.Language)
			assert.Equal(t, tt.article.Error, unmarshaled.Error)
			assert.Equal(t, tt.article.WordCount, unmarshaled.WordCount)
			assert.Equal(t, tt.article.ReadingTimeMinutes, unmarshaled.ReadingTimeMinutes)
			// Favorite field is not persisted to DynamoDB; it's derived from the presence of accountFavorite
			assert.False(t, unmarshaled.Favorite)

			if tt.article.PublishedAt != nil {
				require.NotNil(t, unmarshaled.PublishedAt)
				assert.True(t, tt.article.PublishedAt.Equal(*unmarshaled.PublishedAt))
			} else {
				assert.Nil(t, unmarshaled.PublishedAt)
			}
		})
	}
}

func TestArticle_EmptyFields(t *testing.T) {
	article := Article{
		Account:   "test-account",
		ID:        "test-id",
		URL:       "https://example.com/article",
		CreatedAt: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(article)
	require.NoError(t, err)

	var unmarshaled Article
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Empty(t, unmarshaled.Title)
	assert.Empty(t, unmarshaled.Content)
	assert.Empty(t, unmarshaled.Author)
	assert.Empty(t, unmarshaled.SiteName)
	assert.Empty(t, unmarshaled.SourceDomain)
	assert.Empty(t, unmarshaled.Excerpt)
	assert.Empty(t, unmarshaled.ImageURL)
	assert.Empty(t, unmarshaled.ContentType)
	assert.Empty(t, unmarshaled.Language)
	assert.Empty(t, unmarshaled.Error)
	assert.Zero(t, unmarshaled.WordCount)
	assert.Zero(t, unmarshaled.ReadingTimeMinutes)
	assert.False(t, unmarshaled.Favorite)
	assert.Nil(t, unmarshaled.PublishedAt)
}

func TestErrorResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     ErrorResponse
		wantJSON string
	}{
		{
			name:     "with error message",
			resp:     ErrorResponse{Error: "test error"},
			wantJSON: `{"error":"test error"}`,
		},
		{
			name:     "empty error",
			resp:     ErrorResponse{Error: ""},
			wantJSON: `{"error":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled ErrorResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.Error, unmarshaled.Error)
		})
	}
}
