package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/email"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/service"
)

const (
	testArticleNotFoundError = "article not found"
	dbErrorLogKey            = "db_error"
)

type MockService struct {
	createFunc          func(context.Context, string, string) (*service.CreateArticleResult, error)
	processFunc         func(context.Context, string) (*service.ProcessResult, error)
	sendFunc            func(context.Context, *service.ProcessResult, string, string) (*email.SendEmailResponse, error)
	writeFunc           func(*service.ProcessResult, string) error
	getArticle          func(context.Context, string, string) (*model.Article, error)
	getArticlesMetadata func(context.Context, string, int, int, *bool) (*service.GetArticlesResult, error)
	deleteArticle       func(context.Context, string, string) (*service.DeleteArticleResult, error)
	deleteAllArticles   func(context.Context, string) (*service.DeleteArticleResult, error)
	getUserKindleEmail  func(context.Context, string) (string, error)
	setUserKindleEmail  func(context.Context, string, string) error
	deleteUserProfile   func(context.Context, string) error
	toggleFavorite      func(context.Context, string, string) (bool, error)
	dbError             error
}

func newMockService(
	createFunc func(context.Context, string, string) (*service.CreateArticleResult, error),
) *MockService {
	return &MockService{
		createFunc: createFunc,
	}
}

func (m *MockService) CreateArticle(
	ctx context.Context,
	_ string,
	_ string,
) (*service.CreateArticleResult, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, "", "")
	}
	return nil, nil
}

func (m *MockService) Process(ctx context.Context, url string) (*service.ProcessResult, error) {
	if m.processFunc != nil {
		return m.processFunc(ctx, url)
	}
	return nil, nil
}

func (m *MockService) Send(
	ctx context.Context,
	result *service.ProcessResult,
	subject, destEmail string,
) (*email.SendEmailResponse, error) {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, result, subject, destEmail)
	}
	return &email.SendEmailResponse{
		Status:    "success",
		Message:   "sent",
		EmailUUID: "test-uuid",
	}, nil
}

func (m *MockService) GetUserKindleEmail(ctx context.Context, accountID string) (string, error) {
	if m.getUserKindleEmail != nil {
		return m.getUserKindleEmail(ctx, accountID)
	}
	return "", nil
}

func (m *MockService) SetUserKindleEmail(ctx context.Context, accountID, kindleEmail string) error {
	if m.setUserKindleEmail != nil {
		return m.setUserKindleEmail(ctx, accountID, kindleEmail)
	}
	return nil
}

func (m *MockService) DeleteUserProfile(ctx context.Context, accountID string) error {
	if m.deleteUserProfile != nil {
		return m.deleteUserProfile(ctx, accountID)
	}
	return nil
}

func (m *MockService) WriteToFile(result *service.ProcessResult, outputPath string) error {
	if m.writeFunc != nil {
		return m.writeFunc(result, outputPath)
	}
	return nil
}

func (m *MockService) GetDBError() error {
	return m.dbError
}

func (m *MockService) GetArticle(
	ctx context.Context,
	accountID string,
	articleID string,
) (*model.Article, error) {
	if m.getArticle != nil {
		return m.getArticle(ctx, accountID, articleID)
	}
	return nil, &serviceError{msg: "article not found"}
}

func (m *MockService) GetArticlesMetadata(
	ctx context.Context,
	accountID string,
	page int,
	pageSize int,
	favoriteFilter *bool,
) (*service.GetArticlesResult, error) {
	if m.getArticlesMetadata != nil {
		return m.getArticlesMetadata(ctx, accountID, page, pageSize, favoriteFilter)
	}
	return &service.GetArticlesResult{
		Articles: []*model.Article{},
		Page:     1,
		PageSize: 20,
		Total:    0,
		HasMore:  false,
	}, nil
}

func (m *MockService) DeleteArticle(
	ctx context.Context,
	accountID string,
	articleID string,
) (*service.DeleteArticleResult, error) {
	if m.deleteArticle != nil {
		return m.deleteArticle(ctx, accountID, articleID)
	}
	return &service.DeleteArticleResult{Deleted: 1}, nil
}

