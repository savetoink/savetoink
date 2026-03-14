// Package scheduler provides configuration and initialization for background task scheduling.
package scheduler

import (
	"context"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/internal/task"
)

// RegisterTasks registers all available tasks with the given task runner.
func RegisterTasks(runner *task.TaskRunner) {
	runner.Register(task.Task{
		Name: "backup",
		Run: func(_ context.Context) *task.RunResult {
			return &task.RunResult{
				Output: "to be implemented",
			}
		},
	})
}

// NewBackgroundScheduler creates and initializes a new background scheduler with the given configuration.
func NewBackgroundScheduler(cfg *config.Config) *task.BackgroundScheduler {
	runner := task.NewTaskRunner(cfg)
	RegisterTasks(runner)
	return task.NewBackgroundScheduler(runner, cfg.Tasks)
}
