package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/auth"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractInheritedLogAttrs(t *testing.T) {
	tests := []struct {
		name         string
		recordAttrs  []slog.Attr
		expectedKeys []string
		excludedKeys []string
	}{
		{
			name: "excludes client_ip, user_agent, path, method, url",
			recordAttrs: []slog.Attr{
				slog.String("client_ip", "192.168.1.1"),
				slog.String("user_agent", "test-agent"),
				slog.String("path", "/test"),
				slog.String("method", "GET"),
				slog.String("request_id", "req-123"),
				slog.String("version", "1.0.0"),
				slog.String("account_id", "acc-456"),
				slog.String("url", "https://example.com"),
				slog.String("article_id", "art-789"),
			},
			expectedKeys: []string{"request_id", "version", "account_id", "article_id"},
			excludedKeys: []string{"client_ip", "user_agent", "path", "method", "url"},
		},
		{
			name: "empty record returns nil",
			recordAttrs: []slog.Attr{
				slog.String("client_ip", "192.168.1.1"),
				slog.String("user_agent", "test-agent"),
				slog.String("path", "/test"),
				slog.String("method", "GET"),
			},
			expectedKeys: []string{},
			excludedKeys: []string{"client_ip", "user_agent", "path", "method"},
		},
		{
			name: "partial attrs with url excluded",
			recordAttrs: []slog.Attr{
				slog.String("request_id", "req-123"),
				slog.String("url", "https://example.com"),
				slog.String("account_id", "acc-456"),
			},
			expectedKeys: []string{"request_id", "account_id"},
			excludedKeys: []string{"url"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
			for _, attr := range tt.recordAttrs {
				record.AddAttrs(attr)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &record})

			attrs := extractInheritedLogAttrs(ctx)

			attrMap := make(map[string]string)
			for _, attr := range attrs {
				attrMap[attr.Key] = attr.Value.String()
			}

			for _, key := range tt.expectedKeys {
				assert.Contains(t, attrMap, key, "expected key %s to be present", key)
			}
			for _, key := range tt.excludedKeys {
				assert.NotContains(t, attrMap, key, "expected key %s to be excluded", key)
			}
		})
	}
}

func TestExtractInheritedLogAttrs_NoLogRecord(t *testing.T) {
	ctx := context.Background()
	attrs := extractInheritedLogAttrs(ctx)

	assert.Nil(t, attrs)
}

func TestExtractInheritedLogAttrs_NilLogRecord(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, (*logging.LogRecord)(nil))

	attrs := extractInheritedLogAttrs(ctx)

	assert.Nil(t, attrs)
}

const (
	testArticleDeviceEmail = "device@kindle.com"
)

type mockProcessor struct {
	startProcessingCalled bool
	startProcessingEvent  *content.ProcessArticleEvent
}

func (m *mockProcessor) StartProcessing(
	_ context.Context,
	event *content.ProcessArticleEvent,
) {
	m.startProcessingCalled = true
	m.startProcessingEvent = event
}

type articleMockService struct {
	createArticleFunc func(
		ctx context.Context,
		u *url.URL, accountID string,
	) (*model.Article, error)
	getArticleFunc func(
		ctx context.Context,
		accountID, articleID string,
	) (*model.Article, error)
	getArticlesMetadataFunc func(
		ctx context.Context,
		accountID string,
		page, pageSize int,
		favoriteFilter *bool,
	) (*servicetypes.GetArticlesResult, error)
	deleteArticleFunc func(
		ctx context.Context,
		accountID, articleID string,
	) (*servicetypes.DeleteArticleResult, error)
	deleteAllArticlesFunc func(
		ctx context.Context,
		accountID string,
	) (*servicetypes.DeleteArticleResult, error)
	toggleFavoriteFunc func(
		ctx context.Context,
		accountID, articleID string,
	) (bool, error)
	getUserDeviceEmailFunc func(
		ctx context.Context,
		accountID string,
	) (string, bool, error)
	generateEPUBFunc func(article *model.Article) ([]byte, error)
	sendArticleFunc  func(
		ctx context.Context,
		destEmail string,
		epubBytes []byte,
		title string,
	) (*email.SendEmailResponse, error)
	sendArticleByIDFunc func(
		ctx context.Context,
		accountID, articleID string,
	) (*servicetypes.SendArticleResult, error)
	dbError error
}

func (m *articleMockService) Fetch(
	_ context.Context,
	_ *url.URL,
) (*content.FetchedContent, error) {
	return nil, errors.New("not implemented")
}

