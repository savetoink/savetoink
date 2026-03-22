package handlers

import (
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
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/internal/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/types"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

type sendsMockService struct {
	countSendsFunc func(ctx context.Context, accountID string, startDate, endDate time.Time) (int, error)
}

func (m *sendsMockService) Fetch(
	_ context.Context,
	_ *url.URL,
) (*content.FetchedContent, error) {
	return nil, errors.New("not implemented")
}

func (m *sendsMockService) ParseHTML(
	_ context.Context,
	_ *content.FetchedContent,
) (*html.Node, error) {
	return nil, errors.New("not implemented")
}

func (m *sendsMockService) Clean(
	_ context.Context,
	_ *html.Node,
	_ *url.URL,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *sendsMockService) GenerateEPUB(_ *model.Article) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *sendsMockService) ReadEPUB(_ context.Context, _ string) (io.ReadCloser, string, error) {
	return nil, "", errors.New("not implemented")
}

func (m *sendsMockService) SendArticle(
	_ context.Context,
	_ string,
	_ io.ReadCloser,
	_ string,
) (*email.SendEmailResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *sendsMockService) SendArticleByID(
	_ context.Context,
	_, _ string,
) (*servicetypes.SendArticleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *sendsMockService) CreateArticle(
	_ context.Context,
	_ *url.URL, _ string,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *sendsMockService) UpdateArticle(
	_ context.Context,
	_ *model.Article,
) error {
	return errors.New("not implemented")
}

func (m *sendsMockService) GetArticle(
	_ context.Context,
	_, _ string,
) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *sendsMockService) GetArticlesMetadata(
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

func (m *sendsMockService) DeleteArticle(
	_ context.Context,
	_, _ string,
) (*servicetypes.DeleteArticleResult, error) {
	return &servicetypes.DeleteArticleResult{Deleted: 0}, errors.New("not implemented")
}

func (m *sendsMockService) GetDBError() error {
	return nil
}

func (m *sendsMockService) GetUserDeviceEmailAndAutoSend(
	_ context.Context,
	_ string,
) (_ string, _ bool, _ error) {
	return "", false, nil
}

func (m *sendsMockService) SetUserDeviceEmailWithAutoSend(
	_ context.Context,
	_,
	_ string,
	_ bool,
) error {
	return errors.New("not implemented")
}

func (m *sendsMockService) DeleteUserDeviceEmail(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *sendsMockService) GetUserProfile(_ context.Context, _ string) (*model.UserProfile, error) {
	return &model.UserProfile{
		Email: "user@example.com",
	}, nil
}

func (m *sendsMockService) SetUserEmail(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

func (m *sendsMockService) DeleteUserProfile(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *sendsMockService) ToggleFavorite(_ context.Context, _, _ string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *sendsMockService) CountSendsByAccountDateRange(
	_ context.Context,
	_ string,
	_, _ time.Time,
) (int, error) {
	if m.countSendsFunc != nil {
		return m.countSendsFunc(context.Background(), "", time.Now(), time.Now())
	}
	return 0, nil
}

func (m *sendsMockService) HandleBounce(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

func (m *sendsMockService) IsEmailBouncing(_ context.Context, _, _ string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *sendsMockService) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	return "", errors.New("not implemented")
}

func newSendsTestContext() context.Context {
	ctx := context.Background()
	ctx = auth.AddAccountIDToCtx(ctx, "account-123")
	return ctx
}

func TestHandleGetSends_Success(t *testing.T) {
	mockSvc := &sendsMockService{
		countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
			return 3, nil
		},
	}

	cfg := &config.Config{
		AuthBackend: consts.AuthBackendAuth0,
	}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newSendsTestContext(), "GET", "/v1/sends", nil)
	w := httptest.NewRecorder()

	h.HandleGetSends(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.SendsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 3, resp.CurrentSends)
	assert.Equal(t, 3, resp.TotalSends)
	assert.Equal(t, 10, resp.MaxSendsPerPeriod)
	assert.Equal(t, 10, resp.PeriodDays)
	assert.Equal(t, 7, resp.RemainingSends)
}

func TestHandleGetSends_ZeroSends(t *testing.T) {
	mockSvc := &sendsMockService{
		countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
			return 0, nil
		},
	}

	cfg := &config.Config{
		AuthBackend: consts.AuthBackendAuth0,
	}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newSendsTestContext(), "GET", "/v1/sends", nil)
	w := httptest.NewRecorder()

	h.HandleGetSends(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.SendsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.CurrentSends, "current_sends should be 0 when no sends")
	assert.Equal(t, 0, resp.TotalSends, "total_sends should be 0 when no sends")
	assert.Equal(t, 10, resp.RemainingSends, "remaining_sends should equal max quota when no sends")
}

func TestHandleGetSends_ServiceError(t *testing.T) {
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
			mockSvc := &sendsMockService{
				countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
					return 0, tt.serviceErr
				},
			}

			cfg := &config.Config{
				AuthBackend: consts.AuthBackendAuth0,
			}
			h := New(cfg, mockSvc, http.DefaultClient, nil)

			req := httptest.NewRequestWithContext(newSendsTestContext(), "GET", "/v1/sends", nil)
			w := httptest.NewRecorder()

			h.HandleGetSends(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleGetSends_SharedAPIKey(t *testing.T) {
	mockSvc := &sendsMockService{
		countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
			return 15, nil
		},
	}

	cfg := &config.Config{
		AuthBackend: consts.AuthBackendSharedAPIKey,
	}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newSendsTestContext(), "GET", "/v1/sends", nil)
	w := httptest.NewRecorder()

	h.HandleGetSends(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.SendsResponseNoLimits
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 15, resp.TotalSends)
}

func TestHandleGetSends_SharedAPIKeyError(t *testing.T) {
	mockSvc := &sendsMockService{
		countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
			return 0, errors.New("database error")
		},
	}

	cfg := &config.Config{
		AuthBackend: consts.AuthBackendSharedAPIKey,
	}
	h := New(cfg, mockSvc, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(newSendsTestContext(), "GET", "/v1/sends", nil)
	w := httptest.NewRecorder()

	h.HandleGetSends(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
