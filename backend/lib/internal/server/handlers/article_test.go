package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/internal/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/types"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	internaltype "github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

const (
	testArticleDeviceEmail = "device@kindle.com"
	articleTestErrNotFound = "ErrNotFound"
	testEmailMsgID         = "test-msg-id"
)

type mockProcessor struct {
	startProcessingCalled bool
	startProcessingEvent  *content.ProcessArticleEvent
}

const (
	testArticleTitle = "Test Article"
	testArticleID    = "article-123"
	testArticleURL   = "https://example.com/article"
	tagTech          = "tech"
	tagProgramming   = "programming"
	tagGolang        = "golang"
	testMessageID    = "msg-123"
	testOldTag       = "old-tag"
	testNewTag1      = "new-tag-1"
	testNewTag2      = "new-tag-2"
)

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
		filter *internaltype.ArticleFilter,
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
	addTagsFunc    func(ctx context.Context, accountID, articleID string, tags []string) error
	removeTagsFunc func(ctx context.Context, accountID, articleID string, tags []string) error
	setTagsFunc    func(ctx context.Context, accountID, articleID string, tags []string) error
	getTagsFunc    func(ctx context.Context, accountID, articleID string) ([]string, error)
	getAllTagsFunc func(ctx context.Context, accountID string) ([]string, error)
	dbError        error
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

func (m *articleMockService) ReadEPUB(_ context.Context, _ *url.URL) (io.ReadCloser, string, error) {
	return io.NopCloser(bytes.NewReader([]byte("epub content"))), "Test EPUB", nil
}

func (m *articleMockService) SendArticle(
	ctx context.Context, destEmail string, epubData io.ReadCloser, title string) (*email.SendEmailResponse, error) {
	if m.sendArticleFunc != nil {
		return m.sendArticleFunc(ctx, destEmail, epubData, title)
	}
	if epubData != nil {
		_ = epubData.Close()
	}
	return &email.SendEmailResponse{MessageID: testEmailMsgID}, nil
}

func (m *articleMockService) SendArticleByID(
	ctx context.Context, accountID, articleID string) (*servicetypes.SendArticleResult, error) {
	if m.sendArticleByIDFunc != nil {
		return m.sendArticleByIDFunc(ctx, accountID, articleID)
	}
	return &servicetypes.SendArticleResult{
		Article: &model.Article{
			ID:    articleID,
			URL:   testArticleURL,
			Title: testArticleTitle,
		},
		DeviceEmail: testArticleDeviceEmail,
		EmailResp:   &email.SendEmailResponse{MessageID: testEmailMsgID},
	}, nil
}

func (m *articleMockService) CreateArticle(ctx context.Context, u *url.URL, accountID string) (*model.Article, error) {
	if m.createArticleFunc != nil {
		return m.createArticleFunc(ctx, u, accountID)
	}
	return &model.Article{
		ID:    testArticleID,
		URL:   testArticleURL,
		Title: testArticleTitle,
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
		URL:   testArticleURL,
		Title: testArticleTitle,
	}, nil
}

func (m *articleMockService) GetArticlesMetadata(
	ctx context.Context,
	accountID string,
	page, pageSize int,
	filter *internaltype.ArticleFilter,
) (*servicetypes.GetArticlesResult, error) {
	if m.getArticlesMetadataFunc != nil {
		return m.getArticlesMetadataFunc(ctx, accountID, page, pageSize, filter)
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

func (m *articleMockService) AddArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	if m.addTagsFunc != nil {
		return m.addTagsFunc(ctx, accountID, articleID, tags)
	}
	return nil
}

func (m *articleMockService) RemoveArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	if m.removeTagsFunc != nil {
		return m.removeTagsFunc(ctx, accountID, articleID, tags)
	}
	return nil
}

func (m *articleMockService) SetArticleTags(ctx context.Context, accountID, articleID string, tags []string) error {
	if m.setTagsFunc != nil {
		return m.setTagsFunc(ctx, accountID, articleID, tags)
	}
	return nil
}

func (m *articleMockService) GetArticleTags(ctx context.Context, accountID, articleID string) ([]string, error) {
	if m.getTagsFunc != nil {
		return m.getTagsFunc(ctx, accountID, articleID)
	}
	return []string{tagTech, tagProgramming}, nil
}

