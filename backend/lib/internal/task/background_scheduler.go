// Package task provides background task scheduling and execution capabilities.
package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
)

// BackgroundScheduler manages the execution of scheduled tasks using cron expressions.
type BackgroundScheduler struct {
	cron    *cron.Cron
	runner  *TaskRunner
	configs []consts.TaskConfig
}

// NewBackgroundScheduler creates a new background scheduler with the given task runner and configurations.
func NewBackgroundScheduler(runner *TaskRunner, configs []consts.TaskConfig) *BackgroundScheduler {
	return &BackgroundScheduler{
		cron:    cron.New(cron.WithSeconds()),
		runner:  runner,
		configs: configs,
	}
}

// Start begins the background scheduler, registering all scheduled tasks with their cron schedules.
func (s *BackgroundScheduler) Start(ctx context.Context) error {
	tasksEnabled := make(map[string]struct{})

	for _, cfg := range s.configs {
		if _, ok := tasksEnabled[cfg.Task]; ok {
			continue
		}

		eventDataJSON, err := json.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("encoding event for task %q: %w", cfg.Task, err)
		}

		if _, cronErr := s.cron.AddFunc(cfg.Schedule, func() {
			_ = s.runner.Run(ctx, eventDataJSON)
		}); cronErr != nil {
			return fmt.Errorf("scheduling task %q: %w", cfg.Task, cronErr)
		}
		tasksEnabled[cfg.Task] = struct{}{}
	}

	if len(tasksEnabled) == 0 {
		return nil
	}
	s.cron.Start()

	logging.LogSchedulerStarted(ctx, tasksEnabled)

	return nil
}

// Stop gracefully stops the background scheduler, waiting for running tasks to complete.
func (s *BackgroundScheduler) Stop() {
	ctx := s.cron.Stop() // drains running jobs
	<-ctx.Done()
	slog.Info("background scheduler shutdown successfully")
}
