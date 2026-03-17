package lambda

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestConfig() *config.Config {
	return &config.Config{
		AWSConfig:        &aws.Config{},
		ArticlesTable:    "test-articles",
		UserProfileTable: "test-profiles",
		SendsTable:       "test-sends",
		StorageBackend:   "dynamodb",
		EmailProvider:    "mailjet",
		MailjetAPIKey:    "test-key",
		MailjetAPISecret: "test-secret",
		SenderEmail:      "test@example.com",
		AppURL:           "https://test.com",
	}
}

func TestNewHandler_ReturnsHandler(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)
	require.NotNil(t, handler)
}

func TestHandler_WithValidContext(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "test-request-id",
	})

	ev := []byte(`{"task":"backup","schedule":"0 0 0 * * *"}`)

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Results)
}

func TestHandler_WithNilAWSConfig(t *testing.T) {
	cfg := &config.Config{
		AWSConfig: nil,
	}
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := []byte(`{"task":"backup"}`)

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "task runner not initialized")
}

func TestHandler_WithInvalidTask(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := []byte(`{"task":"nonexistent-task","schedule":"0 0 0 * * *"}`)

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "unknown task")
}

func TestHandler_WithTaskParams(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := []byte(`{"task":"restore","schedule":"0 0 0 * * *","backup_name":"test-backup","overwrite":true}`)

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Errors)
}

func TestHandler_WithMissingRequiredParam(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := []byte(`{"task":"restore","schedule":"0 0 0 * * *","overwrite":true}`)

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "required parameter")
}

func TestHandler_WithEmptyParams(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := []byte(`{"task":"backup","schedule":"0 0 0 * * *"}`)

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Results)
}

func TestHandler_Integration_BackupTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := []byte(`{"task":"backup","schedule":"0 0 0 * * *"}`)

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Results)
}
