package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/auth"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userprofileMockService struct {
	getUserDeviceEmailFunc    func(ctx context.Context, accountID string) (deviceEmail string, autoSend bool, err error)
	getUserProfileFunc        func(ctx context.Context, accountID string) (*model.UserProfile, error)
	setUserDeviceEmailFunc    func(ctx context.Context, accountID, deviceEmail string, autoSend bool) error
	deleteUserDeviceEmailFunc func(ctx context.Context, accountID string) error
}

func (m *userprofileMockService) Fetch(
	_ context.Context,
	_ string,
) ([]byte, content.FetcherType, error) {
	return nil, content.FetcherTypeGo, errors.New("not implemented")
}

func (m *userprofileMockService) Extract(
	_ context.Context,
	_ []byte,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) GenerateEPUB(_ *model.Article) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) SendArticle(
	_ context.Context,
	_ string,
	_ []byte,
	_ string,
) (*email.SendEmailResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) CreateArticle(
	_ context.Context,
	_, _ string,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) UpdateArticle(
	_ context.Context,
	_ *model.Article,
) error {
	return errors.New("not implemented")
}

func (m *userprofileMockService) GetArticle(
	_ context.Context,
	_, _ string,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) GetArticlesMetadata(
	_ context.Context,
	_ string,
	_, _ int,
	_ *bool,
) (*servicetypes.GetArticlesResult, error) {
	return &servicetypes.GetArticlesResult{
		Articles: []*model.Article{},
		Page:     0,
		PageSize: 0,
		Total:    0,
		HasMore:  false,
	}, nil
}

func (m *userprofileMockService) DeleteArticle(
	_ context.Context,
	_, _ string,
) (*servicetypes.DeleteArticleResult, error) {
	return &servicetypes.DeleteArticleResult{Deleted: 0}, errors.New("not implemented")
}

func (m *userprofileMockService) DeleteAllArticles(
	_ context.Context,
	_ string,
) (*servicetypes.DeleteArticleResult, error) {
	return &servicetypes.DeleteArticleResult{Deleted: 0}, errors.New("not implemented")
}

func (m *userprofileMockService) GetDBError() error {
	return nil
}

//nolint:gocritic,unnamedresult // mock function with named returns is OK
func (m *userprofileMockService) GetUserDeviceEmail(
	ctx context.Context,
	accountID string,
) (string, bool, error) {
	if m.getUserDeviceEmailFunc != nil {
		deviceEmail, autoSend, err := m.getUserDeviceEmailFunc(ctx, accountID)
		return deviceEmail, autoSend, err
	}
	return testArticleDeviceEmail, false, nil
}

func (m *userprofileMockService) SetUserDeviceEmailWithAutoSend(ctx context.Context, accountID, deviceEmail string, autoSend bool) error { //nolint:lll // long function signature
	if m.setUserDeviceEmailFunc != nil {
		return m.setUserDeviceEmailFunc(ctx, accountID, deviceEmail, autoSend)
	}
	return nil
}

func (m *userprofileMockService) DeleteUserDeviceEmail(ctx context.Context, accountID string) error {
	if m.deleteUserDeviceEmailFunc != nil {
		return m.deleteUserDeviceEmailFunc(ctx, accountID)
	}
	return nil
}

func (m *userprofileMockService) GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, accountID)
	}
	return &model.UserProfile{
		Email: "user@example.com",
	}, nil
}

func (m *userprofileMockService) SetUserEmail(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

func (m *userprofileMockService) DeleteUserProfile(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *userprofileMockService) ToggleFavorite(_ context.Context, _, _ string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *userprofileMockService) CountSendsByAccountDateRange(
	_ context.Context,
	_ string,
	_, _ time.Time,
) (int, error) {
	return 0, nil
}

func (m *userprofileMockService) HandleBounce(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

func (m *userprofileMockService) IsEmailBouncing(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *userprofileMockService) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	return "", errors.New("not implemented")
}

func newUserprofileTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.AccountIDKey, "account-123")
	return ctx
}

