package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	repo "github.com/shaftoe/savetoink/backend/lib/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestConfig() *config.Config {
	return &config.Config{
		AWSConfig:        &aws.Config{},
		ArticlesTable:    "test-articles",
		UserProfileTable: "test-profiles",
		SendsTable:       "test-sends",
		BackupBucketName: "test-backup-bucket",
		StorageBackend:   "dynamodb",
		EmailProvider:    "mailjet",
		MailjetAPIKey:    "test-key",
		MailjetAPISecret: "test-secret",
		SenderEmail:      "test@example.com",
		AppURL:           "https://test.com",
	}
}

func TestFormatBackupResults_AllSuccess(t *testing.T) {
	results := []*repo.BackupResult{
		{
			TableName:  "table1",
			ItemsCount: 100,
			Key:        "backup/key1",
			Latency:    100 * time.Millisecond,
			Error:      nil,
		},
		{
			TableName:  "table2",
			ItemsCount: 200,
			Key:        "backup/key2",
			Latency:    200 * time.Millisecond,
			Error:      nil,
		},
	}

	out, err := formatBackupResults(results)

	require.NoError(t, err)
	require.Len(t, out, 3)
	assert.Contains(t, out[0], "table1")
	assert.Contains(t, out[1], "table2")
	assert.Contains(t, out[2], "backup summary: 2 succeeded, 0 failed")
}

func TestFormatBackupResults_AllFailed(t *testing.T) {
	results := []*repo.BackupResult{
		{
			TableName: "table1",
			Error:     errors.New("failed to backup"),
		},
		{
			TableName: "table2",
			Error:     errors.New("failed to backup"),
		},
	}

	out, err := formatBackupResults(results)

	require.Error(t, err)
	require.Len(t, out, 1)
	assert.Contains(t, out[0], "backup summary: 0 succeeded, 2 failed")
	assert.Contains(t, err.Error(), "failed to backup table table1")
	assert.Contains(t, err.Error(), "failed to backup table table2")
}

func TestFormatBackupResults_MixedSuccessAndFailure(t *testing.T) {
	results := []*repo.BackupResult{
		{
			TableName:  "table1",
			ItemsCount: 100,
			Key:        "backup/key1",
			Latency:    100 * time.Millisecond,
			Error:      nil,
		},
		{
			TableName: "table2",
			Error:     errors.New("failed to backup"),
		},
	}

	out, err := formatBackupResults(results)

	require.Error(t, err)
	require.Len(t, out, 2)
	assert.Contains(t, out[0], "table1")
	assert.Contains(t, out[1], "backup summary: 1 succeeded, 1 failed")
}

func TestFormatBackupResults_EmptyResults(t *testing.T) {
	results := []*repo.BackupResult{}

	out, err := formatBackupResults(results)

	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Contains(t, out[0], "backup summary: 0 succeeded, 0 failed")
}

func TestFormatRestoreResult_Error(t *testing.T) {
	result := &repo.RestoreResult{
		TableName:  "test-table",
		BackupName: "test-backup",
		Error:      errors.New("restore failed"),
	}

	runResult := formatRestoreResult(result)

	require.NotNil(t, runResult)
	require.Len(t, runResult.Errors, 1)
	assert.Contains(t, runResult.Errors[0], "restore failed")
	assert.Empty(t, runResult.Results)
}

func TestFormatRestoreResult_Success_NoOverwrite(t *testing.T) {
	result := &repo.RestoreResult{
		TableName:     "test-table",
		BackupName:    "test-backup",
		ItemsRestored: 100,
		ItemsSkipped:  50,
		Overwrite:     false,
		Latency:       500 * time.Millisecond,
	}

	runResult := formatRestoreResult(result)

	require.NotNil(t, runResult)
	assert.Empty(t, runResult.Errors)
	require.Len(t, runResult.Results, 2)
	assert.Contains(t, runResult.Results[0], "restore completed for table test-table from backup test-backup")
	assert.Contains(t, runResult.Results[0], "overwrite: false")
	assert.Contains(t, runResult.Results[0], "100 items restored")
	assert.Contains(t, runResult.Results[1], "50 items skipped")
}