func (m *MockService) DeleteAllArticles(
	ctx context.Context,
	accountID string,
) (*service.DeleteArticleResult, error) {
	if m.deleteAllArticles != nil {
		return m.deleteAllArticles(ctx, accountID)
	}
	return &service.DeleteArticleResult{Deleted: 0}, nil
}

func (m *MockService) ToggleFavorite(
	ctx context.Context,
	accountID string,
	articleID string,
) (bool, error) {
	if m.toggleFavorite != nil {
		return m.toggleFavorite(ctx, accountID, articleID)
	}
	return false, nil
}

func TestHandleHealth(t *testing.T) {
	h := newHandlers(nil, nil, nil)

	req := httptest.NewRequest("GET", "/health", http.NoBody)
	w := httptest.NewRecorder()

	h.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp healthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}
}

func TestHandleCreateArticleSuccessWithEmail(t *testing.T) {
	cfg := &config.Config{
		SenderEmail:      "sender@example.com",
		MailjetAPIKey:    "test-key",
		MailjetAPISecret: "test-secret",
	}
	svc := newMockService(func(_ context.Context, _ string, _ string) (*service.CreateArticleResult, error) {
		return &service.CreateArticleResult{
			Article: &model.Article{
				ID:    "test-id",
				Title: testArticleTitle,
				URL:   "https://example.com/article",
			},
			Message: "article sent to Kindle successfully",
			EmailResp: &email.SendEmailResponse{
				Status:    "success",
				Message:   "sent",
				EmailUUID: "test-uuid",
			},
		}, nil
	})
	h := newHandlers(cfg, svc, nil)

	body := articleRequest{URL: "https://example.com/article"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp articleResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Title != testArticleTitle {
		t.Errorf("expected title '%s', got '%s'", testArticleTitle, resp.Title)
	}
	if resp.URL != "https://example.com/article" {
		t.Errorf("expected URL 'https://example.com/article', got '%s'", resp.URL)
	}
	if resp.Message != "article sent to Kindle successfully" {
		t.Errorf("expected message 'article sent to Kindle successfully', got '%s'", resp.Message)
	}
}

func TestHandleCreateArticleSuccessWithoutEmail(t *testing.T) {
	cfg := &config.Config{}
	svc := newMockService(func(_ context.Context, _ string, _ string) (*service.CreateArticleResult, error) {
		return &service.CreateArticleResult{
			Article: &model.Article{
				ID:    "test-id",
				Title: testArticleTitle,
				URL:   "https://example.com/article",
			},
			Message:   "article saved (kindle email not configured)",
			EmailResp: nil,
		}, nil
	})
	h := newHandlers(cfg, svc, nil)

	body := articleRequest{URL: "https://example.com/article"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp articleResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Title != testArticleTitle {
		t.Errorf("expected title '%s', got '%s'", testArticleTitle, resp.Title)
	}
	if resp.Message != statusEmailDisabled {
		t.Errorf("expected message '%s', got '%s'", statusEmailDisabled, resp.Message)
	}
}

func TestHandleCreateArticleInvalidJSON(t *testing.T) {
	cfg := &config.Config{}
	h := newHandlers(cfg, nil, http.DefaultClient)

	req := httptest.NewRequest("POST", "/v1/articles", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedMsg := "failed to decode request body: invalid character 'i' looking for beginning of value"
	if resp.Error != expectedMsg {
		t.Errorf("expected message '%s', got '%s'", expectedMsg, resp.Error)
	}
}

func TestHandleCreateArticleMissingURL(t *testing.T) {
	cfg := &config.Config{}
	h := newHandlers(cfg, nil, http.DefaultClient)

	body := articleRequest{URL: ""}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != "missing URL in request body" {
		t.Errorf("expected message 'missing URL in request body', got '%s'", resp.Error)
	}
}

func TestHandleCreateArticleServiceError(t *testing.T) {
	cfg := &config.Config{}
	svc := newMockService(func(_ context.Context, _ string, _ string) (*service.CreateArticleResult, error) {
		return nil, &serviceError{msg: "extraction failed"}
	})
	h := newHandlers(cfg, svc, nil)

	body := articleRequest{URL: "https://example.com/article"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleCreateArticle(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != "extraction failed" {
		t.Errorf("expected message 'extraction failed', got '%s'", resp.Error)
	}
}

type serviceError struct {
	msg string
}

func (e *serviceError) Error() string {
	return e.msg
}

func TestHandleGetArticlesSuccess(t *testing.T) {
	cfg := &config.Config{}
	svc := newMockService(nil)
	now := time.Now()
	svc.getArticlesMetadata = func(
		_ context.Context, _ string, page, pageSize int, _ *bool,
	) (*service.GetArticlesResult, error) {
		articles := []*model.Article{
			{ID: "5", Title: "Article 5", URL: "https://example.com/5", CreatedAt: now},
			{ID: "4", Title: "Article 4", URL: "https://example.com/4", CreatedAt: now.Add(-1 * time.Hour)},
		}
		return &service.GetArticlesResult{
			Articles: articles,
			Page:     page,
			PageSize: pageSize,
			Total:    5,
			HasMore:  true,
		}, nil
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("GET", "/v1/articles?page=1&page_size=2", http.NoBody)
	w := httptest.NewRecorder()

	h.handleGetArticles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp listArticlesResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(resp.Articles))
	}
	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}
	if resp.PageSize != 2 {
		t.Errorf("expected page_size 2, got %d", resp.PageSize)
	}
	if resp.Total != 5 {
		t.Errorf("expected total 5, got %d", resp.Total)
	}
	if !resp.HasMore {
		t.Errorf("expected has_more true, got false")
	}
}

func TestHandleGetArticlesDefaultParams(t *testing.T) {
	cfg := &config.Config{}
	svc := newMockService(nil)
	svc.getArticlesMetadata = func(
		_ context.Context, _ string, page, pageSize int, _ *bool,
	) (*service.GetArticlesResult, error) {
		return &service.GetArticlesResult{
			Articles: []*model.Article{},
			Page:     page,
			PageSize: pageSize,
			Total:    0,
			HasMore:  false,
		}, nil
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("GET", "/v1/articles", http.NoBody)
	w := httptest.NewRecorder()

	h.handleGetArticles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp listArticlesResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Page != 1 {
		t.Errorf("expected default page 1, got %d", resp.Page)
	}
	if resp.PageSize != 20 {
		t.Errorf("expected default page_size 20, got %d", resp.PageSize)
	}
}

func TestHandleGetArticlesInvalidParams(t *testing.T) {
	cfg := &config.Config{}
	svc := newMockService(nil)
	svc.getArticlesMetadata = func(
		_ context.Context, _ string, page, pageSize int, _ *bool,
	) (*service.GetArticlesResult, error) {
		return &service.GetArticlesResult{
			Articles: []*model.Article{},
			Page:     page,
			PageSize: pageSize,
			Total:    0,
			HasMore:  false,
		}, nil
	}
	h := newHandlers(cfg, svc, nil)

	testCases := []struct {
		name         string
		url          string
		expectedPage int
		expectedSize int
	}{
		{"invalid page uses default", "/v1/articles?page=abc&page_size=10", 1, 10},
		{"negative page uses default", "/v1/articles?page=-1&page_size=10", 1, 10},
		{"zero page uses default", "/v1/articles?page=0&page_size=10", 1, 10},
		{"invalid size uses default", "/v1/articles?page=1&page_size=abc", 1, 20},
		{"size too small uses default", "/v1/articles?page=1&page_size=0", 1, 20},
		{"size too large uses max", "/v1/articles?page=1&page_size=200", 1, 20},
		{"negative size uses default", "/v1/articles?page=1&page_size=-10", 1, 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, http.NoBody)
			w := httptest.NewRecorder()

			h.handleGetArticles(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			var resp listArticlesResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Page != tc.expectedPage {
				t.Errorf("expected page %d, got %d", tc.expectedPage, resp.Page)
			}
			if resp.PageSize != tc.expectedSize {
				t.Errorf("expected page_size %d, got %d", tc.expectedSize, resp.PageSize)
			}
		})
	}
}

func TestHandleGetArticlesServiceError(t *testing.T) {
	cfg := &config.Config{}
	svc := newMockService(nil)
	svc.getArticlesMetadata = func(
		_ context.Context, _ string, _ int, _ int, _ *bool,
	) (*service.GetArticlesResult, error) {
		return nil, &serviceError{msg: "database error"}
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("GET", "/v1/articles", http.NoBody)
	w := httptest.NewRecorder()

	h.handleGetArticles(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != "database error" {
		t.Errorf("expected error 'database error', got '%s'", resp.Error)
	}
}

func TestHandleGetArticleSuccess(t *testing.T) {
	cfg := &config.Config{}
	svc := newMockService(nil)
	svc.getArticle = func(_ context.Context, _ string, _ string) (*model.Article, error) {
		return &model.Article{
			ID:      "test-id",
			Title:   "Test Article",
			URL:     "https://example.com/article",
			Content: "<p>Test content</p>",
		}, nil
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("GET", "/v1/articles/test-id", http.NoBody)
	w := httptest.NewRecorder()

	h.handleGetArticle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp model.Article
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "test-id" {
		t.Errorf("expected id 'test-id', got '%s'", resp.ID)
	}
	if resp.Title != "Test Article" {
		t.Errorf("expected title 'Test Article', got '%s'", resp.Title)
	}
	if resp.Content != "<p>Test content</p>" {
		t.Errorf("expected content to be included, got empty or incorrect")
	}
}

func TestHandleGetArticleNotFound(t *testing.T) {
	cfg := &config.Config{}
	svc := newMockService(nil)
	svc.getArticle = func(_ context.Context, _ string, _ string) (*model.Article, error) {
		return nil, &serviceError{msg: testArticleNotFoundError}
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("GET", "/v1/articles/test-id", http.NoBody)
	w := httptest.NewRecorder()

	h.handleGetArticle(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != testArticleNotFoundError {
		t.Errorf("expected error 'article not found', got '%s'", resp.Error)
	}
}

func TestHandleDeleteArticle(t *testing.T) {
	tests := []struct {
		name         string
		deleteResult *service.DeleteArticleResult
		deleteErr    error
		expectedCode int
		expectedErr  string
	}{
		{
			name:         "success",
			deleteResult: &service.DeleteArticleResult{Deleted: 1},
			deleteErr:    nil,
			expectedCode: http.StatusOK,
			expectedErr:  "",
		},
		{
			name:         "not found",
			deleteResult: &service.DeleteArticleResult{Deleted: 0},
			deleteErr:    nil,
			expectedCode: http.StatusOK,
			expectedErr:  "",
		},
		{
			name:         "service error",
			deleteResult: nil,
			deleteErr:    &serviceError{msg: testDatabaseError},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  testDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			svc := newMockService(nil)
			svc.deleteArticle = func(_ context.Context, _, _ string) (*service.DeleteArticleResult, error) {
				return tt.deleteResult, tt.deleteErr
			}
			h := newHandlers(cfg, svc, nil)

			req := httptest.NewRequest("DELETE", "/v1/articles/123", http.NoBody)
			w := httptest.NewRecorder()

			h.handleDeleteArticle(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.deleteErr != nil {
				var resp model.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error != tt.expectedErr {
					t.Errorf("expected error '%s', got '%s'", tt.expectedErr, resp.Error)
				}
				return
			}

			var resp deleteArticleResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.Deleted != tt.deleteResult.Deleted {
				t.Errorf("expected deleted %d, got %d", tt.deleteResult.Deleted, resp.Deleted)
			}
		})
	}
}

func TestHandleDeleteAllArticles(t *testing.T) {
	tests := []struct {
		name         string
		deleteResult *service.DeleteArticleResult
		deleteErr    error
		expectedCode int
		expectedErr  string
	}{
		{
			name:         "success",
			deleteResult: &service.DeleteArticleResult{Deleted: 5},
			deleteErr:    nil,
			expectedCode: http.StatusOK,
			expectedErr:  "",
		},
		{
			name:         "service error",
			deleteResult: nil,
			deleteErr:    &serviceError{msg: testDatabaseError},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  testDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			svc := newMockService(nil)
			svc.deleteAllArticles = func(_ context.Context, _ string) (*service.DeleteArticleResult, error) {
				return tt.deleteResult, tt.deleteErr
			}
			h := newHandlers(cfg, svc, nil)

			req := httptest.NewRequest("DELETE", "/v1/articles", http.NoBody)
			w := httptest.NewRecorder()

			h.handleDeleteAllArticles(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.deleteErr != nil {
				var resp model.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error != tt.expectedErr {
					t.Errorf("expected error '%s', got '%s'", tt.expectedErr, resp.Error)
				}
				return
			}

			var resp deleteArticleResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.Deleted != tt.deleteResult.Deleted {
				t.Errorf("expected deleted %d, got %d", tt.deleteResult.Deleted, resp.Deleted)
			}
		})
	}
}

type logCaptureHandler struct {
	records []*slog.Record
}

func newLogCaptureHandler() *logCaptureHandler {
	return &logCaptureHandler{
		records: make([]*slog.Record, 0),
	}
}

func (h *logCaptureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *logCaptureHandler) Handle(_ context.Context, r slog.Record) error { //nolint:gocritic
	// slog.Handler interface requires value, not pointer
	recordCopy := r
	h.records = append(h.records, &recordCopy)
	return nil
}

func (h *logCaptureHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *logCaptureHandler) WithGroup(_ string) slog.Handler {
	return h
}

func setupLogCapture() {
	slog.SetDefault(slog.New(newLogCaptureHandler()))
}

func TestHandleGetArticlesLogsDBError(t *testing.T) {
	setupLogCapture()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "request completed", 0)
	logRec := &logRecord{Record: &record}

	cfg := &config.Config{}
	svc := newMockService(nil)
	testDatabaseError := "database connection failed"
	svc.getArticlesMetadata = func(
		_ context.Context, _ string, _ int, _ int, _ *bool,
	) (*service.GetArticlesResult, error) {
		return nil, &serviceError{msg: testDatabaseError}
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("GET", "/v1/articles", http.NoBody)
	ctx := context.WithValue(req.Context(), logRecordKey, logRec)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.handleGetArticles(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	foundDBError := false
	logRec.Attrs(func(a slog.Attr) bool {
		if a.Key == dbErrorLogKey {
			foundDBError = true
			if a.Value.String() != testDatabaseError {
				t.Errorf("expected db_error '%s', got '%s'", testDatabaseError, a.Value.String())
			}
		}
		return true
	})

	if !foundDBError {
		t.Error("expected to find db_error attribute in log record")
	}
}

func TestHandleGetArticleLogsDBError(t *testing.T) {
	setupLogCapture()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "request completed", 0)
	logRec := &logRecord{Record: &record}

	cfg := &config.Config{}
	svc := newMockService(nil)
	testDatabaseError := testArticleNotFoundError
	svc.getArticle = func(_ context.Context, _ string, _ string) (*model.Article, error) {
		return nil, &serviceError{msg: testDatabaseError}
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("GET", "/v1/articles/test-id", http.NoBody)
	ctx := context.WithValue(req.Context(), logRecordKey, logRec)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.handleGetArticle(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	foundDBError := false
	logRec.Attrs(func(a slog.Attr) bool {
		if a.Key == dbErrorLogKey {
			foundDBError = true
			if a.Value.String() != testDatabaseError {
				t.Errorf("expected db_error '%s', got '%s'", testDatabaseError, a.Value.String())
			}
		}
		return true
	})

	if !foundDBError {
		t.Error("expected to find db_error attribute in log record")
	}
}

func TestHandleDeleteArticleLogsDBError(t *testing.T) {
	setupLogCapture()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "request completed", 0)
	logRec := &logRecord{Record: &record}

	cfg := &config.Config{}
	svc := newMockService(nil)
	testDatabaseError := "delete failed"
	svc.deleteArticle = func(_ context.Context, _ string, _ string) (*service.DeleteArticleResult, error) {
		return nil, &serviceError{msg: testDatabaseError}
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("DELETE", "/v1/articles/test-id", http.NoBody)
	ctx := context.WithValue(req.Context(), logRecordKey, logRec)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.handleDeleteArticle(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	foundDBError := false
	logRec.Attrs(func(a slog.Attr) bool {
		if a.Key == dbErrorLogKey {
			foundDBError = true
			if a.Value.String() != testDatabaseError {
				t.Errorf("expected db_error '%s', got '%s'", testDatabaseError, a.Value.String())
			}
		}
		return true
	})

	if !foundDBError {
		t.Error("expected to find db_error attribute in log record")
	}
}

func TestHandleDeleteAllArticlesLogsDBError(t *testing.T) {
	setupLogCapture()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "request completed", 0)
	logRec := &logRecord{Record: &record}

	cfg := &config.Config{}
	svc := newMockService(nil)
	testDatabaseError := "batch delete failed"
	svc.deleteAllArticles = func(_ context.Context, _ string) (*service.DeleteArticleResult, error) {
		return nil, &serviceError{msg: testDatabaseError}
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("DELETE", "/v1/articles", http.NoBody)
	ctx := context.WithValue(req.Context(), logRecordKey, logRec)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.handleDeleteAllArticles(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	foundDBError := false
	logRec.Attrs(func(a slog.Attr) bool {
		if a.Key == dbErrorLogKey {
			foundDBError = true
			if a.Value.String() != testDatabaseError {
				t.Errorf("expected db_error '%s', got '%s'", testDatabaseError, a.Value.String())
			}
		}
		return true
	})

	if !foundDBError {
		t.Error("expected to find db_error attribute in log record")
	}
}

func TestHandleDeleteProfile(t *testing.T) {
	tests := []struct {
		name         string
		deleteErr    error
		expectedCode int
		expectedErr  string
	}{
		{
			name:         "success",
			deleteErr:    nil,
			expectedCode: http.StatusOK,
			expectedErr:  "",
		},
		{
			name:         "service error",
			deleteErr:    &serviceError{msg: testDatabaseError},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  testDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			svc := newMockService(nil)
			svc.deleteUserProfile = func(_ context.Context, _ string) error {
				return tt.deleteErr
			}
			h := newHandlers(cfg, svc, nil)

			req := httptest.NewRequest("DELETE", "/v1/user/profile", http.NoBody)
			w := httptest.NewRecorder()

			h.handleDeleteProfile(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.deleteErr != nil {
				var resp model.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error != tt.expectedErr {
					t.Errorf("expected error '%s', got '%s'", tt.expectedErr, resp.Error)
				}
			}
		})
	}
}

func TestHandleDeleteProfileLogsDBError(t *testing.T) {
	setupLogCapture()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "request completed", 0)
	logRec := &logRecord{Record: &record}

	cfg := &config.Config{}
	svc := newMockService(nil)
	testDatabaseError := "failed to delete user profile"
	svc.deleteUserProfile = func(_ context.Context, _ string) error {
		return &serviceError{msg: testDatabaseError}
	}
	h := newHandlers(cfg, svc, nil)

	req := httptest.NewRequest("DELETE", "/v1/user/profile", http.NoBody)
	ctx := context.WithValue(req.Context(), logRecordKey, logRec)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.handleDeleteProfile(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	foundDBError := false
	logRec.Attrs(func(a slog.Attr) bool {
		if a.Key == dbErrorLogKey {
			foundDBError = true
			if a.Value.String() != testDatabaseError {
				t.Errorf("expected db_error '%s', got '%s'", testDatabaseError, a.Value.String())
			}
		}
		return true
	})

	if !foundDBError {
		t.Error("expected to find db_error attribute in log record")
	}
}
