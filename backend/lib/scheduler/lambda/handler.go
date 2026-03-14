// Package lambda implements the scheduler Lambda handler.
package lambda

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/internal/task"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/scheduler"
)

type event struct {
	Task     string `json:"task"`
	Schedule string `json:"schedule,omitempty"`
}

// NewHandler creates and returns a Lambda handler function for task scheduling.
func NewHandler(cfg *config.Config) func(context.Context, event) *task.RunResult {
	logging.SetupLogging(cfg)

	return func(ctx context.Context, ev event) *task.RunResult {
		if lc, ok := lambdacontext.FromContext(ctx); ok {
			ctx = logging.WithRequestID(ctx, lc.AwsRequestID)
		}

		if ev.Task == "" {
			return &task.RunResult{Error: errors.New("empty task name")}
		}

		runner := task.NewTaskRunner(cfg)
		scheduler.RegisterTasks(runner)
		return runner.Run(ctx, ev.Task, ev.Schedule)
	}
}
