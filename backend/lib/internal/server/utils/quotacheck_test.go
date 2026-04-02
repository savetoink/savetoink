// Package utils provides utility functions for HTTP handlers.
package utils

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	"github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

const testDeviceEmail = "test@kindle.com"

type quotaCheckMockService struct {
	countSendsFunc      func(ctx context.Context, accountID string, startDate, endDate time.Time) (int, error)
	getDeviceEmailFunc  func(ctx context.Context, accountID string) (string, bool, error)
	isEmailBouncingFunc func(ctx context.Context, accountID, email string) (bool, error)
	getUserProfileFunc  func(ctx context.Context, accountID string) (*model.UserProfile, error)
}

func (m *quotaCheckMockService) Fetch(_ context.Context, _ *url.URL) (*content.FetchedContent, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) ParseHTML(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) Clean(_ context.Context, _ *html.Node, _ *url.URL) (*model.Article, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) GenerateEPUB(_ *model.Article) (io.ReadCloser, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) ReadEPUB(_ context.Context, _ *url.URL) (io.ReadCloser, string, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) SendArticle(
	_ context.Context,
	_ string,
	_ io.ReadCloser,
	_ string,
) (*email.SendEmailResponse, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) SendArticleByID(
	_ context.Context,
	_, _ string,
) (*servicetypes.SendArticleResult, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) CreateArticle(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) UpdateArticle(_ context.Context, _ *model.Article) error {
	return nil
}

func (m *quotaCheckMockService) GetArticle(_ context.Context, _, _ string) (*model.Article, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) GetArticlesMetadata(
	_ context.Context,
	_ string,
	_, _ int,
	_ *types.ArticleFilter,
) (*servicetypes.GetArticlesResult, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) DeleteArticle(
	_ context.Context,
	_, _ string,
) (*servicetypes.DeleteArticleResult, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) GetDBError() error {
	panic("not implemented")
}

func (m *quotaCheckMockService) GetUserDeviceEmailAndAutoSend(
	ctx context.Context,
	accountID string,
) (emailAddr string, ok bool, err error) {
	if m.getDeviceEmailFunc != nil {
		return m.getDeviceEmailFunc(ctx, accountID)
	}
	return "", false, nil
}

func (m *quotaCheckMockService) SetUserDeviceEmailWithAutoSend(_ context.Context, _, _ string, _ bool) error {
	panic("not implemented")
}

func (m *quotaCheckMockService) DeleteUserDeviceEmail(_ context.Context, _ string) error {
	panic("not implemented")
}

func (m *quotaCheckMockService) GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, accountID)
	}
	return &model.UserProfile{}, nil
}

func (m *quotaCheckMockService) SetUserEmail(_ context.Context, _, _ string) error {
	panic("not implemented")
}

func (m *quotaCheckMockService) DeleteUserProfile(_ context.Context, _ string) error {
	panic("not implemented")
}

func (m *quotaCheckMockService) ToggleFavorite(_ context.Context, _, _ string) (bool, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) AddArticleTags(_ context.Context, _, _ string, _ []string) error {
	panic("not implemented")
}

func (m *quotaCheckMockService) RemoveArticleTags(_ context.Context, _, _ string, _ []string) error {
	panic("not implemented")
}

func (m *quotaCheckMockService) SetArticleTags(_ context.Context, _, _ string, _ []string) error {
	panic("not implemented")
}

func (m *quotaCheckMockService) GetArticleTags(_ context.Context, _, _ string) ([]string, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) GetAllTagsForAccount(_ context.Context, _ string) ([]string, error) {
	panic("not implemented")
}

func (m *quotaCheckMockService) CountSendsByAccountDateRange(
	ctx context.Context,
	accountID string,
	startDate, endDate time.Time,
) (int, error) {
	if m.countSendsFunc != nil {
		return m.countSendsFunc(ctx, accountID, startDate, endDate)
	}
	return 0, nil
}

func (m *quotaCheckMockService) HandleBounce(_ context.Context, _, _ string) error {
	panic("not implemented")
}

func (m *quotaCheckMockService) IsEmailBouncing(
	ctx context.Context,
	accountID, emailAddr string,
) (bool, error) {
	if m.isEmailBouncingFunc != nil {
		return m.isEmailBouncingFunc(ctx, accountID, emailAddr)
	}
	return false, nil
}

func (m *quotaCheckMockService) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	panic("not implemented")
}