func (m *articleMockService) Extract(
	_ context.Context,
	_ *content.FetchedContent,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *articleMockService) GenerateEPUB(article *model.Article) ([]byte, error) {
	if m.generateEPUBFunc != nil {
		return m.generateEPUBFunc(article)
	}
	return []byte("epub content"), nil
}

func (m *articleMockService) SendArticle(ctx context.Context, destEmail string, epubBytes []byte, title string) (*email.SendEmailResponse, error) { //nolint:lll // long function signature
	if m.sendArticleFunc != nil {
		return m.sendArticleFunc(ctx, destEmail, epubBytes, title)
	}
	return &email.SendEmailResponse{MessageID: "test-msg-id"}, nil
}

func (m *articleMockService) SendArticleByID(ctx context.Context, accountID, articleID string) (*servicetypes.SendArticleResult, error) { //nolint:lll // long function signature
	if m.sendArticleByIDFunc != nil {
		return m.sendArticleByIDFunc(ctx, accountID, articleID)
	}
	return &servicetypes.SendArticleResult{
		Article: &model.Article{
			ID:    articleID,
			URL:   "https://example.com/article",
			Title: "Test Article",
		},
		DeviceEmail: testArticleDeviceEmail,
		EmailResp:   &email.SendEmailResponse{MessageID: "test-msg-id"},
	}, nil
}

func (m *articleMockService) CreateArticle(ctx context.Context, u *url.URL, accountID string) (*model.Article, error) {
	if m.createArticleFunc != nil {
		return m.createArticleFunc(ctx, u, accountID)
	}
	return &model.Article{
		ID:    "article-123",
		URL:   "https://example.com/article",
		Title: "Test Article",
	}, nil
}

func (m *articleMockService) UpdateArticle(_ context.Context, _ *model.Article) error {
	return nil
}

func (m *articleMockService) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	if m.getArticleFunc != nil {
		return m.getArticleFunc(ctx, accountID, articleID)
	}
	return &model.Article{
		ID:    articleID,
		URL:   "https://example.com/article",
		Title: "Test Article",
	}, nil
}

func (m *articleMockService) GetArticlesMetadata(ctx context.Context, accountID string, page, pageSize int, favoriteFilter *bool) (*servicetypes.GetArticlesResult, error) { //nolint:lll // long function signature
	if m.getArticlesMetadataFunc != nil {
		return m.getArticlesMetadataFunc(ctx, accountID, page, pageSize, favoriteFilter)
	}
	return &servicetypes.GetArticlesResult{
		Articles: []*model.Article{},
		Page:     page,
		PageSize: pageSize,
		Total:    0,
		HasMore:  false,
	}, nil
}

func (m *articleMockService) DeleteArticle(ctx context.Context, accountID, articleID string) (*servicetypes.DeleteArticleResult, error) { //nolint:lll // long function signature
	if m.deleteArticleFunc != nil {
		return m.deleteArticleFunc(ctx, accountID, articleID)
	}
	return &servicetypes.DeleteArticleResult{Deleted: 1}, nil
}

func (m *articleMockService) DeleteAllArticles(ctx context.Context, accountID string) (*servicetypes.DeleteArticleResult, error) { //nolint:lll // long function signature
	if m.deleteAllArticlesFunc != nil {
		return m.deleteAllArticlesFunc(ctx, accountID)
	}
	return &servicetypes.DeleteArticleResult{Deleted: 5}, nil
}

func (m *articleMockService) GetDBError() error {
	return m.dbError
}

func (m *articleMockService) GetUserDeviceEmail(ctx context.Context, accountID string) (deviceEmail string, autoSend bool, err error) { //nolint:lll // long function signature
	if m.getUserDeviceEmailFunc != nil {
		return m.getUserDeviceEmailFunc(ctx, accountID)
	}
	return testArticleDeviceEmail, false, nil
}

func (m *articleMockService) SetUserDeviceEmailWithAutoSend(
	_ context.Context,
	_, _ string,
	_ bool,
) error {
	return errors.New("not implemented")
}

func (m *articleMockService) DeleteUserDeviceEmail(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *articleMockService) GetUserProfile(_ context.Context, _ string) (*model.UserProfile, error) {
	return nil, errors.New("not implemented")
}

func (m *articleMockService) SetUserEmail(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

func (m *articleMockService) DeleteUserProfile(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *articleMockService) ToggleFavorite(ctx context.Context, accountID, articleID string) (bool, error) {
	if m.toggleFavoriteFunc != nil {
		return m.toggleFavoriteFunc(ctx, accountID, articleID)
	}
	return true, nil
}

func (m *articleMockService) CountSendsByAccountDateRange(
	_ context.Context,
	_ string,
	_, _ time.Time,
) (int, error) {
	return 0, nil
}

func (m *articleMockService) HandleBounce(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

func (m *articleMockService) IsEmailBouncing(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *articleMockService) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	return "", errors.New("not implemented")
}

func newArticleTestContext() context.Context {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	logRecord := &logging.LogRecord{Record: &record}
	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, logRecord)

	var err error
	ctx = context.WithValue(ctx, logging.RequestErrorKey, &err)

	return ctx
}