func TestFormatRestoreResult_Success_Overwrite(t *testing.T) {
	result := &repo.RestoreResult{
		TableName:     "test-table",
		BackupName:    "test-backup",
		ItemsRestored: 100,
		ItemsDeleted:  20,
		Overwrite:     true,
		Latency:       500 * time.Millisecond,
	}

	runResult := formatRestoreResult(result)

	require.NotNil(t, runResult)
	assert.Empty(t, runResult.Errors)
	require.Len(t, runResult.Results, 2)
	assert.Contains(t, runResult.Results[0], "restore completed for table test-table from backup test-backup")
	assert.Contains(t, runResult.Results[0], "overwrite: true")
	assert.Contains(t, runResult.Results[0], "100 items restored")
	assert.Contains(t, runResult.Results[1], "20 items deleted")
}

func TestRegisterTasks(t *testing.T) {
	runner := task.NewTaskRunner(getTestConfig())
	require.NotNil(t, runner)

	RegisterTasks(runner, getTestConfig())

	result := runner.Run(context.Background(), json.RawMessage(`{"task":"backup","schedule":"0 0 0 * * *"}`))
	require.NotNil(t, result)

	restoreEvent := json.RawMessage(`{"task":"restore","schedule":"0 0 0 * * *","params":{"backup_name":"test"}}`)
	resultRestore := runner.Run(context.Background(), restoreEvent)
	require.NotNil(t, resultRestore)
}

func TestRegisterTasks_BackupTaskExecutes(t *testing.T) {
	runner := task.NewTaskRunner(getTestConfig())
	require.NotNil(t, runner)

	RegisterTasks(runner, getTestConfig())

	result := runner.Run(context.Background(), json.RawMessage(`{"task":"backup","schedule":""}`))

	require.NotNil(t, result)
	assert.NotNil(t, result.Results)
}

func TestRegisterTasks_RestoreTaskExecutes_Success(t *testing.T) {
	runner := task.NewTaskRunner(getTestConfig())
	require.NotNil(t, runner)

	RegisterTasks(runner, getTestConfig())

	restoreJSON := `{"task":"restore","schedule":"","backup_name":"test-backup.json","overwrite":false}`
	restoreEvent := json.RawMessage(restoreJSON)
	result := runner.Run(context.Background(), restoreEvent)

	require.NotNil(t, result)
	require.NotEmpty(t, result.Errors)
}

func TestRegisterTasks_RestoreTaskExecutes_MissingRequiredParam(t *testing.T) {
	runner := task.NewTaskRunner(getTestConfig())
	require.NotNil(t, runner)

	RegisterTasks(runner, getTestConfig())

	restoreEvent := json.RawMessage(`{"task":"restore","schedule":"","overwrite":true}`)
	result := runner.Run(context.Background(), restoreEvent)

	require.NotNil(t, result)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "required parameter 'backup_name' not found")
}

func TestNewBackgroundScheduler_NoTasks(t *testing.T) {
	cfg := &config.Config{
		Tasks: nil,
	}

	scheduler := NewBackgroundScheduler(cfg)

	assert.Nil(t, scheduler)
}

func TestNewBackgroundScheduler_EmptyTasks(t *testing.T) {
	cfg := &config.Config{
		Tasks: []consts.TaskConfig{},
	}

	scheduler := NewBackgroundScheduler(cfg)

	assert.Nil(t, scheduler)
}

func TestNewBackgroundScheduler_WithTasks(t *testing.T) {
	cfg := getTestConfig()
	cfg.Tasks = []consts.TaskConfig{
		{
			Task:     "st-task",
			Schedule: "0 0 0 * * *",
		},
	}

	scheduler := NewBackgroundScheduler(cfg)

	require.NotNil(t, scheduler)
}

func TestNewBackgroundScheduler_NilRunner(t *testing.T) {
	cfg := &config.Config{
		AWSConfig: nil,
		Tasks: []consts.TaskConfig{
			{
				Task:     "st-task",
				Schedule: "0 0 0 * * *",
			},
		},
	}

	scheduler := NewBackgroundScheduler(cfg)

	assert.Nil(t, scheduler)
}
