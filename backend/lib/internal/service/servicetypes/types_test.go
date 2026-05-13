package servicetypes

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccount       = "test-account"
	testArticleID1    = "article-1"
	testArticleURL1   = "https://example.com/article1"
	testID            = "test-id"
	testArticleURL    = "https://example.com/article"
	testArticleTitle  = "Test Article"
	testDeviceEmail   = "device@example.com"
	testStatusSuccess = "success"
	testMsgSuccess    = "Email sent successfully"
	testMsgID         = "msg-123"
)

func TestGetArticlesResult_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		result   GetArticlesResult
		wantJSON string
	}{
		{
			name: "full result with articles",
			result: GetArticlesResult{
				Articles: []*model.Article{
					{
						Account:   testAccount,
						ID:        testArticleID1,
						URL:       testArticleURL1,
						Title:     "Article 1",
						CreatedAt: time.Time{},
					},
					{
						Account:   testAccount,
						ID:        "article-2",
						URL:       "https://example.com/article2",
						Title:     "Article 2",
						CreatedAt: time.Time{},
					},
				},
				Page:     1,
				PageSize: 10,
				Total:    25,
				HasMore:  true,
			},
			wantJSON: `{"Articles":[{"account":"test-account","id":"article-1",` +
				`"url":"https://example.com/article1","title":"Article 1",` +
				`"createdAt":"0001-01-01T00:00:00Z"},` +
				`{"account":"test-account","id":"article-2",` +
				`"url":"https://example.com/article2","title":"Article 2",` +
				`"createdAt":"0001-01-01T00:00:00Z"}],` +
				`"Page":1,"PageSize":10,"Total":25,"HasMore":true}`,
		},
		{
			name: "empty articles",
			result: GetArticlesResult{
				Articles: []*model.Article{},
				Page:     1,
				PageSize: 10,
				Total:    0,
				HasMore:  false,
			},
			wantJSON: `{"Articles":[],"Page":1,"PageSize":10,"Total":0,"HasMore":false}`,
		},
		{
			name: "nil articles",
			result: GetArticlesResult{
				Articles: nil,
				Page:     1,
				PageSize: 10,
				Total:    0,
				HasMore:  false,
			},
			wantJSON: `{"Articles":null,"Page":1,"PageSize":10,"Total":0,"HasMore":false}`,
		},
		{
			name: "first page",
			result: GetArticlesResult{
				Articles: []*model.Article{
					{
						Account:   testAccount,
						ID:        testArticleID1,
						URL:       testArticleURL1,
						CreatedAt: time.Time{},
					},
				},
				Page:     1,
				PageSize: 10,
				Total:    5,
				HasMore:  false,
			},
			wantJSON: `{"Articles":[{"account":"test-account","id":"article-1",` +
				`"url":"https://example.com/article1","createdAt":"0001-01-01T00:00:00Z"}],` +
				`"Page":1,"PageSize":10,"Total":5,"HasMore":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled GetArticlesResult
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.result.Page, unmarshaled.Page)
			assert.Equal(t, tt.result.PageSize, unmarshaled.PageSize)
			assert.Equal(t, tt.result.Total, unmarshaled.Total)
			assert.Equal(t, tt.result.HasMore, unmarshaled.HasMore)
			assert.Equal(t, len(tt.result.Articles), len(unmarshaled.Articles))
		})
	}
}

func TestGetArticlesResult_Pagination(t *testing.T) {
	t.Run("calculate has more - last page", func(t *testing.T) {
		result := GetArticlesResult{
			Articles: []*model.Article{},
			Page:     3,
			PageSize: 10,
			Total:    25,
			HasMore:  false,
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var unmarshaled GetArticlesResult
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		assert.False(t, unmarshaled.HasMore)
	})

	t.Run("calculate has more - middle page", func(t *testing.T) {
		result := GetArticlesResult{
			Articles: []*model.Article{},
			Page:     2,
			PageSize: 10,
			Total:    30,
			HasMore:  true,
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var unmarshaled GetArticlesResult
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		assert.True(t, unmarshaled.HasMore)
	})
}

func TestDeleteArticleResult_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		result   DeleteArticleResult
		wantJSON string
	}{
		{
			name:     "deleted articles",
			result:   DeleteArticleResult{Deleted: 5},
			wantJSON: `{"Deleted":5}`,
		},
		{
			name:     "no articles deleted",
			result:   DeleteArticleResult{Deleted: 0},
			wantJSON: `{"Deleted":0}`,
		},
		{
			name:     "single article deleted",
			result:   DeleteArticleResult{Deleted: 1},
			wantJSON: `{"Deleted":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled DeleteArticleResult
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.result.Deleted, unmarshaled.Deleted)
		})
	}
}

