package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/internal/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/server/types"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

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
	toggleFavoriteFunc func(
		ctx context.Context,
		accountID, articleID string,
	) (bool, error)
	getUserDeviceEmailFunc func(
		ctx context.Context,
		accountID string,
	) (string, bool, error)
	generateEPUBFunc func(article *model.Article) (io.ReadCloser, error)
	sendArticleFunc  func(
		ctx context.Context,
		destEmail string,
		epubData io.ReadCloser,
		title string,
	) (*email.SendEmailResponse, error)
	sendArticleByIDFunc func(
		ctx context.Context,
		accountID, articleID string,
	) (*servicetypes.SendArticleResult, error)
	countSendsFunc func(
		ctx context.Context,
		accountID string,
		startDate, endDate time.Time,
	) (int, error)
	dbError error
}

func (m *articleMockService) Fetch(
	_ context.Context,
	_ *url.URL,
) (*content.FetchedContent, error) {
	return nil, errors.New("not implemented")
}

func (m *articleMockService) ParseHTML(
	_ context.Context,
	_ *content.FetchedContent,
) (*html.Node, error) {
	return nil, errors.New("not implemented")
}

func (m *articleMockService) Clean(
	_ context.Context,
	_ *html.Node,
	_ *url.URL,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *articleMockService) GenerateEPUB(article *model.Article) (io.ReadCloser, error) {
	if m.generateEPUBFunc != nil {
		return m.generateEPUBFunc(article)
	}
	return io.NopCloser(bytes.NewReader([]byte("epub content"))), nil
}

func (m *articleMockService) SendArticle(
	ctx context.Context, destEmail string, epubData io.ReadCloser, title string) (*email.SendEmailResponse, error) {
	if m.sendArticleFunc != nil {
		return m.sendArticleFunc(ctx, destEmail, epubData, title)
	}
	if epubData != nil {
		_ = epubData.Close()
	}
	return &email.SendEmailResponse{MessageID: "test-msg-id"}, nil
}