func (m *articleMockService) GetAllTagsForAccount(ctx context.Context, accountID string) ([]string, error) {
	if m.getAllTagsFunc != nil {
		return m.getAllTagsFunc(ctx, accountID)
	}
	return []string{tagTech, tagProgramming, tagGolang}, nil
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
				ID:    testArticleID,
				URL:   testArticleURL,
				Title: testArticleTitle,
			}, nil
		},
	}
	mockProc := &mockProcessor{}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc, nil)

	body := types.ArticleRequest{URL: testArticleURL}
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
	assert.Equal(t, testArticleID, resp.ID)
	assert.Equal(t, testArticleURL, resp.URL)

	assert.True(t, mockProc.startProcessingCalled)
	assert.NotNil(t, mockProc.startProcessingEvent)
	assert.Equal(t, testArticleURL, mockProc.startProcessingEvent.URL)
	assert.Equal(t, testArticleID, mockProc.startProcessingEvent.ArticleID)
	assert.Equal(t, "account-123", mockProc.startProcessingEvent.AccountID)
}

func TestHandleCreateArticle_SendOnComplete(t *testing.T) {
	mockSvc := &articleMockService{
		createArticleFunc: func(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
			return &model.Article{
				ID:    testArticleID,
				URL:   testArticleURL,
				Title: testArticleTitle,
			}, nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
			return 3, nil
		},
	}
	mockProc := &mockProcessor{}

	cfg := &config.Config{
		EmailProvider: consts.EmailBackendMailjet,
		AuthBackend:   consts.AuthBackendAuth0,
	}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc, nil)

	body := types.ArticleRequest{URL: testArticleURL, SendOnComplete: true}
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
	assert.Equal(t, testArticleID, resp.ID)
	assert.Equal(t, testArticleURL, resp.URL)

	assert.True(t, mockProc.startProcessingCalled)
	assert.NotNil(t, mockProc.startProcessingEvent)
	assert.Equal(t, testArticleURL, mockProc.startProcessingEvent.URL)
	assert.Equal(t, testArticleID, mockProc.startProcessingEvent.ArticleID)
	assert.Equal(t, "account-123", mockProc.startProcessingEvent.AccountID)
	assert.True(t, mockProc.startProcessingEvent.SendOnComplete)
}

