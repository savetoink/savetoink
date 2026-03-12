package processor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

const testEmail = "test@kindle.com"

type testCaptureHandler struct {
	record *slog.Record
}

func (h *testCaptureHandler) Handle(_ context.Context, r slog.Record) error { //nolint:gocritic
	if h.record != nil {
		*h.record = r
	}
	return nil
}

func (h *testCaptureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *testCaptureHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *testCaptureHandler) WithGroup(_ string) slog.Handler {
	return h
}

type mockArticleService struct {
	fetchFunc          func(ctx context.Context, u *url.URL) (*content.FetchedContent, error)
	parseHTMLFunc      func(ctx context.Context, fetched *content.FetchedContent) (*html.Node, error)
	cleanFunc          func(ctx context.Context, doc *html.Node, u *url.URL) (*model.Article, error)
	updateFunc         func(ctx context.Context, article *model.Article) error
	getArticleFunc     func(ctx context.Context, accountID, articleID string) (*model.Article, error)
	getUserDeviceEmail func(ctx context.Context, accountID string) (string, bool, error)
	sendArticleFunc    func(
		ctx context.Context,
		destEmail string,
		epubData io.ReadCloser,
		title string,
	) (*email.SendEmailResponse, error)
	generateEPUBFunc func(article *model.Article) (io.ReadCloser, error)
}

func (m *mockArticleService) Fetch(ctx context.Context, u *url.URL) (*content.FetchedContent, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, u)
	}
	return &content.FetchedContent{
		HTML: io.NopCloser(strings.NewReader("<html><body>test</body></html>")),
		URL:  u,
		Type: content.FetcherTypeGo,
	}, nil
}

func (m *mockArticleService) ParseHTML(ctx context.Context, fetched *content.FetchedContent) (*html.Node, error) {
	if m.parseHTMLFunc != nil {
		return m.parseHTMLFunc(ctx, fetched)
	}
	doc, err := html.Parse(strings.NewReader("<html><body>test</body></html>"))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (m *mockArticleService) Clean(ctx context.Context, doc *html.Node, u *url.URL) (*model.Article, error) {
	if m.cleanFunc != nil {
		return m.cleanFunc(ctx, doc, u)
	}
	return &model.Article{
		ID:    "article-123",
		URL:   u.String(),
		Title: "Test Article",
	}, nil
}

func (m *mockArticleService) UpdateArticle(ctx context.Context, article *model.Article) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, article)
	}
	return nil
}

func (m *mockArticleService) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	if m.getArticleFunc != nil {
		return m.getArticleFunc(ctx, accountID, articleID)
	}
	return &model.Article{
		ID:    articleID,
		URL:   "https://example.com/article",
		Title: "Test Article",
	}, nil
}

func (m *mockArticleService) GetUserDeviceEmailAndAutoSend(
	ctx context.Context, accountID string,
) (emailAddr string, autoSend bool, err error) {
	if m.getUserDeviceEmail != nil {
		return m.getUserDeviceEmail(ctx, accountID)
	}
	return testEmail, true, nil
}

func (m *mockArticleService) SendArticle(
	ctx context.Context,
	destEmail string,
	epubData io.ReadCloser,
	title string,
) (*email.SendEmailResponse, error) {
	if m.sendArticleFunc != nil {
		return m.sendArticleFunc(ctx, destEmail, epubData, title)
	}
	return &email.SendEmailResponse{MessageID: "msg-123"}, nil
}

func (m *mockArticleService) GenerateEPUB(article *model.Article) (io.ReadCloser, error) {
	if m.generateEPUBFunc != nil {
		return m.generateEPUBFunc(article)
	}
	return io.NopCloser(strings.NewReader("epub content")), nil
}

