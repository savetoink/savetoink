package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/email"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/server/auth"
	"github.com/shaftoe/savetoink/internal/service"
)

const (
	errInvalidKindleEmail = "invalid kindle email: must be a valid email ending with @kindle.com or @free.kindle.com"
	testKindleEmail       = "test@kindle.com"
)

func TestHandleSetUserProfile_ValidKindleEmail(t *testing.T) {
	cfg := &config.Config{
		APIKeySecret: "test-key",
	}
	h := newHandlers(cfg, &mockUserProfileService{}, nil)

	body := userProfileRequest{
		KindleEmail: "test@kindle.com",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	middleware := auth.NewAccountIDMiddleware(cfg)
	middleware(http.HandlerFunc(h.handleSetUserProfile)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleSetUserProfile_ValidFreeKindleEmail(t *testing.T) {
	cfg := &config.Config{
		APIKeySecret: "test-key",
	}
	h := newHandlers(cfg, &mockUserProfileService{}, nil)

	body := userProfileRequest{
		KindleEmail: "test@free.kindle.com",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	middleware := auth.NewAccountIDMiddleware(cfg)
	middleware(http.HandlerFunc(h.handleSetUserProfile)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleSetUserProfile_InvalidEmail(t *testing.T) {
	tests := []struct {
		name         string
		kindleEmail  string
		expectedCode int
	}{
		{
			name:         "invalid domain",
			kindleEmail:  "test@gmail.com",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "malformed email",
			kindleEmail:  "@kindle.com",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "email with extra @",
			kindleEmail:  "test@kindle.com@extra",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				APIKeySecret: "test-key",
			}
			h := newHandlers(cfg, &mockUserProfileService{}, nil)

			body := userProfileRequest{
				KindleEmail: tt.kindleEmail,
			}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/v1/profile", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-key")
			w := httptest.NewRecorder()

			middleware := auth.NewAccountIDMiddleware(cfg)
			middleware(http.HandlerFunc(h.handleSetUserProfile)).ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var resp model.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Error != errInvalidKindleEmail {
				t.Errorf("expected error '%s', got: %s", errInvalidKindleEmail, resp.Error)
			}
		})
	}
}

func TestHandleSetUserProfile_MissingEmail(t *testing.T) {
	cfg := &config.Config{
		APIKeySecret: "test-key",
	}
	h := newHandlers(cfg, &mockUserProfileService{}, nil)

	body := userProfileRequest{
		KindleEmail: "",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	middleware := auth.NewAccountIDMiddleware(cfg)
	middleware(http.HandlerFunc(h.handleSetUserProfile)).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleGetUserProfile_Success(t *testing.T) {
	cfg := &config.Config{
		APIKeySecret: "test-key",
	}
	h := newHandlers(cfg, &mockUserProfileService{}, nil)

	req := httptest.NewRequest("GET", "/v1/profile", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	middleware := auth.NewAccountIDMiddleware(cfg)
	middleware(http.HandlerFunc(h.handleGetUserProfile)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp userProfileResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.KindleEmail != testKindleEmail {
		t.Errorf("expected kindle email '%s', got: %s", testKindleEmail, resp.KindleEmail)
	}

	if resp.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got: %s", resp.Email)
	}

	if resp.Account == "" {
		t.Errorf("expected account ID to be set")
	}
}

func TestHandleGetUserProfile_ServiceError(t *testing.T) {
	cfg := &config.Config{
		APIKeySecret: "test-key",
	}
	h := newHandlers(cfg, &mockUserProfileService{shouldError: true}, nil)

	req := httptest.NewRequest("GET", "/v1/profile", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	middleware := auth.NewAccountIDMiddleware(cfg)
	middleware(http.HandlerFunc(h.handleGetUserProfile)).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == "" {
		t.Errorf("expected error message, got empty string")
	}
}

func TestHandleDeleteProfile_Success(t *testing.T) {
	cfg := &config.Config{
		APIKeySecret: "test-key",
	}
	h := newHandlers(cfg, &mockUserProfileService{}, nil)

	req := httptest.NewRequest("DELETE", "/v1/profile", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	middleware := auth.NewAccountIDMiddleware(cfg)
	middleware(http.HandlerFunc(h.handleDeleteProfile)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleDeleteProfile_ServiceError(t *testing.T) {
	cfg := &config.Config{
		APIKeySecret: "test-key",
	}
	h := newHandlers(cfg, &mockUserProfileService{shouldError: true}, nil)

	req := httptest.NewRequest("DELETE", "/v1/profile", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	middleware := auth.NewAccountIDMiddleware(cfg)
	middleware(http.HandlerFunc(h.handleDeleteProfile)).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == "" {
		t.Errorf("expected error message, got empty string")
	}
}

type mockUserProfileService struct {
	shouldError bool
}

func (m *mockUserProfileService) Process(_ context.Context, _ string) (*service.ProcessResult, error) {
	return nil, nil
}

func (m *mockUserProfileService) Send(
	_ context.Context,
	_ *service.ProcessResult,
	_ string,
) (*email.SendEmailResponse, error) {
	return nil, nil
}

func (m *mockUserProfileService) SendArticle(
	_ context.Context,
	_ *model.Article,
	_ string,
) (*email.SendEmailResponse, error) {
	return nil, nil
}

func (m *mockUserProfileService) WriteToFile(_ *service.ProcessResult, _ string) error {
	return nil
}

func (m *mockUserProfileService) CreateArticle(_ context.Context, _, _ string) (*service.CreateArticleResult, error) {
	return nil, nil
}

func (m *mockUserProfileService) GetArticle(_ context.Context, _, _ string) (*model.Article, error) {
	return nil, nil
}

func (m *mockUserProfileService) GetArticlesMetadata(
	_ context.Context,
	_ string,
	_, _ int,
	_ *bool,
) (*service.GetArticlesResult, error) {
	return nil, nil
}

func (m *mockUserProfileService) DeleteArticle(_ context.Context, _, _ string) (*service.DeleteArticleResult, error) {
	return nil, nil
}

func (m *mockUserProfileService) DeleteAllArticles(_ context.Context, _ string) (*service.DeleteArticleResult, error) {
	return nil, nil
}

func (m *mockUserProfileService) GetDBError() error {
	return nil
}

func (m *mockUserProfileService) GetUserKindleEmail(_ context.Context, _ string) (string, error) {
	if m.shouldError {
		return "", errors.New("database error")
	}
	return testKindleEmail, nil
}

func (m *mockUserProfileService) SetUserKindleEmail(_ context.Context, _, _ string) error {
	if m.shouldError {
		return errors.New("database error")
	}
	return nil
}

func (m *mockUserProfileService) GetUserProfile(_ context.Context, _ string) (*model.UserProfile, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	return &model.UserProfile{
		Email:       "test@example.com",
		KindleEmail: testKindleEmail,
	}, nil
}

func (m *mockUserProfileService) SetUserEmail(_ context.Context, _, _ string) error {
	if m.shouldError {
		return errors.New("database error")
	}
	return nil
}

func (m *mockUserProfileService) DeleteUserProfile(_ context.Context, _ string) error {
	if m.shouldError {
		return errors.New("database error")
	}
	return nil
}

func (m *mockUserProfileService) ToggleFavorite(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}
