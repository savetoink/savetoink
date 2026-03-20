package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/internal/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/types"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

const (
	testUserEmail           = "user@example.com"
	testAccountID           = "account-123"
	testDevicesEndpoint     = "/v1/devices"
	testUserProfileEndpoint = "/v1/user/profile"
	genericErrorMessage     = "generic error"
	someErrorMessage        = "some error"
)

type userprofileMockService struct {
	getUserDeviceEmailFunc    func(ctx context.Context, accountID string) (deviceEmail string, autoSend bool, err error)
	getUserProfileFunc        func(ctx context.Context, accountID string) (*model.UserProfile, error)
	setUserDeviceEmailFunc    func(ctx context.Context, accountID, deviceEmail string, autoSend bool) error
	deleteUserDeviceEmailFunc func(ctx context.Context, accountID string) error
}

func (m *userprofileMockService) Fetch(
	_ context.Context,
	_ *url.URL,
) (*content.FetchedContent, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) ParseHTML(
	_ context.Context,
	_ *content.FetchedContent,
) (*html.Node, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) Clean(
	_ context.Context,
	_ *html.Node,
	_ *url.URL,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) GenerateEPUB(_ *model.Article) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) SendArticle(
	_ context.Context,
	_ string,
	_ io.ReadCloser,
	_ string,
) (*email.SendEmailResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) SendArticleByID(
	_ context.Context,
	_, _ string,
) (*servicetypes.SendArticleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *userprofileMockService) CreateArticle(
	_ context.Context,
	_ *url.URL, _ string,
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

func (m *userprofileMockService) GetDBError() error {
	return nil
}

//nolint:gocritic // mock function with named returns is OK
func (m *userprofileMockService) GetUserDeviceEmailAndAutoSend(
	ctx context.Context,
	accountID string,
) (string, bool, error) {
	if m.getUserDeviceEmailFunc != nil {
		deviceEmail, autoSend, err := m.getUserDeviceEmailFunc(ctx, accountID)
		return deviceEmail, autoSend, err
	}
	return testArticleDeviceEmail, false, nil
}

func (m *userprofileMockService) SetUserDeviceEmailWithAutoSend(
	ctx context.Context,
	accountID,
	deviceEmail string,
	autoSend bool,
) error {
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
		Email: testUserEmail,
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
	ctx = auth.AddAccountIDToCtx(ctx, testAccountID)
	return ctx
}

func TestHandleGetUserProfile_Success(t *testing.T) {
	mockSvc := &userprofileMockService{
		getUserDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
			return testArticleDeviceEmail, true, nil
		},
		getUserProfileFunc: func(_ context.Context, _ string) (*model.UserProfile, error) {
			return &model.UserProfile{
				Email: testUserEmail,
			}, nil
		},
	}

	cfg := &config.Config{}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newUserprofileTestContext(), "GET", testUserProfileEndpoint, nil)
	w := httptest.NewRecorder()

	h.HandleGetUserProfile(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.UserProfileResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testAccountID, resp.Account)
	assert.Equal(t, testUserEmail, resp.Email)
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
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newUserprofileTestContext(), "GET", testUserProfileEndpoint, nil)
	w := httptest.NewRecorder()

	h.HandleGetUserProfile(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.UserProfileResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testAccountID, resp.Account)
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
			name:           genericErrorMessage,
			serviceErr:     errors.New(someErrorMessage),
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
			h := New(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newUserprofileTestContext(), "GET", testUserProfileEndpoint, nil)
			w := httptest.NewRecorder()

			h.HandleGetUserProfile(w, req)

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
			name:           genericErrorMessage,
			serviceErr:     errors.New(someErrorMessage),
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
			h := New(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newUserprofileTestContext(), "GET", testUserProfileEndpoint, nil)
			w := httptest.NewRecorder()

			h.HandleGetUserProfile(w, req)

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
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	body := types.DeviceRequest{
		DeviceEmail: testArticleDeviceEmail,
		AutoSend:    true,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		newUserprofileTestContext(),
		"PUT",
		testDevicesEndpoint,
		bytes.NewReader(bodyBytes),
	)
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleSetDevice(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.DeviceResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testArticleDeviceEmail, resp.DeviceEmail)
	assert.True(t, resp.AutoSend)
}

func TestHandleSetDevice_InvalidJSON(t *testing.T) {
	mockSvc := &userprofileMockService{}

	cfg := &config.Config{}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newUserprofileTestContext(),
		"PUT",
		testDevicesEndpoint,
		bytes.NewReader([]byte("invalid json")))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleSetDevice(w, req)

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
			name:           genericErrorMessage,
			serviceErr:     errors.New(someErrorMessage),
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
			h := New(cfg, mockSvc, http.DefaultClient, nil)

			body := types.DeviceRequest{
				DeviceEmail: testArticleDeviceEmail,
				AutoSend:    true,
			}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequestWithContext(
				newUserprofileTestContext(),
				"PUT",
				testDevicesEndpoint,
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set(contentTypeHeader, contentTypeJSON)
			w := httptest.NewRecorder()

			h.HandleSetDevice(w, req)

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
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newUserprofileTestContext(), "DELETE", testDevicesEndpoint, nil)
	w := httptest.NewRecorder()

	h.HandleDeleteDevice(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.DeviceResponse
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
			name:           genericErrorMessage,
			serviceErr:     errors.New(someErrorMessage),
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
			h := New(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newUserprofileTestContext(), "DELETE", testDevicesEndpoint, nil)
			w := httptest.NewRecorder()

			h.HandleDeleteDevice(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
