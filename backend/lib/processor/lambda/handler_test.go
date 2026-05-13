package lambda

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/html"
)

const (
	testHandlerLambdaRequestID = "lambda-req-123"
	testHandlerRequestID       = "req-123"
	testHandlerArticleID       = "article-123"
	testHandlerAccountID       = "account-456"
	testHandlerURL             = "https://example.com/article"
	testHandlerTestArticle     = "Test Article"
	testHandlerSourceAPI       = "api"
	testHandlerSourceKey       = "source"
	testHandlerUserID          = "user-123"
	testUserIDKey              = "user_id"
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

func (m *mockService) GetUserDeviceEmailAndAutoSend( //nolint:gocritic
	ctx context.Context, accountID string) (string, bool, error) {
	args := m.Called(ctx, accountID)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *mockService) SendArticle(
	ctx context.Context,
	destEmail string,
	epubData io.ReadCloser,
	title string,
) (*email.SendEmailResponse, error) {
	args := m.Called(ctx, destEmail, epubData, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*email.SendEmailResponse), args.Error(1)
}

func (m *mockService) GenerateEPUB(article *model.Article) (io.ReadCloser, error) {
	args := m.Called(article)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func TestHandleEvent_NilEvent(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})

	err := handleEvent(ctx, nil, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_MissingRequestID(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		URL:       testHandlerURL,
		ArticleID: testHandlerArticleID,
		AccountID: testHandlerAccountID,
	}

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_MissingURL(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID: testHandlerRequestID,
		ArticleID: testHandlerArticleID,
		AccountID: testHandlerAccountID,
	}

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_MissingArticleID(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID: testHandlerRequestID,
		URL:       testHandlerURL,
		AccountID: testHandlerAccountID,
	}

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_MissingAccountID(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID: testHandlerRequestID,
		URL:       testHandlerURL,
		ArticleID: testHandlerArticleID,
	}

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertNotCalled(t, "Fetch")
}

func TestHandleEvent_FetchError(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID: testHandlerRequestID,
		URL:       testHandlerURL,
		ArticleID: testHandlerArticleID,
		AccountID: testHandlerAccountID,
	}

	eventURL, _ := url.Parse(event.URL)
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{}, errors.New("fetch failed"))
	mockSvc.On(
		"GetArticle", mock.Anything, event.AccountID, event.ArticleID).Return(&model.Article{ID: event.ArticleID}, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "Fetch", mock.Anything, eventURL)
	mockSvc.AssertCalled(t, "UpdateArticle", mock.Anything, mock.Anything)
}

func TestHandleEvent_ParseHTMLError(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID: testHandlerRequestID,
		URL:       testHandlerURL,
		ArticleID: testHandlerArticleID,
		AccountID: testHandlerAccountID,
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

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "Fetch", mock.Anything, eventURL)
}

func TestHandleEvent_UpdateError(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID: testHandlerRequestID,
		URL:       testHandlerURL,
		ArticleID: testHandlerArticleID,
		AccountID: testHandlerAccountID,
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: testHandlerTestArticle}
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

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "Fetch", mock.Anything, eventURL)
	mockSvc.AssertCalled(t, "ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent"))
	mockSvc.AssertCalled(t, "Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything)
	mockSvc.AssertCalled(t, "UpdateArticle", mock.Anything, mock.Anything)
}

func TestHandleEvent_Success(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID: testHandlerRequestID,
		URL:       testHandlerURL,
		ArticleID: testHandlerArticleID,
		AccountID: testHandlerAccountID,
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: testHandlerTestArticle}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "Fetch", mock.Anything, eventURL)
	mockSvc.AssertCalled(t, "ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent"))
	mockSvc.AssertCalled(t, "Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything)
	mockSvc.AssertCalled(t, "UpdateArticle", mock.Anything, mock.Anything)
}

