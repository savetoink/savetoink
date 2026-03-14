package task

import (
	"context"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackgroundScheduler_Start_DuplicateTasks_Ignored(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	assert.NotNil(t, runner)

	task1Count := 0
	runner.Register(Task{
		Name: "task1",
		Run: func(_ context.Context, _ map[string]ParamValue) *RunResult {
			task1Count++
			return &RunResult{Output: []string{"task1 completed"}}
		},
	})

	runner.Register(Task{
		Name: "task2",
		Run: func(_ context.Context, _ map[string]ParamValue) *RunResult {
			return &RunResult{Output: []string{"task2 completed"}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Name:     "task1",
			Enabled:  true,
			Schedule: "0 0 * * * *",
		},
		{
			Name:     "task1",
			Enabled:  true,
			Schedule: "0 */5 * * * *",
		},
		{
			Name:     "task2",
			Enabled:  true,
			Schedule: "0 0 * * * *",
		},
	}

	scheduler := NewBackgroundScheduler(runner, configs)
	ctx := context.Background()

	err := scheduler.Start(ctx)
	require.NoError(t, err)
	defer scheduler.Stop()

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, task1Count)
}

func TestBackgroundScheduler_Start_EnabledTasksOnly(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	taskExecuted := false
	runner.Register(Task{
		Name: "task1",
		Run: func(_ context.Context, _ map[string]ParamValue) *RunResult {
			taskExecuted = true
			return &RunResult{Output: []string{"task1 completed"}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Name:     "task1",
			Enabled:  true,
			Schedule: "*/1 * * * * *",
		},
		{
			Name:     "task2",
			Enabled:  false,
			Schedule: "*/1 * * * * *",
		},
	}

	scheduler := NewBackgroundScheduler(runner, configs)
	ctx := context.Background()

	err := scheduler.Start(ctx)
	require.NoError(t, err)
	defer scheduler.Stop()

	time.Sleep(1500 * time.Millisecond)
	assert.True(t, taskExecuted)
}

func TestBackgroundScheduler_Start_InvalidSchedule(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	runner.Register(Task{
		Name: "task1",
		Run: func(_ context.Context, _ map[string]ParamValue) *RunResult {
			return &RunResult{Output: []string{"task1 completed"}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Name:     "task1",
			Enabled:  true,
			Schedule: "invalid schedule",
		},
	}

	scheduler := NewBackgroundScheduler(runner, configs)
	ctx := context.Background()

	err := scheduler.Start(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scheduling task")
}

func TestBackgroundScheduler_Start_WithParams(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	paramsExecuted := false
	runner.Register(Task{
		Name: "task1",
		Run: func(_ context.Context, params map[string]ParamValue) *RunResult {
			backupName, err := RequireString(params, "backup_name")
			if err != nil {
				return &RunResult{Error: err}
			}
			if backupName == "test-backup" {
				paramsExecuted = true
			}
			return &RunResult{Output: []string{"processed: " + backupName}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Name:     "task1",
			Enabled:  true,
			Schedule: "*/1 * * * * *",
			Params: map[string]string{
				"backup_name": "test-backup",
			},
		},
	}

	scheduler := NewBackgroundScheduler(runner, configs)
	ctx := context.Background()

	err := scheduler.Start(ctx)
	require.NoError(t, err)
	defer scheduler.Stop()

	time.Sleep(1500 * time.Millisecond)
	assert.True(t, paramsExecuted)
}

func TestBackgroundScheduler_Stop(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	runner.Register(Task{
		Name: "task1",
		Run: func(_ context.Context, _ map[string]ParamValue) *RunResult {
			return &RunResult{Output: []string{"task1 completed"}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Name:     "task1",
			Enabled:  true,
			Schedule: "0 0 * * * *",
		},
	}

	scheduler := NewBackgroundScheduler(runner, configs)
	ctx := context.Background()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	scheduler.Stop()
}