func newDefaultMockArticleService(capturedArticle **model.Article) *mockArticleService {
	return &mockArticleService{
		fetchFunc: func(_ context.Context, u *url.URL) (*content.FetchedContent, error) {
			return &content.FetchedContent{
				HTML: io.NopCloser(strings.NewReader("<html><body>test</body></html>")),
				URL:  u,
				Type: content.FetcherTypeGo,
			}, nil
		},
		parseHTMLFunc: func(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
			doc, _ := html.Parse(strings.NewReader("<html><body>test</body></html>"))
			return doc, nil
		},
		cleanFunc: func(_ context.Context, _ *html.Node, u *url.URL) (*model.Article, error) {
			return &model.Article{
				ID:    "article-123",
				URL:   u.String(),
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, article *model.Article) error {
			if capturedArticle != nil {
				*capturedArticle = article
			}
			return nil
		},
	}
}

func TestNewLocalProcessor(t *testing.T) {
	mockSvc := &mockArticleService{}
	processor := NewLocalProcessor(mockSvc)

	assert.NotNil(t, processor)
	assert.Equal(t, mockSvc, processor.service)
}

func TestLocalProcessor_StartProcessing(_ *testing.T) {
	mockSvc := &mockArticleService{}
	processor := NewLocalProcessor(mockSvc)

	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	ctx := context.Background()

	processor.StartProcessing(ctx, event)

	time.Sleep(100 * time.Millisecond)
}

func TestProcessArticle_Success(t *testing.T) {
	var capturedArticle *model.Article

	mockSvc := newDefaultMockArticleService(&capturedArticle)

	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	ctx := context.Background()

	ProcessArticle(ctx, mockSvc, event)

	require.NotNil(t, capturedArticle)
	assert.Equal(t, "article-123", capturedArticle.ID)
	assert.Equal(t, "https://example.com/article", capturedArticle.URL)
	assert.Equal(t, "account-456", capturedArticle.Account)
	assert.NotZero(t, capturedArticle.CreatedAt)
}

func TestProcessArticle_FetchError(t *testing.T) {
	var articleUpdated bool

	mockSvc := &mockArticleService{
		fetchFunc: func(_ context.Context, _ *url.URL) (*content.FetchedContent, error) {
			return nil, errors.New("fetch failed")
		},
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				URL:   "https://example.com/article",
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, article *model.Article) error {
			articleUpdated = true
			assert.Contains(t, article.Error, "fetch failed")
			return nil
		},
	}

	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	ProcessArticle(ctx, mockSvc, event)

	assert.True(t, articleUpdated)
}

func TestProcessArticle_ParseHTMLError(t *testing.T) {
	var articleUpdated bool

	mockSvc := &mockArticleService{
		fetchFunc: func(_ context.Context, u *url.URL) (*content.FetchedContent, error) {
			return &content.FetchedContent{
				HTML: io.NopCloser(strings.NewReader("<html><body>test</body></html>")),
				URL:  u,
				Type: content.FetcherTypeGo,
			}, nil
		},
		parseHTMLFunc: func(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
			return nil, errors.New("parse html failed")
		},
		cleanFunc: func(_ context.Context, _ *html.Node, _ *url.URL) (*model.Article, error) {
			return nil, errors.New("clean failed")
		},
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				URL:   "https://example.com/article",
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, article *model.Article) error {
			articleUpdated = true
			assert.Equal(t, "parse html failed", article.Error)
			return nil
		},
	}

	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	ProcessArticle(ctx, mockSvc, event)

	assert.True(t, articleUpdated)
}

func TestProcessArticle_NilCleanedArticle(t *testing.T) {
	var articleUpdated bool

	mockSvc := &mockArticleService{
		fetchFunc: func(_ context.Context, u *url.URL) (*content.FetchedContent, error) {
			return &content.FetchedContent{
				HTML: io.NopCloser(strings.NewReader("<html><body>test</body></html>")),
				URL:  u,
				Type: content.FetcherTypeGo,
			}, nil
		},
		parseHTMLFunc: func(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
			doc, err := html.Parse(strings.NewReader("<html><body>test</body></html>"))
			if err != nil {
				return nil, err
			}
			return doc, nil
		},
		cleanFunc: func(_ context.Context, _ *html.Node, _ *url.URL) (*model.Article, error) {
			return nil, nil
		},
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				URL:   "https://example.com/article",
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, article *model.Article) error {
			articleUpdated = true
			assert.Contains(t, article.Error, "cleaned article is nil")
			return nil
		},
	}

	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	ProcessArticle(ctx, mockSvc, event)

	assert.True(t, articleUpdated)
}

