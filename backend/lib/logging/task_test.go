package logging

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRunID(t *testing.T) {
	id1 := GenerateRunID()
	id2 := GenerateRunID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestLogTaskExecution_Success(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	start := time.Now()

	LogTaskExecution(ctx, "test_task", "run-123", start, []string{}, []string{"task output completed"}, nil)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelInfo, record.Level)
	assert.Equal(t, "task execution complete", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "test_task", attrMap["task_name"])
	assert.Equal(t, "run-123", attrMap["run_id"])
	assert.Contains(t, attrMap, "latency")
	assert.Equal(t, "task output completed", attrMap["output"])
	assert.NotContains(t, attrMap, "error")
	assert.NotContains(t, attrMap, "scheduled_next")
}

func TestLogTaskExecution_WithScheduledNext(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	start := time.Now()
	nextRun := time.Now().Add(1 * time.Hour)

	LogTaskExecution(ctx, "test_task", "run-123", start, []string{}, []string{"task output completed"}, &nextRun)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "test_task", attrMap["task_name"])
	assert.Equal(t, "run-123", attrMap["run_id"])
	assert.Equal(t, "task output completed", attrMap["output"])
	scheduledNext, ok := attrMap["scheduled_next"].(time.Time)
	require.True(t, ok)
	assert.WithinDuration(t, nextRun, scheduledNext, time.Second)
}

func TestLogTaskExecution_WithError(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	start := time.Now()
	testErr := "task failed"

	LogTaskExecution(ctx, "test_task", "run-456", start, []string{testErr}, nil, nil)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelError, record.Level)
	assert.Equal(t, "task execution complete", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "test_task", attrMap["task_name"])
	assert.Equal(t, "run-456", attrMap["run_id"])
	assert.Contains(t, attrMap, "latency")
	assert.Equal(t, testErr, attrMap["errors"])
	assert.NotContains(t, attrMap, "output")
}

func TestLogTaskExecution_WithErrorAndOutput(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	start := time.Now()
	testErr := errors.New("partial failure")

	LogTaskExecution(ctx, "test_task", "run-789", start, []string{testErr.Error()}, []string{"partial output"}, nil)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelError, record.Level)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, testErr.Error(), attrMap["errors"])
	assert.Equal(t, "partial output", attrMap["output"])
}

func TestLogTaskExecution_NoOutput(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	start := time.Now()

	LogTaskExecution(ctx, "test_task", "run-abc", start, []string{}, nil, nil)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.NotContains(t, attrMap, "output")
}

func TestLogTaskExecution_Latency(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	start := time.Now()

	time.Sleep(10 * time.Millisecond)
	LogTaskExecution(ctx, "test_task", "run-latency", start, []string{}, nil, nil)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	latency, ok := attrMap["latency"].(time.Duration)
	require.True(t, ok)
	assert.Greater(t, latency.Milliseconds(), int64(0))
}

func TestLogTaskExecution_HandlerError(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	start := time.Now()

	assert.NotPanics(t, func() {
		LogTaskExecution(ctx, "test_task", "run-error", start, []string{}, []string{"output"}, nil)
	})

	require.Len(t, capture.records, 1)
}

func TestLogSchedulerStarted(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	tasks := map[string]struct{}{"task1": {}, "task2": {}, "task3": {}}

	LogSchedulerStarted(ctx, tasks)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelInfo, record.Level)
	assert.Equal(t, "background scheduler started", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, []string{"task1", "task2", "task3"}, attrMap["tasks"])
}

func TestLogSchedulerStarted_EmptyTasks(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	tasks := map[string]struct{}{}

	LogSchedulerStarted(ctx, tasks)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Nil(t, attrMap["tasks"])
}

func TestLogTaskFailed(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	testErr := errors.New("unknown task: missing_task")

	LogTaskFailed(ctx, "missing_task", testErr, nil)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelError, record.Level)
	assert.Equal(t, "task execution failed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "missing_task", attrMap["task_name"])
	assert.Equal(t, testErr.Error(), attrMap["error"])
	assert.NotContains(t, attrMap, "scheduled_next")
}

func TestLogTaskFailed_WithScheduledNext(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	testErr := errors.New("task failed")
	nextRun := time.Now().Add(30 * time.Minute)

	LogTaskFailed(ctx, "test_task", testErr, &nextRun)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "test_task", attrMap["task_name"])
	assert.Equal(t, testErr.Error(), attrMap["error"])
	scheduledNext, ok := attrMap["scheduled_next"].(time.Time)
	require.True(t, ok)
	assert.WithinDuration(t, nextRun, scheduledNext, time.Second)
}

func TestLogTaskUnknown_HandlerError(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()
	testErr := errors.New("test error")

	assert.NotPanics(t, func() {
		LogTaskFailed(ctx, "unknown_task", testErr, nil)
	})

	require.Len(t, capture.records, 1)
}