func TestHandleEvent_SendOnComplete_Success(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID:      testHandlerRequestID,
		URL:            testHandlerURL,
		ArticleID:      testHandlerArticleID,
		AccountID:      testHandlerAccountID,
		SendOnComplete: true,
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: testHandlerTestArticle}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)
	mockSvc.On("GetUserDeviceEmailAndAutoSend", mock.Anything, event.AccountID).Return(
		"device@example.com", true, nil)
	mockSvc.On("GenerateEPUB", mock.AnythingOfType("*model.Article")).Return(
		io.NopCloser(bytes.NewReader(htmlBytes)), nil)
	mockSvc.On("SendArticle", mock.Anything, "device@example.com",
		mock.Anything, testHandlerTestArticle).Return(
		&email.SendEmailResponse{MessageID: "msg-123"}, nil)

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "GetUserDeviceEmailAndAutoSend", mock.Anything, event.AccountID)
	mockSvc.AssertCalled(t, "GenerateEPUB", mock.AnythingOfType("*model.Article"))
	mockSvc.AssertCalled(t, "SendArticle", mock.Anything, "device@example.com",
		mock.Anything, testHandlerTestArticle)
}

func TestHandleEvent_SendOnComplete_NoDeviceEmail(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID:      testHandlerRequestID,
		URL:            testHandlerURL,
		ArticleID:      testHandlerArticleID,
		AccountID:      testHandlerAccountID,
		SendOnComplete: true,
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: testHandlerTestArticle}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)
	mockSvc.On("GetUserDeviceEmailAndAutoSend", mock.Anything, event.AccountID).Return("", false, nil)

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "GetUserDeviceEmailAndAutoSend", mock.Anything, event.AccountID)
	mockSvc.AssertNotCalled(t, "SendArticle")
}

func TestHandleEvent_SendOnComplete_SendError(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID:      testHandlerRequestID,
		URL:            testHandlerURL,
		ArticleID:      testHandlerArticleID,
		AccountID:      testHandlerAccountID,
		SendOnComplete: true,
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: testHandlerTestArticle}
	doc, _ := html.Parse(bytes.NewReader(htmlBytes))
	mockSvc.On("Fetch", mock.Anything, eventURL).Return(&content.FetchedContent{
		HTML: io.NopCloser(bytes.NewReader(htmlBytes)),
		URL:  eventURL,
		Type: content.FetcherTypeBrowserless,
	}, nil)
	mockSvc.On("ParseHTML", mock.Anything, mock.AnythingOfType("*content.FetchedContent")).Return(doc, nil)
	mockSvc.On("Clean", mock.Anything, mock.AnythingOfType("*html.Node"), mock.Anything).Return(article, nil)
	mockSvc.On("UpdateArticle", mock.Anything, mock.Anything).Return(nil)
	mockSvc.On("GetUserDeviceEmailAndAutoSend", mock.Anything, event.AccountID).Return(
		"device@example.com", true, nil)
	mockSvc.On("GenerateEPUB", mock.AnythingOfType("*model.Article")).Return(
		io.NopCloser(bytes.NewReader(htmlBytes)), nil)
	mockSvc.On("SendArticle", mock.Anything, "device@example.com",
		mock.Anything, testHandlerTestArticle).Return(nil,
		errors.New("send failed"))

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	mockSvc.AssertCalled(t, "GenerateEPUB", mock.AnythingOfType("*model.Article"))
	mockSvc.AssertCalled(t, "SendArticle", mock.Anything, "device@example.com",
		mock.Anything, testHandlerTestArticle)
}

func TestHandleEvent_InheritedAttrs(t *testing.T) {
	mockSvc := new(mockService)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: testHandlerLambdaRequestID,
	})
	event := &content.ProcessArticleEvent{
		RequestID: testHandlerRequestID,
		URL:       testHandlerURL,
		ArticleID: testHandlerArticleID,
		AccountID: testHandlerAccountID,
		InheritedAttrs: []map[string]any{
			{testUserIDKey: testHandlerUserID},
			{testHandlerSourceKey: testHandlerSourceAPI},
		},
	}

	eventURL, _ := url.Parse(event.URL)
	htmlBytes := []byte("<html><body>Test</body></html>")
	article := &model.Article{Title: testHandlerTestArticle}
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

	err := handleEvent(ctx, event, mockSvc)

	assert.NoError(t, err)
	assert.Contains(t, event.InheritedAttrs, map[string]any{attrKeyOrigRequestID: testHandlerRequestID})
	assert.Contains(t, event.InheritedAttrs, map[string]any{attrKeyRequestID: testHandlerLambdaRequestID})
	assert.NotContains(t, event.InheritedAttrs, map[string]any{testHandlerSourceKey: testHandlerSourceAPI})
}

var _ processor.Service = (*mockService)(nil)
