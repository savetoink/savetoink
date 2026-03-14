package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	repository "github.com/shaftoe/savetoink/backend/lib/repository/dynamodb"
)

// Task represents a scheduled task that can be executed.
type Task struct {
	Name string
	Run  func(ctx context.Context) *RunResult
}

// RunResult contains the output and any error from task execution.
type RunResult struct {
	Output string
	Error  error
	ID     string
}

// TaskRunner manages and executes registered tasks.
type TaskRunner struct { //nolint:revive // task.TaskRunner is acceptable naming
	tasks   map[string]Task
	config  *config.Config
	BkpRepo *repository.BackupRepository
}

// NewTaskRunner creates a new TaskRunner with the given configuration.
// Returns nil if AWSConfig is not configured.
func NewTaskRunner(cfg *config.Config) *TaskRunner {
	if cfg.AWSConfig == nil {
		return nil
	}
	bkpRepo := repository.NewBackupRepository(cfg)

	return &TaskRunner{tasks: make(map[string]Task), config: cfg, BkpRepo: bkpRepo}
}

// Register adds a task to the runner's task registry.
func (r *TaskRunner) Register(t Task) {
	r.tasks[t.Name] = t
}

func (r *TaskRunner) calculateNextRun(schedule string) (*time.Time, error) {
	if schedule == "" {
		return nil, nil
	}
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	parsed, err := parser.Parse(schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schedule %q: %w", schedule, err)
	}
	next := parsed.Next(time.Now())
	return &next, nil
}

// Run executes a task with the given name and schedule, logging the result.
func (r *TaskRunner) Run(ctx context.Context, name, schedule string) *RunResult {
	runID := logging.GenerateRunID()
	t, ok := r.tasks[name]
	if !ok {
		slog.With("run_id", runID).Error("unknown task")
		return &RunResult{
			Error: fmt.Errorf("unknown task: %s", name),
			ID:    runID,
		}
	}

	scheduledNext, err := r.calculateNextRun(schedule)
	if err != nil {
		err = fmt.Errorf("failed to calculate next run time: %w", err)
		slog.With("run_id", runID).Error(err.Error(),
			slog.String("schedule", schedule))
		return &RunResult{
			Error: err,
			ID:    runID,
		}
	}

	start := time.Now()
	result := t.Run(ctx)
	result.ID = runID
	logging.LogTaskExecution(ctx, t.Name, runID, start, result.Error, result.Output, scheduledNext)
	return result
}