func newArticleTestContextWithAccount(accountID string) context.Context {
	ctx := newArticleTestContext()
	return context.WithValue(ctx, auth.AccountIDKey, accountID)
}

func newArticleTestContextWithSendsCount(count int) context.Context {
	ctx := newArticleTestContextWithAccount("test-account")
	ctx = context.WithValue(ctx, auth.SendsCountKey, count)
	return ctx
}

func TestConvertSlogAttrsToMap(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := convertSlogAttrsToMap(nil)
		assert.Nil(t, result)
	})

	t.Run("empty slice returns empty slice", func(t *testing.T) {
		result := convertSlogAttrsToMap([]slog.Attr{})
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})

	t.Run("converts slog.Attr with string value", func(t *testing.T) {
		attrs := []slog.Attr{
			slog.String("key1", "value1"),
			slog.String("key2", "value2"),
		}

		result := convertSlogAttrsToMap(attrs)

		require.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "value1", result[0]["key1"])
		assert.Equal(t, "value2", result[1]["key2"])
	})

	t.Run("converts slog.Attr with int value", func(t *testing.T) {
		attrs := []slog.Attr{
			slog.Int("count", 42),
		}

		result := convertSlogAttrsToMap(attrs)

		require.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, int64(42), result[0]["count"])
	})

	t.Run("converts slog.Attr with bool value", func(t *testing.T) {
		attrs := []slog.Attr{
			slog.Bool("enabled", true),
		}

		result := convertSlogAttrsToMap(attrs)

		require.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, true, result[0]["enabled"])
	})
}

func TestHandleCreateArticle_Success(t *testing.T) {
	mockSvc := &articleMockService{
		createArticleFunc: func(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
			return &model.Article{
				ID:    "article-123",
				URL:   "https://example.com/article",
				Title: "Test Article",
			}, nil
		},
	}
	mockProc := &mockProcessor{}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, mockProc)

	body := articleRequest{URL: "https://example.com/article"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles", bytes.NewReader(bodyBytes)) //nolint:lll // long function signature
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp articleResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "article-123", resp.ID)
	assert.Equal(t, "Test Article", resp.Title)
	assert.Equal(t, "https://example.com/article", resp.URL)

	assert.True(t, mockProc.startProcessingCalled)
	assert.NotNil(t, mockProc.startProcessingEvent)
	assert.Equal(t, "https://example.com/article", mockProc.startProcessingEvent.URL)
	assert.Equal(t, "article-123", mockProc.startProcessingEvent.ArticleID)
	assert.Equal(t, "account-123", mockProc.startProcessingEvent.AccountID)
}

func TestHandleCreateArticle_MissingURL(t *testing.T) {
	mockSvc := &articleMockService{}
	mockProc := &mockProcessor{}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, mockProc)

	body := articleRequest{}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles", bytes.NewReader(bodyBytes)) //nolint:lll // long function signature
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "missing URL")

	assert.False(t, mockProc.startProcessingCalled)
}

func TestHandleCreateArticle_InvalidURL(t *testing.T) {
	mockSvc := &articleMockService{}
	mockProc := &mockProcessor{}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, mockProc)

	body := articleRequest{URL: "not-a-valid-url"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles", bytes.NewReader(bodyBytes)) //nolint:lll // long function signature
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "invalid URL")

	assert.False(t, mockProc.startProcessingCalled)
}

func TestHandleCreateArticle_ServiceError(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           "ErrNotFound",
			serviceErr:     apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ErrInvalid",
			serviceErr:     apperrors.ErrInvalid,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrUnauthorized",
			serviceErr:     apperrors.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrConflict",
			serviceErr:     apperrors.ErrConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "generic error",
			serviceErr:     errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &articleMockService{
				createArticleFunc: func(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
					return nil, tt.serviceErr
				},
			}
			mockProc := &mockProcessor{}

			cfg := &config.Config{}
			h := newHandlers(cfg, mockSvc, http.DefaultClient, mockProc)

			body := articleRequest{URL: "https://example.com/article"}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles", bytes.NewReader(bodyBytes)) //nolint:lll // long function signature
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.handleCreateArticle(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.False(t, mockProc.startProcessingCalled)
		})
	}
}

