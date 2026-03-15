package scheduler

import (
	"context"
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
	require.NotNil(t, runResult.Error)
	assert.Contains(t, runResult.Error.Error(), "restore failed")
	assert.Contains(t, runResult.Output[0], "restore failed")
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
	assert.Nil(t, runResult.Error)
	require.Len(t, runResult.Output, 2)
	assert.Contains(t, runResult.Output[0], "restore completed for table test-table from backup test-backup")
	assert.Contains(t, runResult.Output[0], "overwrite: false")
	assert.Contains(t, runResult.Output[0], "100 items restored")
	assert.Contains(t, runResult.Output[1], "50 items skipped")
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
	assert.Nil(t, runResult.Error)
	require.Len(t, runResult.Output, 2)
	assert.Contains(t, runResult.Output[0], "restore completed for table test-table from backup test-backup")
	assert.Contains(t, runResult.Output[0], "overwrite: true")
	assert.Contains(t, runResult.Output[0], "100 items restored")
	assert.Contains(t, runResult.Output[1], "20 items deleted")
}

func TestRegisterTasks(t *testing.T) {
	runner := task.NewTaskRunner(getTestConfig())
	require.NotNil(t, runner)

	RegisterTasks(runner, getTestConfig())

	result := runner.Run(context.Background(), "backup", "0 0 0 * * *", nil)
	require.NotNil(t, result)

	resultRestore := runner.Run(context.Background(), "restore", "0 0 0 * * *", map[string]task.ParamValue{
		"backup_name": task.StringParam("test"),
	})
	require.NotNil(t, resultRestore)
}

func TestRegisterTasks_BackupTaskExecutes(t *testing.T) {
	runner := task.NewTaskRunner(getTestConfig())
	require.NotNil(t, runner)

	RegisterTasks(runner, getTestConfig())

	result := runner.Run(context.Background(), "backup", "", nil)

	require.NotNil(t, result)
	assert.NotNil(t, result.Output)
}

func TestRegisterTasks_RestoreTaskExecutes_Success(t *testing.T) {
	runner := task.NewTaskRunner(getTestConfig())
	require.NotNil(t, runner)

	RegisterTasks(runner, getTestConfig())

	params := map[string]task.ParamValue{
		"backup_name": task.StringParam("test-backup"),
		"overwrite":   task.StringParam("false"),
	}

	result := runner.Run(context.Background(), "restore", "", params)

	require.NotNil(t, result)
	assert.NotNil(t, result.Output)
}

func TestRegisterTasks_RestoreTaskExecutes_MissingRequiredParam(t *testing.T) {
	runner := task.NewTaskRunner(getTestConfig())
	require.NotNil(t, runner)

	RegisterTasks(runner, getTestConfig())

	params := map[string]task.ParamValue{
		"overwrite": task.StringParam("true"),
	}

	result := runner.Run(context.Background(), "restore", "", params)

	require.NotNil(t, result)
	require.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "required parameter 'backup_name' not found")
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
			Name:     "test-task",
			Schedule: "0 0 0 * * *",
			Enabled:  true,
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
				Name:     "test-task",
				Schedule: "0 0 0 * * *",
				Enabled:  true,
			},
		},
	}

	scheduler := NewBackgroundScheduler(cfg)

	assert.Nil(t, scheduler)
}
