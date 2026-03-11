package lambda

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/html"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Fetch(ctx context.Context, u *url.URL) (*content.FetchedContent, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*content.FetchedContent), args.Error(1)
}

func (m *mockService) ParseHTML(ctx context.Context, fetched *content.FetchedContent) (*html.Node, error) {
	args := m.Called(ctx, fetched)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*html.Node), args.Error(1)
}

func (m *mockService) Clean(ctx context.Context, doc *html.Node, u *url.URL) (*model.Article, error) {
	args := m.Called(ctx, doc, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Article), args.Error(1)
}

func (m *mockService) UpdateArticle(ctx context.Context, article *model.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m *mockService) GetArticle(ctx context.Context, accountID, articleID string) (*model.Article, error) {
	args := m.Called(ctx, accountID, articleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Article), args.Error(1)
}

func (m *mockService) GetUserDeviceEmail( //nolint:gocritic
	ctx context.Context, accountID string) (string, bool, error) {
	args := m.Called(ctx, accountID)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *mockService) SendArticleByID(ctx context.Context, accountID, articleID string) (
	*servicetypes.SendArticleResult, error) {
	args := m.Called(ctx, accountID, articleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*servicetypes.SendArticleResult), args.Error(1)
}

func TestHandleEvent_NilEvent(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})

	err := HandleEvent(ctx, nil, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_MissingRequestID(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_MissingURL(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID: "req-123",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_MissingArticleID(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		AccountID: "account-456",
	}

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_MissingAccountID(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		ArticleID: "article-123",
	}

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_FetchError(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	eventURL, _ := url.Parse(event.URL)
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{}, errors.New("fetch failed"))
	mockSvc.On(
		"GetArticle", mock.Anything, event.AccountID, event.ArticleID).Return(&model.Article{ID: event.ArticleID}, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "Fetch", mock.Anything, eventURL)
	mockSvc.AssertCalled(t, "UpdateArticle", mock.Anything, mock.Anything)
}

func TestHandleEvent_ParseHTMLError(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).
		Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).
		Return(nil, errors.New("clean failed"))
	mockSvc.On(
		"GetArticle", mock.Anything, event.AccountID, event.ArticleID).Return(&model.Article{ID: event.ArticleID}, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "Fetch", mock.Anything, eventURL)
}

func TestHandleEvent_UpdateError(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: "Test Article"}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(errors.New("update failed"))
	mockSvc.On(
		"GetArticle", mock.Anything, event.AccountID, event.ArticleID).Return(&model.Article{ID: event.ArticleID}, nil)

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "Fetch", mock.Anything, eventURL)
	mockSvc.AssertCalled(t, "ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent"))
	mockSvc.AssertCalled(t, "Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything)
	mockSvc.AssertCalled(t, "UpdateArticle", mock.Anything, mock.Anything)
}

func TestHandleEvent_Success(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: "Test Article"}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "Fetch", mock.Anything, eventURL)
	mockSvc.AssertCalled(t, "ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent"))
	mockSvc.AssertCalled(t, "Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything)
	mockSvc.AssertCalled(t, "UpdateArticle", mock.Anything, mock.Anything)
}

func TestHandleEvent_SendOnComplete_Success(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID:      "req-123",
		URL:            "https://example.com/article",
		ArticleID:      "article-123",
		AccountID:      "account-456",
		SendOnComplete: true,
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: "Test Article"}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)
	mockSvc.On("GetUserDeviceEmail", mock.Anything, event.AccountID).Return("device@example.com", true, nil)
	mockSvc.On("SendArticleByID", mock.Anything, event.AccountID, event.ArticleID).Return(&servicetypes.SendArticleResult{
		DeviceEmail: "device@example.com",
	}, nil)

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "GetUserDeviceEmail", mock.Anything, event.AccountID)
	mockSvc.AssertCalled(t, "SendArticleByID", mock.Anything, event.AccountID, event.ArticleID)
}

func TestHandleEvent_SendOnComplete_NoDeviceEmail(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID:      "req-123",
		URL:            "https://example.com/article",
		ArticleID:      "article-123",
		AccountID:      "account-456",
		SendOnComplete: true,
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: "Test Article"}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)
	mockSvc.On("GetUserDeviceEmail", mock.Anything, event.AccountID).Return("", false, nil)

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "GetUserDeviceEmail", mock.Anything, event.AccountID)
	mockSvc.AssertNotCalled(t, "SendArticleByID")
}

func TestHandleEvent_SendOnComplete_SendError(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID:      "req-123",
		URL:            "https://example.com/article",
		ArticleID:      "article-123",
		AccountID:      "account-456",
		SendOnComplete: true,
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: "Test Article"}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)
	mockSvc.On("GetUserDeviceEmail", mock.Anything, event.AccountID).Return("device@example.com", true, nil)
	mockSvc.On("SendArticleByID", mock.Anything, event.AccountID, event.ArticleID).Return(nil, errors.New("send failed"))

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "SendArticleByID", mock.Anything, event.AccountID, event.ArticleID)
}

func TestHandleEvent_InheritedAttrs(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "lambda-req-123",
	})
	event := &content.ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		ArticleID: "article-123",
		AccountID: "account-456",
		InheritedAttrs: []map[string]any{
			{"user_id": "user-123"},
			{"source": "api"},
		},
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: "Test Article"}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.MatchedBy(func(_ *model.Article) bool {
		return true
	})).Return(nil)

	err := HandleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	assert.Contains(t, event.InheritedAttrs, map[string]any{"orig_request_id": "req-123"})
	assert.Contains(t, event.InheritedAttrs, map[string]any{"request_id": "lambda-req-123"})
	assert.NotContains(t, event.InheritedAttrs, map[string]any{"source": "api"})
}

var _ processor.Service = (*mockService)(nil)