func (m *articleMockService) SendArticleByID(
	ctx context.Context, accountID, articleID string) (*servicetypes.SendArticleResult, error) {
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

func (m *articleMockService) GetArticlesMetadata(
	ctx context.Context,
	accountID string,
	page, pageSize int,
	favoriteFilter *bool,
) (*servicetypes.GetArticlesResult, error) {
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

func (m *articleMockService) DeleteArticle(
	ctx context.Context, accountID, articleID string) (*servicetypes.DeleteArticleResult, error) {
	if m.deleteArticleFunc != nil {
		return m.deleteArticleFunc(ctx, accountID, articleID)
	}
	return &servicetypes.DeleteArticleResult{Deleted: 1}, nil
}

func (m *articleMockService) GetDBError() error {
	return m.dbError
}

func (m *articleMockService) GetUserDeviceEmailAndAutoSend(
	ctx context.Context, accountID string) (deviceEmail string, autoSend bool, err error) {
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
	ctx context.Context,
	accountID string,
	startDate, endDate time.Time,
) (int, error) {
	if m.countSendsFunc != nil {
		return m.countSendsFunc(ctx, accountID, startDate, endDate)
	}
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
	ctx = logging.WithLogRecordValue(ctx, logRecord)
	ctx = logging.WithRequestError(ctx)

	return ctx
}

func newArticleTestContextWithAccount() context.Context {
	ctx := newArticleTestContext()
	return auth.AddAccountIDToCtx(ctx, "account-123")
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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc)

	body := types.ArticleRequest{URL: "https://example.com/article"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCreateArticle(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp types.ArticleResponse
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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc)

	body := types.ArticleRequest{}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCreateArticle(w, req)

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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc)

	body := types.ArticleRequest{URL: "not-a-valid-url"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCreateArticle(w, req)

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

			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
			h := New(cfg, mockSvc, http.DefaultClient, mockProc)

			body := types.ArticleRequest{URL: "https://example.com/article"}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"POST", "/v1/articles", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.HandleCreateArticle(w, req)

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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc)

	body := types.ArticleRequest{URL: "https://example.com/article"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCreateArticle(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, mockProc.startProcessingCalled)
}

func TestHandleGetArticles_Success(t *testing.T) {
	mockSvc := &articleMockService{
		getArticlesMetadataFunc: func(
			_ context.Context,
			_ string,
			page,
			pageSize int,
			_ *bool,
		) (*servicetypes.GetArticlesResult, error) {
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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles?page=1&page_size=10",
		nil)
	w := httptest.NewRecorder()

	h.HandleGetArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.ListArticlesResponse
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
		getArticlesMetadataFunc: func(
			_ context.Context,
			_ string,
			_, _ int,
			favoriteFilter *bool,
		) (*servicetypes.GetArticlesResult, error) {
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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles?favorite=true",
		nil)
	w := httptest.NewRecorder()

	h.HandleGetArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetArticles_InvalidPage(t *testing.T) {
	mockSvc := &articleMockService{
		getArticlesMetadataFunc: func(
			_ context.Context,
			_ string,
			page,
			_ int,
			_ *bool,
		) (*servicetypes.GetArticlesResult, error) {
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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles?page=invalid",
		nil)
	w := httptest.NewRecorder()

	h.HandleGetArticles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetArticles_PageSizeCapped(t *testing.T) {
	mockSvc := &articleMockService{
		getArticlesMetadataFunc: func(
			_ context.Context,
			_ string,
			_,
			pageSize int,
			_ *bool,
		) (*servicetypes.GetArticlesResult, error) {
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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles?page_size=200",
		nil)
	w := httptest.NewRecorder()

	h.HandleGetArticles(w, req)

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
				getArticlesMetadataFunc: func(
					_ context.Context,
					_ string,
					_, _ int,
					_ *bool,
				) (*servicetypes.GetArticlesResult, error) {
					return nil, tt.serviceErr
				},
			}

			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
			h := New(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(), "GET", "/v1/articles", nil)
			w := httptest.NewRecorder()

			h.HandleGetArticles(w, req)

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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles/article-123",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleGetArticle(w, req)

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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles/article-123",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleGetArticle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDeleteArticle_Success(t *testing.T) {
	mockSvc := &articleMockService{
		deleteArticleFunc: func(_ context.Context, _, _ string) (*servicetypes.DeleteArticleResult, error) {
			return &servicetypes.DeleteArticleResult{Deleted: 1}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"DELETE",
		"/v1/articles/article-123",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleDeleteArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.DeleteArticleResponse
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

			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
			h := New(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"DELETE",
				"/v1/articles/article-123",
				nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "article-123")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			h.HandleDeleteArticle(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleToggleFavorite_Success(t *testing.T) {
	mockSvc := &articleMockService{
		toggleFavoriteFunc: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"PUT",
		"/v1/articles/article-123/favorite",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleToggleFavorite(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.FavoriteResponse
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

			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
			h := New(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"PUT",
				"/v1/articles/article-123/favorite",
				nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "article-123")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			h.HandleToggleFavorite(w, req)

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
		generateEPUBFunc: func(_ *model.Article) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("epub content"))), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: "msg-123"}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.SendArticleResponse
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
		generateEPUBFunc: func(_ *model.Article) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("epub content"))), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: "msg-123"}, nil
		},
		countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
			return 5, nil
		},
	}

	cfg := &config.Config{
		AuthBackend:   consts.AuthBackendAuth0,
		EmailProvider: consts.EmailBackendMailjet,
	}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	ctx := newArticleTestContextWithAccount()
	req := httptest.NewRequestWithContext(ctx, "POST", "/v1/articles/article-123/send", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.SendArticleResponseWithCount
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "sent", resp.Status)
	assert.Equal(t, 6, resp.SendsCount)
}

func TestHandleCreateArticle_SendOnCompleteWithoutEmailBackend(t *testing.T) {
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

	cfg := &config.Config{EmailProvider: ""}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc)

	body := types.ArticleRequest{URL: "https://example.com/article", SendOnComplete: true}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCreateArticle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "email backend not configured: invalid input", errResp.Error)

	assert.False(t, mockProc.startProcessingCalled)
}

func TestHandleSendArticle_ArticleNotFound(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, apperrors.ErrNotFound
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleSendArticle_GenerateEPUBError(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, errors.New("epub generation failed")
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSendArticle_GetDeviceEmailError(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, apperrors.ErrNotFound
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleSendArticle_NoDeviceEmail(t *testing.T) {
	mockSvc := &articleMockService{
		sendArticleByIDFunc: func(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
			return nil, apperrors.ErrInvalid
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

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

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

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
		generateEPUBFunc: func(_ *model.Article) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("epub content"))), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: "msg-123"}, nil
		},
		dbError: errors.New("database error"),
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "article-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