func TestHandleCreateArticle_MissingURL(t *testing.T) {
	mockSvc := &articleMockService{}
	mockProc := &mockProcessor{}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc, nil)

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
	h := New(cfg, mockSvc, http.DefaultClient, mockProc, nil)

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
			name:           articleTestErrNotFound,
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
			name:           genericErrorMessage,
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
			h := New(cfg, mockSvc, http.DefaultClient, mockProc, nil)

			body := types.ArticleRequest{URL: testArticleURL}
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
				ID:    testArticleID,
				URL:   testArticleURL,
				Title: testArticleTitle,
			}, nil
		},
	}
	mockProc := &mockProcessor{}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc, nil)

	body := types.ArticleRequest{URL: testArticleURL}
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
			_ *internaltype.ArticleFilter,
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

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
			filter *internaltype.ArticleFilter,
		) (*servicetypes.GetArticlesResult, error) {
			assert.NotNil(t, filter)
			assert.True(t, *filter.Favorite)
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

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
			_ *internaltype.ArticleFilter,
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

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
			_ *internaltype.ArticleFilter,
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

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
			name:           articleTestErrNotFound,
			serviceErr:     apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           genericErrorMessage,
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
					_ *internaltype.ArticleFilter,
				) (*servicetypes.GetArticlesResult, error) {
					return nil, tt.serviceErr
				},
			}

			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
			h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(), "GET", "/v1/articles", nil)
			w := httptest.NewRecorder()

			h.HandleGetArticles(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleGetArticles_InvalidTag(t *testing.T) {
	tests := []struct {
		name         string
		tag          string
		expectedCode int
	}{
		{
			name:         "empty tag",
			tag:          "",
			expectedCode: http.StatusOK, // Empty tag should be ignored
		},
		{
			name:         "tag with invalid characters",
			tag:          "tag with spaces!",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "tag that normalizes to empty",
			tag:          "   ", // Only whitespace
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}

			mockSvc := &articleMockService{
				getArticlesMetadataFunc: func(
					_ context.Context,
					_ string,
					_, _ int,
					_ *internaltype.ArticleFilter,
				) (*servicetypes.GetArticlesResult, error) {
					return &servicetypes.GetArticlesResult{
						Articles: []*model.Article{},
						Page:     1,
						PageSize: 10,
						Total:    0,
						HasMore:  false,
					}, nil
				},
			}

			h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"GET",
				"/v1/articles?tag="+url.QueryEscape(tt.tag),
				nil)
			w := httptest.NewRecorder()

			h.HandleGetArticles(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestHandleGetArticles_PaginationEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		page         string
		pageSize     string
		expectedPage int
		expectedSize int
	}{
		{
			name:         "default pagination",
			page:         "",
			pageSize:     "",
			expectedPage: 1,
			expectedSize: 20, // consts.DefaultPageSize
		},
		{
			name:         "explicit page and size",
			page:         "3",
			pageSize:     "20",
			expectedPage: 3,
			expectedSize: 20,
		},
		{
			name:         "page zero uses default",
			page:         "0",
			pageSize:     "10",
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "negative page uses default",
			page:         "-1",
			pageSize:     "10",
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "page size exceeds max",
			page:         "1",
			pageSize:     "200",
			expectedPage: 1,
			expectedSize: 20, // consts.MaxPageSize
		},
		{
			name:         "page size below min",
			page:         "1",
			pageSize:     "0",
			expectedPage: 1,
			expectedSize: 20, // Default
		},
		{
			name:         "invalid page size ignored",
			page:         "1",
			pageSize:     "invalid",
			expectedPage: 1,
			expectedSize: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calledPage := -1
			calledSize := -1

			mockSvc := &articleMockService{
				getArticlesMetadataFunc: func(
					_ context.Context,
					_ string,
					page, pageSize int,
					_ *internaltype.ArticleFilter,
				) (*servicetypes.GetArticlesResult, error) {
					calledPage = page
					calledSize = pageSize
					return &servicetypes.GetArticlesResult{
						Articles: []*model.Article{},
						Page:     page,
						PageSize: pageSize,
						Total:    0,
						HasMore:  false,
					}, nil
				},
			}

			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
			h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"GET",
				fmt.Sprintf("/v1/articles?page=%s&page_size=%s", tt.page, tt.pageSize),
				nil)
			w := httptest.NewRecorder()

			h.HandleGetArticles(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedPage, calledPage)
			assert.Equal(t, tt.expectedSize, calledSize)
		})
	}
}

func TestHandleGetArticles_EmptyResults(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		description string
	}{
		{
			name:        "empty result with favorite and tag",
			queryParams: "favorite=true&tag=nonexistent",
			description: "No articles match both criteria",
		},
		{
			name:        "empty result with tag only",
			queryParams: "tag=nonexistent",
			description: "No articles with this tag",
		},
		{
			name:        "empty result with favorite only",
			queryParams: "favorite=true",
			description: "No favorite articles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &articleMockService{
				getArticlesMetadataFunc: func(
					_ context.Context,
					_ string,
					_, _ int,
					_ *internaltype.ArticleFilter,
				) (*servicetypes.GetArticlesResult, error) {
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
			h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"GET",
				"/v1/articles?"+tt.queryParams,
				nil)
			w := httptest.NewRecorder()

			h.HandleGetArticles(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp types.ListArticlesResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Empty(t, resp.Articles)
			assert.Zero(t, resp.Total)
			assert.False(t, resp.HasMore)
		})
	}
}

func TestHandleGetArticle_Success(t *testing.T) {
	mockSvc := &articleMockService{
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				Title: testArticleTitle,
				URL:   testArticleURL,
			}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles/article-123",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleGetArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp model.Article
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testArticleID, resp.ID)
	assert.Equal(t, testArticleTitle, resp.Title)
}