func TestCheckEmailBackendEnabled(t *testing.T) {
	tests := []struct {
		name           string
		emailProvider  consts.EmailProvider
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Mailjet backend configured",
			emailProvider:  consts.EmailBackendMailjet,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "empty email provider",
			emailProvider:  "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email backend not configured: invalid input",
		},
		{
			name:           "different provider configured",
			emailProvider:  "other-provider",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email backend not configured: invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)

			err := CheckEmailBackendEnabled(w, r, tt.emailProvider)

			if tt.expectedError != "" {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Equal(t, tt.expectedStatus, w.Code)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCheckQuotaAndDeviceEmail(t *testing.T) {
	t.Run("shared API key backend skips quota check", func(t *testing.T) {
		mockSvc := &quotaCheckMockService{}
		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		ctx := context.Background()

		count, err := CheckQuotaAndDeviceEmail(ctx, w, r, mockSvc, consts.AuthBackendSharedAPIKey, "test-account", false)

		assert.Nil(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Auth0 with quota check disabled skips quota check", func(t *testing.T) {
		mockSvc := &quotaCheckMockService{}
		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		ctx := context.Background()

		count, err := CheckQuotaAndDeviceEmail(ctx, w, r, mockSvc, consts.AuthBackendAuth0, "test-account", true)

		assert.Nil(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Auth0 with count under limit", func(t *testing.T) {
		mockSvc := &quotaCheckMockService{
			countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
				return 5, nil
			},
			getDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
				return testDeviceEmail, true, nil
			},
			isEmailBouncingFunc: func(_ context.Context, _, _ string) (bool, error) {
				return false, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		ctx := context.Background()

		count, err := CheckQuotaAndDeviceEmail(ctx, w, r, mockSvc, consts.AuthBackendAuth0, "test-account", false)

		assert.Nil(t, err)
		assert.Equal(t, 5, count)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Auth0 with count at limit", func(t *testing.T) {
		mockSvc := &quotaCheckMockService{
			countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
				return consts.MaxFreeTierSendsPerPeriod, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		ctx := context.Background()

		count, err := CheckQuotaAndDeviceEmail(ctx, w, r, mockSvc, consts.AuthBackendAuth0, "test-account", false)

		assert.NotNil(t, err)
		assert.Equal(t, "free tier limit exceeded: quota exceeded", err.Error())
		assert.Equal(t, consts.MaxFreeTierSendsPerPeriod, count)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("Auth0 with count over limit", func(t *testing.T) {
		mockSvc := &quotaCheckMockService{
			countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
				return consts.MaxFreeTierSendsPerPeriod + 1, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		ctx := context.Background()

		count, err := CheckQuotaAndDeviceEmail(ctx, w, r, mockSvc, consts.AuthBackendAuth0, "test-account", false)

		assert.NotNil(t, err)
		assert.Equal(t, "free tier limit exceeded: quota exceeded", err.Error())
		assert.Equal(t, consts.MaxFreeTierSendsPerPeriod+1, count)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("Auth0 with service error counting sends", func(t *testing.T) {
		mockSvc := &quotaCheckMockService{
			countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
				return 0, errors.New("database error")
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		ctx := context.Background()

		count, err := CheckQuotaAndDeviceEmail(ctx, w, r, mockSvc, consts.AuthBackendAuth0, "test-account", false)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to check subscription limit")
		assert.Equal(t, 0, count)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("device email bouncing with bounce info", func(t *testing.T) {
		mockSvc := &quotaCheckMockService{
			countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
				return 5, nil
			},
			getDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
				return testDeviceEmail, true, nil
			},
			isEmailBouncingFunc: func(_ context.Context, _, _ string) (bool, error) {
				return true, nil
			},
			getUserProfileFunc: func(_ context.Context, _ string) (*model.UserProfile, error) {
				return &model.UserProfile{
					BouncedEmails: map[string]model.BounceInfo{
						testDeviceEmail: {Error: "bounced due to invalid mailbox"},
					},
				}, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		ctx := context.Background()

		count, err := CheckQuotaAndDeviceEmail(ctx, w, r, mockSvc, consts.AuthBackendAuth0, "test-account", false)

		assert.NotNil(t, err)
		assert.Equal(
			t,
			"device email "+testDeviceEmail+" is blocked due to previous bounce: bounced due to invalid mailbox: invalid input",
			err.Error(),
		)
		assert.Equal(t, 5, count)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("device email bouncing without bounce info", func(t *testing.T) {
		mockSvc := &quotaCheckMockService{
			countSendsFunc: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
				return 5, nil
			},
			getDeviceEmailFunc: func(_ context.Context, _ string) (string, bool, error) {
				return testDeviceEmail, true, nil
			},
			isEmailBouncingFunc: func(_ context.Context, _, _ string) (bool, error) {
				return true, nil
			},
			getUserProfileFunc: func(_ context.Context, _ string) (*model.UserProfile, error) {
				return &model.UserProfile{BouncedEmails: map[string]model.BounceInfo{}}, nil
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		ctx := context.Background()

		count, err := CheckQuotaAndDeviceEmail(ctx, w, r, mockSvc, consts.AuthBackendAuth0, "test-account", false)

		assert.NotNil(t, err)
		assert.Equal(t, "device email "+testDeviceEmail+" is blocked due to previous bounce: invalid input", err.Error())
		assert.Equal(t, 5, count)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
