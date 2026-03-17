// Package lambda implements the scheduler Lambda handler.
package lambda

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/internal/task"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/scheduler"
)

// NewHandler creates and returns a Lambda handler function for task scheduling.
func NewHandler(cfg *config.Config) func(context.Context, json.RawMessage) (*task.RunResult, error) {
	return func(ctx context.Context, event json.RawMessage) (*task.RunResult, error) {
		logging.SetupLogging(cfg)

		if lc, ok := lambdacontext.FromContext(ctx); ok {
			ctx = logging.WithRequestID(ctx, lc.AwsRequestID)
		}

		runner := task.NewTaskRunner(cfg)
		if runner == nil {
			return &task.RunResult{
				Error: errors.New("task runner not initialized: missing AWSConfig"),
			}, nil
		}

		scheduler.RegisterTasks(runner, cfg)

		result := runner.Run(ctx, event)
		return result, nil
	}
}
