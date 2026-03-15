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

	ev := event{
		Task:     "backup",
		Schedule: "0 0 0 * * *",
	}

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Output)
}

func TestHandler_WithNilAWSConfig(t *testing.T) {
	cfg := &config.Config{
		AWSConfig: nil,
	}
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := event{
		Task: "backup",
	}

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "task runner not initialized")
}

func TestHandler_WithInvalidTask(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := event{
		Task:     "nonexistent-task",
		Schedule: "0 0 0 * * *",
	}

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "unknown task")
}

func TestHandler_WithTaskParams(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := event{
		Task:     "restore",
		Schedule: "0 0 0 * * *",
		Params: map[string]map[string]any{
			"backup_name": {"value": "test-backup"},
			"overwrite":   {"value": "true"},
		},
	}

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Output)
}

func TestHandler_WithMissingRequiredParam(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := event{
		Task:     "restore",
		Schedule: "0 0 0 * * *",
		Params: map[string]map[string]any{
			"overwrite": {"value": "true"},
		},
	}

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "required parameter")
}

func TestHandler_WithInvalidParamType(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := event{
		Task:     "restore",
		Schedule: "0 0 0 * * *",
		Params: map[string]map[string]any{
			"backup_name": {"value": 123},
		},
	}

	result, err := handler(ctx, ev)

	require.Error(t, err)
	require.NotNil(t, result)
	assert.Contains(t, err.Error(), "unsupported parameter type")
	assert.NotNil(t, result.Error)
}

func TestHandler_WithEmptyParams(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := event{
		Task:     "backup",
		Schedule: "0 0 0 * * *",
		Params:   map[string]map[string]any{},
	}

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Output)
}

func TestHandler_WithNilParams(t *testing.T) {
	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := event{
		Task:   "backup",
		Params: nil,
	}

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Output)
}

func TestParseTaskParams_Nil(t *testing.T) {
	result, err := parseTaskParams(nil)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestParseTaskParams_Empty(t *testing.T) {
	result, err := parseTaskParams(map[string]map[string]any{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestParseTaskParams_StringValue(t *testing.T) {
	params := map[string]map[string]any{
		"key1": {"value": "string-value"},
	}

	result, err := parseTaskParams(params)
	require.NoError(t, err)
	require.Len(t, result, 1)

	val, err := result["key1"].String()
	require.NoError(t, err)
	assert.Equal(t, "string-value", val)
}

func TestParseTaskParams_UnsupportedType(t *testing.T) {
	params := map[string]map[string]any{
		"key1": {"value": 123},
	}

	result, err := parseTaskParams(params)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported parameter type")
}

func TestParseTaskParams_MultipleValues_Ignored(t *testing.T) {
	params := map[string]map[string]any{
		"key1": {"value": "first", "extra": "ignored"},
	}

	result, err := parseTaskParams(params)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestParseTaskParams_MultipleKeys(t *testing.T) {
	params := map[string]map[string]any{
		"key1": {"value": "value1"},
		"key2": {"value": "value2"},
	}

	result, err := parseTaskParams(params)
	require.NoError(t, err)
	require.Len(t, result, 2)

	val1, err := result["key1"].String()
	require.NoError(t, err)
	assert.Equal(t, "value1", val1)

	val2, err := result["key2"].String()
	require.NoError(t, err)
	assert.Equal(t, "value2", val2)
}

func TestHandler_Integration_BackupTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := getTestConfig()
	handler := NewHandler(cfg)

	ctx := context.Background()
	ev := event{
		Task:     "backup",
		Schedule: "0 0 0 * * *",
	}

	result, err := handler(ctx, ev)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Output)
}