func TestHandleGetArticle_NotFound(t *testing.T) {
	mockSvc := &articleMockService{
		getArticleFunc: func(_ context.Context, _, _ string) (*model.Article, error) {
			return nil, apperrors.ErrNotFound
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles/article-123",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"DELETE",
		"/v1/articles/article-123",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
			name:           articleTestErrNotFound,
			serviceErr:     apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           genericErrorMessage,
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
			h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"DELETE",
				"/v1/articles/article-123",
				nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", testArticleID)
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"PUT",
		"/v1/articles/article-123/favorite",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
			name:           articleTestErrNotFound,
			serviceErr:     apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           genericErrorMessage,
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
			h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"PUT",
				"/v1/articles/article-123/favorite",
				nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", testArticleID)
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
				Title: testArticleTitle,
			}, nil
		},
		generateEPUBFunc: func(_ *model.Article) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("epub content"))), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: testMessageID}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
				Title: testArticleTitle,
			}, nil
		},
		generateEPUBFunc: func(_ *model.Article) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("epub content"))), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: testMessageID}, nil
		},
		countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
			return 5, nil
		},
	}

	cfg := &config.Config{
		AuthBackend:   consts.AuthBackendAuth0,
		EmailProvider: consts.EmailBackendMailjet,
	}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	ctx := newArticleTestContextWithAccount()
	req := httptest.NewRequestWithContext(ctx, "POST", "/v1/articles/article-123/send", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
				ID:    testArticleID,
				URL:   testArticleURL,
				Title: testArticleTitle,
			}, nil
		},
	}
	mockProc := &mockProcessor{}

	cfg := &config.Config{EmailProvider: ""}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc, nil)

	body := types.ArticleRequest{URL: testArticleURL, SendOnComplete: true}
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
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
				Title: testArticleTitle,
			}, nil
		},
		generateEPUBFunc: func(_ *model.Article) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("epub content"))), nil
		},
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: testMessageID}, nil
		},
		dbError: errors.New("database error"),
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"POST",
		"/v1/articles/article-123/send",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSendArticle(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Tag endpoint tests

func TestHandleAddTags_Success(t *testing.T) {
	mockSvc := &articleMockService{
		addTagsFunc: func(_ context.Context, _, _ string, tags []string) error {
			assert.ElementsMatch(t, []string{tagTech, tagProgramming}, tags)
			return nil
		},
		getTagsFunc: func(_ context.Context, _, _ string) ([]string, error) {
			return []string{tagProgramming, tagTech, testOldTag}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	body := types.TagsRequest{Tags: []string{tagTech, tagProgramming}}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "POST", "/v1/articles/article-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleAddTags(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{tagProgramming, tagTech, testOldTag}, resp.Tags)
}

func TestHandleAddTags_InvalidTags(t *testing.T) {
	mockSvc := &articleMockService{
		addTagsFunc: func(_ context.Context, _, _ string, tags []string) error {
			// Simulate service layer validation - invalid tags should fail
			if len(tags) > 0 {
				if slices.Contains(tags, "tag with special characters!@#") {
					return fmt.Errorf("failed to validate tags: %w", apperrors.ErrInvalid)
				}
			}
			return nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	body := types.TagsRequest{Tags: []string{"tag with special characters!@#"}}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "POST", "/v1/articles/article-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleAddTags(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "failed to validate tags")
}

func TestHandleAddTags_TooManyTags(t *testing.T) {
	mockSvc := &articleMockService{
		addTagsFunc: func(_ context.Context, _, _ string, tags []string) error {
			// Simulate service layer validation - too many tags should fail
			if len(tags) > 10 {
				return fmt.Errorf("failed to validate tags: %w", apperrors.ErrInvalid)
			}
			return nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	// Create 11 tags (max is 10)
	tags := make([]string, 11)
	for i := range 11 {
		tags[i] = "tag" + strconv.Itoa(i)
	}

	body := types.TagsRequest{Tags: tags}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "POST", "/v1/articles/article-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleAddTags(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp model.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "failed to validate tags")
}

func TestHandleSetTags_Success(t *testing.T) {
	mockSvc := &articleMockService{
		setTagsFunc: func(_ context.Context, _, _ string, tags []string) error {
			assert.ElementsMatch(t, []string{testNewTag1, testNewTag2}, tags)
			return nil
		},
		getTagsFunc: func(_ context.Context, _, _ string) ([]string, error) {
			return []string{testNewTag1, testNewTag2}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	body := types.TagsRequest{Tags: []string{testNewTag1, testNewTag2}}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "PUT", "/v1/articles/article-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSetTags(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{testNewTag1, testNewTag2}, resp.Tags)
}

func TestHandleSetTags_EmptyTags(t *testing.T) {
	mockSvc := &articleMockService{
		setTagsFunc: func(_ context.Context, _, _ string, tags []string) error {
			assert.Empty(t, tags)
			return nil
		},
		getTagsFunc: func(_ context.Context, _, _ string) ([]string, error) {
			return []string{}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	body := types.TagsRequest{Tags: []string{}}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "PUT", "/v1/articles/article-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleSetTags(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Tags)
}

func TestHandleGetTags_Success(t *testing.T) {
	mockSvc := &articleMockService{
		getTagsFunc: func(_ context.Context, _, _ string) ([]string, error) {
			return []string{tagProgramming, tagTech}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles/article-123/tags",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleGetTags(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resp.Tags))
	assert.Contains(t, resp.Tags, tagProgramming)
	assert.Contains(t, resp.Tags, tagTech)
}

func TestHandleGetTags_Empty(t *testing.T) {
	mockSvc := &articleMockService{
		getTagsFunc: func(_ context.Context, _, _ string) ([]string, error) {
			return []string{}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/articles/article-123/tags",
		nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleGetTags(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Tags)
}

func TestHandleGetTags_ServiceError(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           articleTestErrNotFound,
			serviceErr:     apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           genericErrorMessage,
			serviceErr:     errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &articleMockService{
				getTagsFunc: func(_ context.Context, _, _ string) ([]string, error) {
					return nil, tt.serviceErr
				},
			}

			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
			h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
				"GET", "/v1/articles/article-123/tags", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", testArticleID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			h.HandleGetTags(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleRemoveTags_Success(t *testing.T) {
	mockSvc := &articleMockService{
		removeTagsFunc: func(_ context.Context, _, _ string, tags []string) error {
			assert.ElementsMatch(t, []string{testOldTag}, tags)
			return nil
		},
		getTagsFunc: func(_ context.Context, _, _ string) ([]string, error) {
			return []string{tagProgramming, tagTech}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	body := types.TagsRequest{Tags: []string{testOldTag}}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newArticleTestContextWithAccount(), "DELETE", "/v1/articles/article-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testArticleID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.HandleRemoveTags(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{tagProgramming, tagTech}, resp.Tags)
}

func TestHandleGetAllTags_Success(t *testing.T) {
	mockSvc := &articleMockService{
		getAllTagsFunc: func(_ context.Context, _ string) ([]string, error) {
			return []string{tagGolang, tagProgramming, tagTech}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/tags",
		nil)
	w := httptest.NewRecorder()

	h.HandleGetAllTags(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 3, len(resp.Tags))
	assert.Contains(t, resp.Tags, tagGolang)
	assert.Contains(t, resp.Tags, tagProgramming)
	assert.Contains(t, resp.Tags, tagTech)
}

func TestHandleGetAllTags_Empty(t *testing.T) {
	mockSvc := &articleMockService{
		getAllTagsFunc: func(_ context.Context, _ string) ([]string, error) {
			return []string{}, nil
		},
	}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(),
		"GET",
		"/v1/tags",
		nil)
	w := httptest.NewRecorder()

	h.HandleGetAllTags(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Tags)
}

func TestHandleGetAllTags_ServiceError(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           genericErrorMessage,
			serviceErr:     errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &articleMockService{
				getAllTagsFunc: func(_ context.Context, _ string) ([]string, error) {
					return nil, tt.serviceErr
				},
			}

			cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
			h := New(cfg, mockSvc, http.DefaultClient, nil, nil)

			req := httptest.NewRequestWithContext(newArticleTestContextWithAccount(), "GET", "/v1/tags", nil)
			w := httptest.NewRecorder()

			h.HandleGetAllTags(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleCreateArticle_WithTags(t *testing.T) {
	mockSvc := &articleMockService{
		createArticleFunc: func(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
			return &model.Article{
				ID:    testArticleID,
				URL:   testArticleURL,
				Title: testArticleTitle,
			}, nil
		},
		setTagsFunc: func(_ context.Context, _, _ string, tags []string) error {
			assert.ElementsMatch(t, []string{tagTech, tagProgramming}, tags)
			return nil
		},
	}
	mockProc := &mockProcessor{}

	cfg := &config.Config{EmailProvider: consts.EmailBackendMailjet}
	h := New(cfg, mockSvc, http.DefaultClient, mockProc, nil)

	body := types.ArticleRequest{
		URL:  testArticleURL,
		Tags: []string{tagTech, tagProgramming},
	}
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
	assert.Equal(t, testArticleID, resp.ID)

	assert.True(t, mockProc.startProcessingCalled)
}
