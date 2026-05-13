package task

import (
	"context"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	scheduleHourly     = "0 0 * * * *"
	scheduleEveryMin   = "*/1 * * * * *"
	testTask1          = "task1"
	testTask2          = "task2"
	testTask1Completed = "task1 completed"
	testBackupName     = "test-backup"
)

func TestBackgroundScheduler_Start_DuplicateTasks_Ignored(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	assert.NotNil(t, runner)

	task1Count := 0
	runner.Register(Task{
		Name: testTask1,
		Run: func(_ context.Context, _ consts.TaskConfig) *RunResult {
			task1Count++
			return &RunResult{Results: []string{testTask1Completed}}
		},
	})

	runner.Register(Task{
		Name: testTask2,
		Run: func(_ context.Context, _ consts.TaskConfig) *RunResult {
			return &RunResult{Results: []string{"task2 completed"}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Task:     testTask1,
			Schedule: scheduleHourly,
		},
		{
			Task:     testTask1,
			Schedule: testSchedule5min,
		},
		{
			Task:     testTask2,
			Schedule: scheduleHourly,
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

func TestBackgroundScheduler_Start_ScheduledTasks(t *testing.T) {
	runner := NewTaskRunner(getTestConfig(t))

	taskExecuted := false
	runner.Register(Task{
		Name: testTask1,
		Run: func(_ context.Context, _ consts.TaskConfig) *RunResult {
			taskExecuted = true
			return &RunResult{Results: []string{testTask1Completed}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Task:     testTask1,
			Schedule: scheduleEveryMin,
		},
		{
			Task:     testTask2,
			Schedule: scheduleEveryMin,
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
		Name: testTask1,
		Run: func(_ context.Context, _ consts.TaskConfig) *RunResult {
			return &RunResult{Results: []string{testTask1Completed}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Task:     testTask1,
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
		Name: testTask1,
		Run: func(_ context.Context, cfg consts.TaskConfig) *RunResult {
			if cfg.BackupName == testBackupName {
				paramsExecuted = true
			}
			return &RunResult{Results: []string{"processed: " + cfg.BackupName}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Task:       testTask1,
			Schedule:   scheduleEveryMin,
			BackupName: testBackupName,
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
		Name: testTask1,
		Run: func(_ context.Context, _ consts.TaskConfig) *RunResult {
			return &RunResult{Results: []string{testTask1Completed}}
		},
	})

	configs := []consts.TaskConfig{
		{
			Task:     testTask1,
			Schedule: scheduleHourly,
		},
	}

	scheduler := NewBackgroundScheduler(runner, configs)
	ctx := context.Background()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	scheduler.Stop()
}
