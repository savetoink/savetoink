// Package lambda implements the scheduler Lambda handler.
package lambda

import (
	"context"

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
func NewHandler(cfg *config.Config) func(context.Context, event) (*task.RunResult, error) {
	return func(ctx context.Context, ev event) (*task.RunResult, error) {
		logging.SetupLogging(cfg)

		if lc, ok := lambdacontext.FromContext(ctx); ok {
			ctx = logging.WithRequestID(ctx, lc.AwsRequestID)
		}

		runner := task.NewTaskRunner(cfg)
		scheduler.RegisterTasks(runner)
		result := runner.Run(ctx, ev.Task, ev.Schedule)
		return result, nil
	}
}
