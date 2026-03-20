package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testFunctionName = "test-function"
	testRequestID    = "req-123"
	testURL          = "https://example.com/article"
	testArticleID    = "article-456"
	testAccountID    = "account-789"
)

type mockLambdaInvoker struct {
	invoked      bool
	invokeError  error
	payload      []byte
	functionName string
}

func (m *mockLambdaInvoker) Invoke(
	_ context.Context,
	params *lambda.InvokeInput,
	_ ...func(*lambda.Options),
) (*lambda.InvokeOutput, error) {
	m.invoked = true
	m.functionName = aws.ToString(params.FunctionName)
	m.payload = params.Payload

	if m.invokeError != nil {
		return nil, m.invokeError
	}

	return &lambda.InvokeOutput{
		StatusCode: 202,
	}, nil
}

func TestNewProcessor(t *testing.T) {
	awsCfg := aws.Config{
		Region: "us-east-1",
	}
	proc := NewProcessor(testFunctionName, &awsCfg)

	assert.NotNil(t, proc)
	assert.Equal(t, testFunctionName, proc.functionName)
	assert.NotNil(t, proc.lambdaInvoker)
}

func TestStartProcessing_Success(t *testing.T) {
	mockInvoker := &mockLambdaInvoker{}
	proc := newProcessorWithInvoker(testFunctionName, mockInvoker)

	ctx := context.Background()
	event := &content.ProcessArticleEvent{
		RequestID:      testRequestID,
		URL:            testURL,
		ArticleID:      testArticleID,
		AccountID:      testAccountID,
		SendOnComplete: true,
	}

	proc.StartProcessing(ctx, event)

	assert.True(t, mockInvoker.invoked)
	assert.Equal(t, testFunctionName, mockInvoker.functionName)

	var unmarshaled content.ProcessArticleEvent
	err := json.Unmarshal(mockInvoker.payload, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, event.RequestID, unmarshaled.RequestID)
	assert.Equal(t, event.URL, unmarshaled.URL)
	assert.Equal(t, event.ArticleID, unmarshaled.ArticleID)
	assert.Equal(t, event.AccountID, unmarshaled.AccountID)
	assert.Equal(t, event.SendOnComplete, unmarshaled.SendOnComplete)
}

func TestStartProcessing_MarshalError(t *testing.T) {
	mockInvoker := &mockLambdaInvoker{}
	proc := newProcessorWithInvoker(testFunctionName, mockInvoker)

	ctx := context.Background()

	event := &content.ProcessArticleEvent{
		RequestID: testRequestID,
		URL:       testURL,
		ArticleID: testArticleID,
		AccountID: testAccountID,
		InheritedAttrs: []map[string]any{
			{"key": make(chan int)},
		},
	}

	proc.StartProcessing(ctx, event)

	assert.False(t, mockInvoker.invoked)
	assert.Nil(t, mockInvoker.payload)
}

func TestStartProcessing_InvokeError(t *testing.T) {
	mockInvoker := &mockLambdaInvoker{
		invokeError: errors.New("lambda invoke failed"),
	}
	proc := newProcessorWithInvoker(testFunctionName, mockInvoker)

	ctx := context.Background()
	event := &content.ProcessArticleEvent{
		RequestID: testRequestID,
		URL:       testURL,
		ArticleID: testArticleID,
		AccountID: testAccountID,
	}

	proc.StartProcessing(ctx, event)

	assert.True(t, mockInvoker.invoked)
	assert.NotNil(t, mockInvoker.payload)
}

func TestStartProcessing_NilEvent(t *testing.T) {
	mockInvoker := &mockLambdaInvoker{}
	proc := newProcessorWithInvoker(testFunctionName, mockInvoker)

	ctx := context.Background()

	proc.StartProcessing(ctx, nil)

	assert.True(t, mockInvoker.invoked)
	assert.NotNil(t, mockInvoker.payload)
	assert.Equal(t, []byte("null"), mockInvoker.payload)
}

func TestStartProcessing_WithInheritedAttrs(t *testing.T) {
	mockInvoker := &mockLambdaInvoker{}
	proc := newProcessorWithInvoker(testFunctionName, mockInvoker)

	ctx := context.Background()
	event := &content.ProcessArticleEvent{
		RequestID: testRequestID,
		URL:       testURL,
		ArticleID: testArticleID,
		AccountID: testAccountID,
		InheritedAttrs: []map[string]any{
			{"user_id": "user-123"},
			{"source": "api"},
		},
		SendOnComplete: false,
	}

	proc.StartProcessing(ctx, event)

	assert.True(t, mockInvoker.invoked)

	var unmarshaled content.ProcessArticleEvent
	err := json.Unmarshal(mockInvoker.payload, &unmarshaled)
	require.NoError(t, err)
	assert.Len(t, unmarshaled.InheritedAttrs, 2)
}

func TestStartProcessing_ContextCancellation(t *testing.T) {
	mockInvoker := &mockLambdaInvoker{
		invokeError: context.Canceled,
	}
	proc := newProcessorWithInvoker(testFunctionName, mockInvoker)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	event := &content.ProcessArticleEvent{
		RequestID: testRequestID,
		URL:       testURL,
		ArticleID: testArticleID,
		AccountID: testAccountID,
	}

	proc.StartProcessing(ctx, event)

	assert.True(t, mockInvoker.invoked)
}

func TestProcessor_ImplementsInterface(_ *testing.T) {
	var _ processor.Processor = (*Processor)(nil)
}