func TestProcessArticle_UpdateError(t *testing.T) {
	var getArticleCalled bool

	mockSvc := &mockArticleService{
		fetchFunc: func(_ context.Context, u *url.URL) (*content.FetchedContent, error) {
			return &content.FetchedContent{
				HTML: io.NopCloser(strings.NewReader("<html><body>test</body></html>")),
				URL:  u,
				Type: content.FetcherTypeGo,
			}, nil
		},
		parseHTMLFunc: func(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
			doc, err := html.Parse(strings.NewReader("<html><body>test</body></html>"))
			if err != nil {
				return nil, err
			}
			return doc, nil
		},
		cleanFunc: func(_ context.Context, _ *html.Node, u *url.URL) (*model.Article, error) {
			return &model.Article{
				ID:    "article-123",
				URL:   u.String(),
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, _ *model.Article) error {
			return errors.New("update failed")
		},
		getArticleFunc: func(_ context.Context, _, _ string) (*model.Article, error) {
			getArticleCalled = true
			return &model.Article{
				ID:    "article-123",
				URL:   "https://example.com/article",
				Title: "Test Article",
			}, nil
		},
	}

	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	ProcessArticle(ctx, mockSvc, event)

	assert.True(t, getArticleCalled)
}

func TestProcessArticle_URLMismatch(t *testing.T) {
	var capturedArticle *model.Article

	mockSvc := newDefaultMockArticleService(&capturedArticle)
	mockSvc.cleanFunc = func(_ context.Context, _ *html.Node, u *url.URL) (*model.Article, error) {
		return &model.Article{
			ID:    "different-id",
			URL:   u.String(),
			Title: "Test Article",
		}, nil
	}

	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	ProcessArticle(ctx, mockSvc, event)

	require.NotNil(t, capturedArticle)
	assert.Equal(t, "https://example.com/article", capturedArticle.URL)
	assert.Equal(t, "article-123", capturedArticle.ID)
}

func TestProcessArticle_Timeout(t *testing.T) {
	mockSvc := &mockArticleService{
		fetchFunc: func(ctx context.Context, u *url.URL) (*content.FetchedContent, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(5 * time.Second):
				return &content.FetchedContent{
					HTML: io.NopCloser(strings.NewReader("<html><body>test</body></html>")),
					URL:  u,
					Type: content.FetcherTypeGo,
				}, nil
			}
		},
	}

	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	startTime := time.Now()
	ProcessArticle(ctx, mockSvc, event)
	duration := time.Since(startTime)

	assert.Less(t, duration.Milliseconds(), int64(31*1000))
}