func TestHandleCreateArticle_DBError(t *testing.T) {
	mockSvc := &articleMockService{
		createArticleFunc: func(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
			return &model.Article{
				ID:    "article-123",
				URL:   "https://example.com/article",
				Title: "Test Article",
			}, nil
		},
	}
	mockProc := &mockProcessor{}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, mockProc)

	body := articleRequest{URL: "https://example.com/article"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles", bytes.NewReader(bodyBytes)) //nolint:lll // long function signature
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, mockProc.startProcessingCalled)
}

func TestHandleGetArticles_Success(t *testing.T) {
	mockSvc := &articleMockService{
		getArticlesMetadataFunc: func(_ context.Context, _ string, page, pageSize int, _ *bool) (*servicetypes.GetArticlesResult, error) { //nolint:lll // long function signature
			return &servicetypes.GetArticlesResult{
				Articles: []*model.Article{
					{ID: "article-1", Title: "Article 1", URL: "https://example.com/1"},
					{ID: "article-2", Title: "Article 2", URL: "https://example.com/2"},
				},
				Page:     page,
				PageSize: pageSize,
				Total:    10,
				HasMore:  true,
			}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "GET", "/v1/articles?page=1&page_size=10", nil) //nolint:lll // long function signature
	w := httptest.NewRecorder()

	h.handleGetArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp listArticlesResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resp.Articles))
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
	assert.Equal(t, 10, resp.Total)
	assert.True(t, resp.HasMore)
}