func TestHandleGetUserProfile_Success(t *testing.T) {
	mockSvc := &userprofileMockService{
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, true, nil
		},
		getUserProfileFunc: func(_ context.Context, _ string) (*model.UserProfile, error) {
			return &model.UserProfile{
				Email: "user@example.com",
			}, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newUserprofileTestContext(), "GET", "/v1/user/profile", nil)
	w := httptest.NewRecorder()

	h.handleGetUserProfile(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp userProfileResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "account-123", resp.Account)
	assert.Equal(t, "user@example.com", resp.Email)
	assert.Equal(t, testArticleDeviceEmail, resp.DeviceEmail)
	assert.True(t, resp.AutoSend)
}

func TestHandleGetUserProfile_NilProfile(t *testing.T) {
	mockSvc := &userprofileMockService{
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, false, nil
		},
		getUserProfileFunc: func(_ context.Context, _ string) (*model.UserProfile, error) {
			return nil, nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newUserprofileTestContext(), "GET", "/v1/user/profile", nil)
	w := httptest.NewRecorder()

	h.handleGetUserProfile(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp userProfileResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "account-123", resp.Account)
	assert.Equal(t, "", resp.Email)
	assert.Equal(t, testArticleDeviceEmail, resp.DeviceEmail)
	assert.False(t, resp.AutoSend)
}

func TestHandleGetUserProfile_GetUserDeviceEmailError(t *testing.T) {
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
			mockSvc := &userprofileMockService{
				getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
					return "", false, tt.serviceErr
				},
			}

			cfg := &config.Config{}
			h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newUserprofileTestContext(), "GET", "/v1/user/profile", nil)
			w := httptest.NewRecorder()

			h.handleGetUserProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleGetUserProfile_GetUserProfileError(t *testing.T) {
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
			mockSvc := &userprofileMockService{
				getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
					return testArticleDeviceEmail, false, nil
				},
				getUserProfileFunc: func(_ context.Context, _ string) (*model.UserProfile, error) {
					return nil, tt.serviceErr
				},
			}

			cfg := &config.Config{}
			h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newUserprofileTestContext(), "GET", "/v1/user/profile", nil)
			w := httptest.NewRecorder()

			h.handleGetUserProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleSetDevice_Success(t *testing.T) {
	mockSvc := &userprofileMockService{
		setUserDeviceEmailFunc: func(_ context.Context, _, _ string, _ bool) error {
			return nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	body := deviceRequest{
		DeviceEmail: testArticleDeviceEmail,
		AutoSend:    true,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(newUserprofileTestContext(), "PUT", "/v1/devices", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleSetDevice(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp deviceResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testArticleDeviceEmail, resp.DeviceEmail)
	assert.True(t, resp.AutoSend)
}

func TestHandleSetDevice_InvalidJSON(t *testing.T) {
	mockSvc := &userprofileMockService{}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newUserprofileTestContext(), "PUT", "/v1/devices", bytes.NewReader([]byte("invalid json"))) //nolint:lll // long function signature
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleSetDevice(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleSetDevice_ServiceError(t *testing.T) {
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
			mockSvc := &userprofileMockService{
				setUserDeviceEmailFunc: func(_ context.Context, _, _ string, _ bool) error {
					return tt.serviceErr
				},
			}

			cfg := &config.Config{}
			h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

			body := deviceRequest{
				DeviceEmail: testArticleDeviceEmail,
				AutoSend:    true,
			}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequestWithContext(newUserprofileTestContext(), "PUT", "/v1/devices", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.handleSetDevice(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleDeleteDevice_Success(t *testing.T) {
	mockSvc := &userprofileMockService{
		deleteUserDeviceEmailFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}

	cfg := &config.Config{}
	h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newUserprofileTestContext(), "DELETE", "/v1/devices", nil)
	w := httptest.NewRecorder()

	h.handleDeleteDevice(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp deviceResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "", resp.DeviceEmail)
	assert.False(t, resp.AutoSend)
}

func TestHandleDeleteDevice_ServiceError(t *testing.T) {
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
			mockSvc := &userprofileMockService{
				deleteUserDeviceEmailFunc: func(_ context.Context, _ string) error {
					return tt.serviceErr
				},
			}

			cfg := &config.Config{}
			h := newHandlers(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newUserprofileTestContext(), "DELETE", "/v1/devices", nil)
			w := httptest.NewRecorder()

			h.handleDeleteDevice(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
