// Package scheduler provides configuration and initialization for background task scheduling.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/internal/task"
	repo "github.com/shaftoe/savetoink/backend/lib/repository/dynamodb"
)

// RegisterTasks registers all available tasks with the given task runner and configuration.
func RegisterTasks(runner *task.TaskRunner, cfg *config.Config) {
	runner.Register(task.Task{
		Name: "backup",
		Run: func(ctx context.Context) *task.RunResult {
			tables := []string{
				cfg.ArticlesTable,
				cfg.UserProfileTable,
				cfg.SendsTable,
			}
			results := runner.BkpRepo.BackupAllTables(ctx, tables)
			out, err := formatBackupResults(results)

			return &task.RunResult{
				Error:  err,
				Output: out,
			}
		},
	})
}

func formatBackupResults(results []*repo.BackupResult) (string, error) {
	var out []string
	var errs error
	successCount := 0
	failCount := 0

	for _, result := range results {
		if result.Error != nil {
			failCount++
			errs = errors.Join(errs, fmt.Errorf("failed to backup table %s: %w", result.TableName, result.Error))
		} else {
			successCount++
			out = append(out, fmt.Sprintf("backed up table %s: %d items -> %s latency: %v",
				result.TableName, result.ItemsCount, result.Key, result.Latency))
		}
	}

	out = append(out, fmt.Sprintf("backup summary: %d succeeded, %d failed", successCount, failCount))

	return strings.Join(out, ", "), errs
}

// NewBackgroundScheduler creates and initializes a new background scheduler
// with the given configuration.
// NOTICE: if the task runner is not initialized, it returns nil.
func NewBackgroundScheduler(cfg *config.Config) *task.BackgroundScheduler {
	if len(cfg.Tasks) == 0 {
		return nil
	}

	runner := task.NewTaskRunner(cfg)
	if runner == nil {
		return nil
	}

	RegisterTasks(runner, cfg)
	return task.NewBackgroundScheduler(runner, cfg.Tasks)
}