func TestHandleGetArticles_WithFavoriteFilter(t *testing.T) {
	mockSvc := &articleMockService{
		getArticlesMetadataFunc: func(_ context.Context, _ string, _, _ int, favoriteFilter *bool) (*servicetypes.GetArticlesResult, error) { //nolint:lll // long function signature
			assert.NotNil(t, favoriteFilter)
			assert.True(t, *favoriteFilter)
			return &servicetypes.GetArticlesResult{
				Articles: []*model.Article{},
				Page:     1,
				PageSize: 10,
				Total:    0,
				HasMore:  false,
			}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "GET", "/v1/articles?favorite=true", nil) //nolint:lll // long function signature
	w := httptest.NewRecorder()

	h.handleGetArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetArticles_InvalidPage(t *testing.T) {
	mockSvc := &articleMockService{
		getArticlesMetadataFunc: func(_ context.Context, _ string, page, _ int, _ *bool) (*servicetypes.GetArticlesResult, error) { //nolint:lll // long function signature
			assert.Equal(t, 1, page)
			return &servicetypes.GetArticlesResult{
				Articles: []*model.Article{},
				Page:     1,
				PageSize: 10,
				Total:    0,
				HasMore:  false,
			}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "GET", "/v1/articles?page=invalid", nil) //nolint:lll // long function signature
	w := httptest.NewRecorder()

	h.handleGetArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetArticles_PageSizeCapped(t *testing.T) {
	mockSvc := &articleMockService{
		getArticlesMetadataFunc: func(_ context.Context, _ string, _, pageSize int, _ *bool) (*servicetypes.GetArticlesResult, error) { //nolint:lll // long function signature
			assert.Equal(t, 20, pageSize)
			return &servicetypes.GetArticlesResult{
				Articles: []*model.Article{},
				Page:     1,
				PageSize: 20,
				Total:    0,
				HasMore:  false,
			}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "GET", "/v1/articles?page_size=200", nil) //nolint:lll // long function signature
	w := httptest.NewRecorder()

	h.handleGetArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetArticles_ServiceError(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           "ErrNotFound",
			serviceErr:     apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "generic error",
			serviceErr:     errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &articleMockService{
				getArticlesMetadataFunc: func(_ context.Context, _ string, _, _ int, _ *bool) (*servicetypes.GetArticlesResult, error) { //nolint:lll // long function signature
					return nil, tt.serviceErr
				},
			}

			cfg := &config.Config{}
			h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "GET", "/v1/articles", nil)
			w := httptest.NewRecorder()

			h.handleGetArticles(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleGetArticle_Success(t *testing.T) {
	mockSvc := &articleMockService{
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				Title: "Test Article",
				URL:   "https://example.com/article",
			}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "GET", "/v1/articles/article-123", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleGetArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp model.Article
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "article-123", resp.ID)
	assert.Equal(t, "Test Article", resp.Title)
}

func TestHandleGetArticle_NotFound(t *testing.T) {
	mockSvc := &articleMockService{
		getArticleFunc: func(_ context.Context, _, _ string) (*model.Article, error) {
			return nil, apperrors.ErrNotFound
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "GET", "/v1/articles/article-123", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleGetArticle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDeleteArticle_Success(t *testing.T) {
	mockSvc := &articleMockService{
		deleteArticleFunc: func(_ context.Context, _, _ string) (*servicetypes.DeleteArticleResult, error) {
			return &servicetypes.DeleteArticleResult{Deleted: 1}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "DELETE", "/v1/articles/article-123", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleDeleteArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp deleteArticleResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Deleted)
}

func TestHandleDeleteArticle_ServiceError(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           "ErrNotFound",
			serviceErr:     apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "generic error",
			serviceErr:     errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &articleMockService{
				deleteArticleFunc: func(_ context.Context, _, _ string) (*servicetypes.DeleteArticleResult, error) {
					return nil, tt.serviceErr
				},
			}

			cfg := &config.Config{}
			h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "DELETE", "/v1/articles/article-123", nil) //nolint:lll // long function signature
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "article-123")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			h.handleDeleteArticle(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleDeleteAllArticles_Success(t *testing.T) {
	mockSvc := &articleMockService{
		deleteAllArticlesFunc: func(_ context.Context, _ string) (*servicetypes.DeleteArticleResult, error) {
			return &servicetypes.DeleteArticleResult{Deleted: 5}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "DELETE", "/v1/articles", nil)
	w := httptest.NewRecorder()

	h.handleDeleteAllArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp deleteArticleResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 5, resp.Deleted)
}

func TestHandleDeleteAllArticles_ServiceError(t *testing.T) {
	mockSvc := &articleMockService{
		deleteAllArticlesFunc: func(_ context.Context, _ string) (*servicetypes.DeleteArticleResult, error) {
			return nil, errors.New("some error")
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "DELETE", "/v1/articles", nil)
	w := httptest.NewRecorder()

	h.handleDeleteAllArticles(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleToggleFavorite_Success(t *testing.T) {
	mockSvc := &articleMockService{
		toggleFavoriteFunc: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "PUT", "/v1/articles/article-123/favorite", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleToggleFavorite(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp favoriteResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Favorite)
}

func TestHandleToggleFavorite_ServiceError(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           "ErrNotFound",
			serviceErr:     apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "generic error",
			serviceErr:     errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &articleMockService{
				toggleFavoriteFunc: func(_ context.Context, _, _ string) (bool, error) {
					return false, tt.serviceErr
				},
			}

			cfg := &config.Config{}
			h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "PUT", "/v1/articles/article-123/favorite", nil) //nolint:lll // long function signature
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "article-123")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			h.handleToggleFavorite(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleSendArticle_Success(t *testing.T) {
	mockSvc := &articleMockService{
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				Title: "Test Article",
			}, nil
		},
		generateEPUBFunc: func(_ *model.Article) ([]byte, error) {
			return []byte("epub content"), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ []byte, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: "msg-123"}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles/article-123/send", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleSendArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp sendArticleResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "sent", resp.Status)
}

func TestHandleSendArticle_WithSendsCount(t *testing.T) {
	mockSvc := &articleMockService{
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				Title: "Test Article",
			}, nil
		},
		generateEPUBFunc: func(_ *model.Article) ([]byte, error) {
			return []byte("epub content"), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ []byte, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: "msg-123"}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	ctx := newArticleTestContextWithSendsCount(5)
	req := httptest.NewRequestWithContext(ctx, "POST", "/v1/articles/article-123/send", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleSendArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp sendArticleResponseWithCount
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "sent", resp.Status)
	assert.Equal(t, 6, resp.SendsCount)
}

func TestHandleSendArticle_ArticleNotFound(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, apperrors.ErrNotFound
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles/article-123/send", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleSendArticle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleSendArticle_GenerateEPUBError(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, errors.New("epub generation failed")
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles/article-123/send", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleSendArticle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSendArticle_GetDeviceEmailError(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, apperrors.ErrNotFound
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles/article-123/send", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleSendArticle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleSendArticle_NoDeviceEmail(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, apperrors.ErrInvalid
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles/article-123/send", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleSendArticle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "invalid input")
}

func TestHandleSendArticle_SendArticleError(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, errors.New("send failed")
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles/article-123/send", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleSendArticle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSendArticle_DBError(t *testing.T) {
	mockSvc := &articleMockService{
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				Title: "Test Article",
			}, nil
		},
		generateEPUBFunc: func(_ *model.Article) ([]byte, error) {
			return []byte("epub content"), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ []byte, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: "msg-123"}, nil
		},
		dbError: errors.New("database error"),
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount("account-123"), "POST", "/v1/articles/article-123/send", nil) //nolint:lll // long function signature
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.handleSendArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
