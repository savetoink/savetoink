package task

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCronScheduleHourly = "0 0 * * * *"
	testSchedule5min       = "0 */5 * * * *"
	testTaskName           = "test_task"
	testTaskCompleted      = "task completed"
)

func getTestConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		AWSConfig: &aws.Config{},
	}
}

func TestCalculateNextRun_Success(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	tests := []struct {
		name            string
		schedule        string
		wantWithinRange bool
	}{
		{
			name:            "hourly schedule",
			schedule:        testCronScheduleHourly,
			wantWithinRange: true,
		},
		{
			name:            "daily schedule",
			schedule:        "0 0 0 * * *",
			wantWithinRange: true,
		},
		{
			name:            "every 5 minutes",
			schedule:        testSchedule5min,
			wantWithinRange: true,
		},
		{
			name:            "every 30 seconds",
			schedule:        "*/30 * * * * *",
			wantWithinRange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextRun, err := runner.calculateNextRun(tt.schedule)
			require.NoError(t, err)
			require.NotNil(t, nextRun)
			assert.True(t, nextRun.After(time.Now()))
		})
	}
}

func TestCalculateNextRun_EmptySchedule(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	nextRun, err := runner.calculateNextRun("")
	require.NoError(t, err)
	assert.Nil(t, nextRun)
}

func TestCalculateNextRun_InvalidSchedule(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	tests := []struct {
		name     string
		schedule string
		wantErr  bool
	}{
		{
			name:     "invalid format",
			schedule: "invalid",
			wantErr:  true,
		},
		{
			name:     "too many fields",
			schedule: "* * * * * * *",
			wantErr:  true,
		},
		{
			name:     "invalid second",
			schedule: "60 * * * * *",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextRun, err := runner.calculateNextRun(tt.schedule)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, nextRun)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRun_WithSchedule(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))
	taskExecuted := false

	runner.Register(Task{
		Name: testTaskName,
		Run: func(_ context.Context, _ consts.TaskConfig) *RunResult {
			taskExecuted = true
			return &RunResult{Results: []string{testTaskCompleted}}
		},
	})

	ctx := context.Background()
	result := runner.Run(ctx, json.RawMessage(fmt.Sprintf(`{"task":%q,"schedule":"0 0 * * * *"}`, testTaskName)))
	require.Empty(t, result.Errors)
	assert.True(t, taskExecuted)
}

func TestRun_WithInvalidSchedule(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	runner.Register(Task{
		Name: testTaskName,
		Run: func(_ context.Context, _ consts.TaskConfig) *RunResult {
			return &RunResult{Results: []string{testTaskCompleted}}
		},
	})

	ctx := context.Background()

	result := runner.Run(ctx, json.RawMessage(fmt.Sprintf(`{"task":%q,"schedule":"invalid schedule"}`, testTaskName)))
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "failed to calculate next run time")
}

func TestRun_WithEmptySchedule(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))
	taskExecuted := false

	runner.Register(Task{
		Name: testTaskName,
		Run: func(_ context.Context, _ consts.TaskConfig) *RunResult {
			taskExecuted = true
			return &RunResult{Results: []string{testTaskCompleted}}
		},
	})

	ctx := context.Background()

	result := runner.Run(ctx, json.RawMessage(fmt.Sprintf(`{"task":%q,"schedule":""}`, testTaskName)))
	require.Empty(t, result.Errors)
	assert.True(t, taskExecuted)
}

func TestRun_UnknownTask(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	ctx := context.Background()
	result := runner.Run(ctx, json.RawMessage(`{"task":"unknown_task","schedule":"0 0 * * * *"}`))
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "unknown task")
}