func TestMarkArticleError_Success(t *testing.T) {
	var updatedArticle *model.Article

	mockSvc := &mockArticleService{
		getArticleFunc: func(_ context.Context, _, articleID string) (*model.Article, error) {
			return &model.Article{
				ID:    articleID,
				URL:   "https://example.com/article",
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, article *model.Article) error {
			updatedArticle = article
			return nil
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	testErr := errors.New("test error")
	markArticleError(ctx, mockSvc, "account-456", "article-123", "extract", testErr)

	require.NotNil(t, updatedArticle)
	assert.Contains(t, updatedArticle.Error, "test error")
}

func TestMarkArticleError_GetArticleError(t *testing.T) {
	mockSvc := &mockArticleService{
		getArticleFunc: func(_ context.Context, _, _ string) (*model.Article, error) {
			return nil, errors.New("article not found")
		},
		updateFunc: func(_ context.Context, _ *model.Article) error {
			return nil
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	testErr := errors.New("test error")
	markArticleError(ctx, mockSvc, "account-456", "article-123", "extract", testErr)

	err := logging.GetRequestError(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
	assert.Contains(t, err.Error(), "failed to get article")
}

func TestLogArticleResult(t *testing.T) {
	var capturedRecord slog.Record

	captureHandler := &testCaptureHandler{record: &capturedRecord}
	logger := slog.New(captureHandler)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	inheritedAttrs := []map[string]any{
		{"request_id": "req-123"},
		{"account_id": "acc-456"},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	logArticleResult(ctx, inheritedAttrs)

	assert.Equal(t, "article processing completed", capturedRecord.Message)

	var attrs []slog.Attr
	capturedRecord.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, int64(200), attrMap["status"])
}

func TestProcessArticle_SendOnComplete_Success(t *testing.T) {
	var capturedArticle *model.Article
	var sendCalled bool

	mockSvc := &mockArticleService{
		fetchFunc: func(_ context.Context, u *url.URL) (*content.FetchedContent, error) {
			return &content.FetchedContent{
				HTML: io.NopCloser(strings.NewReader("<html><body>test content</body></html>")),
				URL:  u,
				Type: content.FetcherTypeGo,
			}, nil
		},
		parseHTMLFunc: func(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
			doc, err := html.Parse(strings.NewReader("<html><body>test content</body></html>"))
			if err != nil {
				return nil, err
			}
			return doc, nil
		},
		cleanFunc: func(_ context.Context, _ *html.Node, u *url.URL) (*model.Article, error) {
			return &model.Article{
				ID:    "article-123",
				URL:   u.String(),
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, article *model.Article) error {
			capturedArticle = article
			return nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			sendCalled = true
			return &email.SendEmailResponse{MessageID: "msg-123"}, nil
		},
	}

	event := &content.ProcessArticleEvent{
		URL:            "https://example.com/article",
		ArticleID:      "article-123",
		AccountID:      "account-456",
		SendOnComplete: true,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	ProcessArticle(ctx, mockSvc, event)

	require.NotNil(t, capturedArticle)
	assert.True(t, sendCalled)
	assert.Equal(t, "article-123", capturedArticle.ID)
}

func TestProcessArticle_SendOnComplete_DeviceEmailNotSet(t *testing.T) {
	var capturedArticle *model.Article

	mockSvc := &mockArticleService{
		fetchFunc: func(_ context.Context, u *url.URL) (*content.FetchedContent, error) {
			return &content.FetchedContent{
				HTML: io.NopCloser(strings.NewReader("<html><body>test content</body></html>")),
				URL:  u,
				Type: content.FetcherTypeGo,
			}, nil
		},
		parseHTMLFunc: func(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
			doc, err := html.Parse(strings.NewReader("<html><body>test content</body></html>"))
			if err != nil {
				return nil, err
			}
			return doc, nil
		},
		cleanFunc: func(_ context.Context, _ *html.Node, u *url.URL) (*model.Article, error) {
			return &model.Article{
				ID:    "article-123",
				URL:   u.String(),
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, article *model.Article) error {
			capturedArticle = article
			return nil
		},
		getUserDeviceEmail: func(_ context.Context, _ string) (string, bool, error) {
			return "", true, nil
		},
	}

	event := &content.ProcessArticleEvent{
		URL:            "https://example.com/article",
		ArticleID:      "article-123",
		AccountID:      "account-456",
		SendOnComplete: true,
	}

	ctx := context.Background()

	ProcessArticle(ctx, mockSvc, event)

	require.NotNil(t, capturedArticle)
	assert.Equal(t, "article-123", capturedArticle.ID)
	assert.Empty(t, capturedArticle.Error)
}

func TestProcessArticle_SendOnComplete_SendError(t *testing.T) {
	var capturedArticle *model.Article

	mockSvc := &mockArticleService{
		fetchFunc: func(_ context.Context, u *url.URL) (*content.FetchedContent, error) {
			return &content.FetchedContent{
				HTML: io.NopCloser(strings.NewReader("<html><body>test content</body></html>")),
				URL:  u,
				Type: content.FetcherTypeGo,
			}, nil
		},
		parseHTMLFunc: func(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
			doc, err := html.Parse(strings.NewReader("<html><body>test content</body></html>"))
			if err != nil {
				return nil, err
			}
			return doc, nil
		},
		cleanFunc: func(_ context.Context, _ *html.Node, u *url.URL) (*model.Article, error) {
			return &model.Article{
				ID:    "article-123",
				URL:   u.String(),
				Title: "Test Article",
			}, nil
		},
		updateFunc: func(_ context.Context, article *model.Article) error {
			capturedArticle = article
			return nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			return nil, errors.New("email send failed")
		},
	}

	event := &content.ProcessArticleEvent{
		URL:            "https://example.com/article",
		ArticleID:      "article-123",
		AccountID:      "account-456",
		SendOnComplete: true,
	}

	ctx := context.Background()

	ProcessArticle(ctx, mockSvc, event)

	require.NotNil(t, capturedArticle)
	assert.Equal(t, "article-123", capturedArticle.ID)
	assert.Empty(t, capturedArticle.Error)
}

func TestSendArticle_GetDeviceEmailError(t *testing.T) {
	mockSvc := &mockArticleService{
		getUserDeviceEmail: func(_ context.Context, _ string) (string, bool, error) {
			return "", false, errors.New("failed to get device email")
		},
	}

	article := &model.Article{
		ID:      "article-123",
		Account: "account-456",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	err := sendArticle(ctx, mockSvc, article)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get device email")
}

func TestSendArticle_EmailRespWithMessageID(t *testing.T) {
	mockSvc := &mockArticleService{
		getUserDeviceEmail: func(_ context.Context, _ string) (string, bool, error) {
			return "test@kindle.com", true, nil
		},
		sendArticleFunc: func(_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
			return &email.SendEmailResponse{MessageID: "msg-123"}, nil
		},
	}

	article := &model.Article{
		ID:      "article-123",
		Account: "account-456",
		Title:   "Test Article",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &slog.Record{}})
	ctx = context.WithValue(ctx, logging.RequestErrorKey, new(error))

	err := sendArticle(ctx, mockSvc, article)

	require.NoError(t, err)
}
