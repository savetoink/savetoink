package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArticleRequest_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		req      ArticleRequest
		wantJSON string
	}{
		{
			name: "with send on complete",
			req: ArticleRequest{
				URL:            "https://example.com/article",
				SendOnComplete: true,
			},
			wantJSON: `{"url":"https://example.com/article","send_on_complete":true}`,
		},
		{
			name: "without send on complete",
			req: ArticleRequest{
				URL:            "https://example.com/article",
				SendOnComplete: false,
			},
			wantJSON: `{"url":"https://example.com/article","send_on_complete":false}`,
		},
		{
			name: "minimal request",
			req: ArticleRequest{
				URL: "https://example.com/article",
			},
			wantJSON: `{"url":"https://example.com/article","send_on_complete":false}`,
		},
		{
			name: "with tags",
			req: ArticleRequest{
				URL:  "https://example.com/article",
				Tags: []string{"tech", "programming"},
			},
			wantJSON: `{"url":"https://example.com/article","send_on_complete":false,"tags":["tech","programming"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled ArticleRequest
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.req.URL, unmarshaled.URL)
			assert.Equal(t, tt.req.SendOnComplete, unmarshaled.SendOnComplete)
			assert.Equal(t, tt.req.Tags, unmarshaled.Tags)
		})
	}
}

func TestArticleResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     ArticleResponse
		wantJSON string
	}{
		{
			name: "full response",
			resp: ArticleResponse{
				ID:  "test-id",
				URL: "https://example.com/article",
			},
			wantJSON: `{"id":"test-id","url":"https://example.com/article"}`,
		},
		{
			name: "minimal response",
			resp: ArticleResponse{
				ID: "test-id",
			},
			wantJSON: `{"id":"test-id","url":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled ArticleResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.ID, unmarshaled.ID)
			assert.Equal(t, tt.resp.URL, unmarshaled.URL)
		})
	}
}

func TestHealthResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     HealthResponse
		wantJSON string
	}{
		{
			name:     "healthy",
			resp:     HealthResponse{Status: "ok"},
			wantJSON: `{"status":"ok"}`,
		},
		{
			name:     "empty status",
			resp:     HealthResponse{Status: ""},
			wantJSON: `{"status":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled HealthResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.Status, unmarshaled.Status)
		})
	}
}

func TestListArticlesResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     ListArticlesResponse
		wantJSON string
	}{
		{
			name: "full response",
			resp: ListArticlesResponse{
				Articles: []*model.Article{
					{
						Account:   "test-account",
						ID:        "article-1",
						URL:       "https://example.com/article1",
						CreatedAt: time.Time{},
					},
					{
						Account:   "test-account",
						ID:        "article-2",
						URL:       "https://example.com/article2",
						CreatedAt: time.Time{},
					},
				},
				Page:     1,
				PageSize: 10,
				Total:    25,
				HasMore:  true,
			},
			wantJSON: `{"articles":[{"account":"test-account","id":"article-1",` +
				`"url":"https://example.com/article1","createdAt":"0001-01-01T00:00:00Z"},` +
				`{"account":"test-account","id":"article-2",` +
				`"url":"https://example.com/article2","createdAt":"0001-01-01T00:00:00Z"}],` +
				`"page":1,"page_size":10,"total":25,"has_more":true}`,
		},
		{
			name: "empty articles",
			resp: ListArticlesResponse{
				Articles: []*model.Article{},
				Page:     1,
				PageSize: 10,
				Total:    0,
				HasMore:  false,
			},
			wantJSON: `{"articles":[],"page":1,"page_size":10,"total":0,"has_more":false}`,
		},
		{
			name: "nil articles",
			resp: ListArticlesResponse{
				Articles: nil,
				Page:     1,
				PageSize: 10,
				Total:    0,
				HasMore:  false,
			},
			wantJSON: `{"articles":null,"page":1,"page_size":10,"total":0,"has_more":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled ListArticlesResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.Page, unmarshaled.Page)
			assert.Equal(t, tt.resp.PageSize, unmarshaled.PageSize)
			assert.Equal(t, tt.resp.Total, unmarshaled.Total)
			assert.Equal(t, tt.resp.HasMore, unmarshaled.HasMore)
			assert.Equal(t, len(tt.resp.Articles), len(unmarshaled.Articles))
		})
	}
}

func TestDeleteArticleResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     DeleteArticleResponse
		wantJSON string
	}{
		{
			name:     "deleted articles",
			resp:     DeleteArticleResponse{Deleted: 5},
			wantJSON: `{"deleted":5}`,
		},
		{
			name:     "no articles deleted",
			resp:     DeleteArticleResponse{Deleted: 0},
			wantJSON: `{"deleted":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled DeleteArticleResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.Deleted, unmarshaled.Deleted)
		})
	}
}

func TestFavoriteResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     FavoriteResponse
		wantJSON string
	}{
		{
			name:     "favorited",
			resp:     FavoriteResponse{Favorite: true},
			wantJSON: `{"favorite":true}`,
		},
		{
			name:     "unfavorited",
			resp:     FavoriteResponse{Favorite: false},
			wantJSON: `{"favorite":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled FavoriteResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.Favorite, unmarshaled.Favorite)
		})
	}
}

func TestSendArticleResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     SendArticleResponse
		wantJSON string
	}{
		{
			name:     "success",
			resp:     SendArticleResponse{Status: "queued"},
			wantJSON: `{"status":"queued"}`,
		},
		{
			name:     "processing",
			resp:     SendArticleResponse{Status: "processing"},
			wantJSON: `{"status":"processing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled SendArticleResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.Status, unmarshaled.Status)
		})
	}
}

func TestSendArticleResponseWithCount_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     SendArticleResponseWithCount
		wantJSON string
	}{
		{
			name: "successful send",
			resp: SendArticleResponseWithCount{
				Status:     "queued",
				SendsCount: 5,
			},
			wantJSON: `{"status":"queued","sends_count":5}`,
		},
		{
			name: "first send",
			resp: SendArticleResponseWithCount{
				Status:     "queued",
				SendsCount: 1,
			},
			wantJSON: `{"status":"queued","sends_count":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled SendArticleResponseWithCount
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.Status, unmarshaled.Status)
			assert.Equal(t, tt.resp.SendsCount, unmarshaled.SendsCount)
		})
	}
}

func TestAuthTokenExchangeRequest_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		req      AuthTokenExchangeRequest
		wantJSON string
	}{
		{
			name: "full request",
			req: AuthTokenExchangeRequest{
				Code:        "test-code",
				RedirectURI: "https://example.com/callback",
				GrantType:   "authorization_code",
			},
			wantJSON: `{"code":"test-code","redirect_uri":"https://example.com/callback",` +
				`"grant_type":"authorization_code"}`,
		},
		{
			name: "minimal request",
			req: AuthTokenExchangeRequest{
				Code: "test-code",
			},
			wantJSON: `{"code":"test-code","redirect_uri":"","grant_type":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled AuthTokenExchangeRequest
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.req.Code, unmarshaled.Code)
			assert.Equal(t, tt.req.RedirectURI, unmarshaled.RedirectURI)
			assert.Equal(t, tt.req.GrantType, unmarshaled.GrantType)
		})
	}
}

func TestAuthTokenExchangeResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     AuthTokenExchangeResponse
		wantJSON string
	}{
		{
			name: "full response",
			resp: AuthTokenExchangeResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				IDToken:      "id-token",
				Email:        "user@example.com",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
			},
			wantJSON: `{"access_token":"access-token","refresh_token":"refresh-token",` +
				`"id_token":"id-token","email":"user@example.com","token_type":"Bearer","expires_in":3600}`,
		},
		{
			name: "minimal response",
			resp: AuthTokenExchangeResponse{
				AccessToken: "access-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			},
			wantJSON: `{"access_token":"access-token","token_type":"Bearer","expires_in":3600}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//nolint:gosec // test data, not a real secret
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled AuthTokenExchangeResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.AccessToken, unmarshaled.AccessToken)
			assert.Equal(t, tt.resp.RefreshToken, unmarshaled.RefreshToken)
			assert.Equal(t, tt.resp.IDToken, unmarshaled.IDToken)
			assert.Equal(t, tt.resp.Email, unmarshaled.Email)
			assert.Equal(t, tt.resp.TokenType, unmarshaled.TokenType)
			assert.Equal(t, tt.resp.ExpiresIn, unmarshaled.ExpiresIn)
		})
	}
}