func TestSendArticleResult_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		result   SendArticleResult
		wantJSON string
	}{
		{
			name: "full result",
			result: SendArticleResult{
				Article: &model.Article{
					Account:   testAccount,
					ID:        testID,
					URL:       testArticleURL,
					Title:     testArticleTitle,
					CreatedAt: time.Time{},
				},
				DeviceEmail: testDeviceEmail,
				EmailResp: &email.SendEmailResponse{
					Status:    testStatusSuccess,
					Message:   testMsgSuccess,
					MessageID: testMsgID,
				},
			},
			wantJSON: `{"Article":{"account":"test-account","id":"test-id",` +
				`"url":"https://example.com/article","title":"Test Article",` +
				`"createdAt":"0001-01-01T00:00:00Z"},` +
				`"DeviceEmail":"device@example.com","EmailResp":{"status":"success",` +
				`"message":"Email sent successfully","message_id":"msg-123"}}`,
		},
		{
			name: "minimal result",
			result: SendArticleResult{
				Article: &model.Article{
					Account:   testAccount,
					ID:        testID,
					URL:       testArticleURL,
					CreatedAt: time.Time{},
				},
				DeviceEmail: testDeviceEmail,
			},
			wantJSON: `{"Article":{"account":"test-account","id":"test-id",` +
				`"url":"https://example.com/article","createdAt":"0001-01-01T00:00:00Z"},` +
				`"DeviceEmail":"device@example.com","EmailResp":null}`,
		},
		{
			name: "result with nil email response",
			result: SendArticleResult{
				Article: &model.Article{
					Account:   testAccount,
					ID:        testID,
					URL:       testArticleURL,
					CreatedAt: time.Time{},
				},
				DeviceEmail: testDeviceEmail,
				EmailResp:   nil,
			},
			wantJSON: `{"Article":{"account":"test-account","id":"test-id",` +
				`"url":"https://example.com/article","createdAt":"0001-01-01T00:00:00Z"},` +
				`"DeviceEmail":"device@example.com","EmailResp":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled SendArticleResult
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			if tt.result.Article != nil {
				require.NotNil(t, unmarshaled.Article)
				assert.Equal(t, tt.result.Article.Account, unmarshaled.Article.Account)
				assert.Equal(t, tt.result.Article.ID, unmarshaled.Article.ID)
				assert.Equal(t, tt.result.Article.URL, unmarshaled.Article.URL)
			} else {
				assert.Nil(t, unmarshaled.Article)
			}

			assert.Equal(t, tt.result.DeviceEmail, unmarshaled.DeviceEmail)

			if tt.result.EmailResp != nil {
				require.NotNil(t, unmarshaled.EmailResp)
				assert.Equal(t, tt.result.EmailResp.Status, unmarshaled.EmailResp.Status)
				assert.Equal(t, tt.result.EmailResp.Message, unmarshaled.EmailResp.Message)
				assert.Equal(t, tt.result.EmailResp.MessageID, unmarshaled.EmailResp.MessageID)
			} else {
				assert.Nil(t, unmarshaled.EmailResp)
			}
		})
	}
}

func TestSendArticleResult_NilFields(t *testing.T) {
	result := SendArticleResult{
		Article:     nil,
		DeviceEmail: "",
		EmailResp:   nil,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled SendArticleResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Nil(t, unmarshaled.Article)
	assert.Empty(t, unmarshaled.DeviceEmail)
	assert.Nil(t, unmarshaled.EmailResp)
}

func TestSendArticleResult_FullArticle(t *testing.T) {
	result := SendArticleResult{
		Article: &model.Article{
			Account:            testAccount,
			ID:                 testID,
			URL:                testArticleURL,
			CreatedAt:          time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
			Title:              testArticleTitle,
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
		DeviceEmail: testDeviceEmail,
		EmailResp: &email.SendEmailResponse{
			Status:    testStatusSuccess,
			Message:   testMsgSuccess,
			MessageID: testMsgID,
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled SendArticleResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	require.NotNil(t, unmarshaled.Article)
	assert.Equal(t, result.Article.Account, unmarshaled.Article.Account)
	assert.Equal(t, result.Article.ID, unmarshaled.Article.ID)
	assert.Equal(t, result.Article.URL, unmarshaled.Article.URL)
	assert.Equal(t, result.Article.Title, unmarshaled.Article.Title)
	assert.Equal(t, result.Article.Content, unmarshaled.Article.Content)
	assert.Equal(t, result.Article.Author, unmarshaled.Article.Author)
	assert.Equal(t, result.Article.SiteName, unmarshaled.Article.SiteName)
	assert.Equal(t, result.Article.SourceDomain, unmarshaled.Article.SourceDomain)
	assert.Equal(t, result.Article.Excerpt, unmarshaled.Article.Excerpt)
	assert.Equal(t, result.Article.ImageURL, unmarshaled.Article.ImageURL)
	assert.Equal(t, result.Article.ContentType, unmarshaled.Article.ContentType)
	assert.Equal(t, result.Article.Language, unmarshaled.Article.Language)
	assert.Equal(t, result.Article.WordCount, unmarshaled.Article.WordCount)
	assert.Equal(t, result.Article.ReadingTimeMinutes, unmarshaled.Article.ReadingTimeMinutes)
	assert.Equal(t, result.Article.Favorite, unmarshaled.Article.Favorite)

	assert.Equal(t, result.DeviceEmail, unmarshaled.DeviceEmail)

	require.NotNil(t, unmarshaled.EmailResp)
	assert.Equal(t, result.EmailResp.Status, unmarshaled.EmailResp.Status)
	assert.Equal(t, result.EmailResp.Message, unmarshaled.EmailResp.Message)
	assert.Equal(t, result.EmailResp.MessageID, unmarshaled.EmailResp.MessageID)
}
