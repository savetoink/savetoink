// Package scheduler provides configuration and initialization for background task scheduling.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/shaftoe/savetoink/backend/lib/backup"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/internal/task"
)

// RegisterTasks registers all available tasks with the given task runner.
func RegisterTasks(runner *task.TaskRunner) {
	runner.Register(task.Task{
		Name: "backup",
		Run: func(ctx context.Context) *task.RunResult {
			cfg := runner.GetConfig()

			if cfg.BackupBucketName == "" {
				return &task.RunResult{
					Error: fmt.Errorf("backup bucket name is not configured"),
				}
			}

			awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
			if err != nil {
				return &task.RunResult{
					Error: fmt.Errorf("failed to load AWS config: %w", err),
				}
			}

			dynamoDBClient := dynamodb.NewFromConfig(awsConfig)
			s3Client := s3.NewFromConfig(awsConfig)

			backupManager := backup.NewBackupManager(dynamoDBClient, s3Client, cfg.BackupBucketName, slog.Default())

			tables := []string{
				cfg.ArticlesTable,
				cfg.ArticleTagsTable,
				cfg.UserProfileTable,
				cfg.SendsTable,
			}

			results := backupManager.BackupAllTables(ctx, tables)

			var output strings.Builder
			var errors []string

			for _, result := range results {
				if result.Error != nil {
					errors = append(errors, fmt.Sprintf("%s: %v", result.TableName, result.Error))
				} else {
					output.WriteString(fmt.Sprintf("%s: %d items backed up to %s\n", result.TableName, result.ItemsCount, result.Key))
				}
			}

			if len(errors) > 0 {
				return &task.RunResult{
					Output: output.String(),
					Error:  fmt.Errorf("backup failed for tables: %s", strings.Join(errors, "; ")),
				}
			}

			return &task.RunResult{
				Output: output.String(),
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