func TestDeviceRequest_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		req      DeviceRequest
		wantJSON string
	}{
		{
			name: "full request",
			req: DeviceRequest{
				DeviceEmail: "device@example.com",
				AutoSend:    true,
			},
			wantJSON: `{"device_email":"device@example.com","auto_send":true}`,
		},
		{
			name: "minimal request",
			req: DeviceRequest{
				DeviceEmail: "device@example.com",
			},
			wantJSON: `{"device_email":"device@example.com","auto_send":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled DeviceRequest
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.req.DeviceEmail, unmarshaled.DeviceEmail)
			assert.Equal(t, tt.req.AutoSend, unmarshaled.AutoSend)
		})
	}
}

func TestDeviceResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     DeviceResponse
		wantJSON string
	}{
		{
			name: "with auto send enabled",
			resp: DeviceResponse{
				DeviceEmail: "device@example.com",
				AutoSend:    true,
			},
			wantJSON: `{"device_email":"device@example.com","auto_send":true}`,
		},
		{
			name: "without auto send",
			resp: DeviceResponse{
				DeviceEmail: "device@example.com",
				AutoSend:    false,
			},
			wantJSON: `{"device_email":"device@example.com","auto_send":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled DeviceResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.DeviceEmail, unmarshaled.DeviceEmail)
			assert.Equal(t, tt.resp.AutoSend, unmarshaled.AutoSend)
		})
	}
}

func TestUserProfileResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     UserProfileResponse
		wantJSON string
	}{
		{
			name: "full profile",
			resp: UserProfileResponse{
				Account:     "test-account",
				Email:       "user@example.com",
				DeviceEmail: "device@example.com",
				AutoSend:    true,
			},
			wantJSON: `{"account":"test-account","email":"user@example.com",` +
				`"device_email":"device@example.com","auto_send":true}`,
		},
		{
			name: "minimal profile",
			resp: UserProfileResponse{
				Account: "test-account",
				Email:   "user@example.com",
			},
			wantJSON: `{"account":"test-account","email":"user@example.com",` +
				`"device_email":"","auto_send":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled UserProfileResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.Account, unmarshaled.Account)
			assert.Equal(t, tt.resp.Email, unmarshaled.Email)
			assert.Equal(t, tt.resp.DeviceEmail, unmarshaled.DeviceEmail)
			assert.Equal(t, tt.resp.AutoSend, unmarshaled.AutoSend)
		})
	}
}

func TestSendsResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     SendsResponse
		wantJSON string
	}{
		{
			name: "full response",
			resp: SendsResponse{
				TotalSends:        5,
				CurrentSends:      3,
				MaxSendsPerPeriod: 10,
				PeriodDays:        30,
				RemainingSends:    7,
			},
			wantJSON: `{"total_sends":5,"current_sends":3,"max_sends_per_period":10,` +
				`"period_days":30,"remaining_sends":7}`,
		},
		{
			name: "no sends",
			resp: SendsResponse{
				TotalSends:        0,
				CurrentSends:      0,
				MaxSendsPerPeriod: 10,
				PeriodDays:        30,
				RemainingSends:    10,
			},
			wantJSON: `{"total_sends":0,"current_sends":0,"max_sends_per_period":10,` +
				`"period_days":30,"remaining_sends":10}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled SendsResponse
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.TotalSends, unmarshaled.TotalSends)
			assert.Equal(t, tt.resp.CurrentSends, unmarshaled.CurrentSends)
			assert.Equal(t, tt.resp.MaxSendsPerPeriod, unmarshaled.MaxSendsPerPeriod)
			assert.Equal(t, tt.resp.PeriodDays, unmarshaled.PeriodDays)
			assert.Equal(t, tt.resp.RemainingSends, unmarshaled.RemainingSends)
		})
	}
}

func TestSendsResponseNoLimits_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		resp     SendsResponseNoLimits
		wantJSON string
	}{
		{
			name:     "with sends",
			resp:     SendsResponseNoLimits{TotalSends: 5},
			wantJSON: `{"total_sends":5}`,
		},
		{
			name:     "no sends",
			resp:     SendsResponseNoLimits{TotalSends: 0},
			wantJSON: `{"total_sends":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled SendsResponseNoLimits
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.resp.TotalSends, unmarshaled.TotalSends)
		})
	}
}
