// Package lambda implements the scheduler Lambda handler.
package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/internal/task"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/scheduler"
)

type event struct {
	Task     string                    `json:"task"`
	Schedule string                    `json:"schedule,omitempty"`
	Params   map[string]map[string]any `json:"params,omitempty"`
}

// NewHandler creates and returns a Lambda handler function for task scheduling.
func NewHandler(cfg *config.Config) func(context.Context, event) (*task.RunResult, error) {
	return func(ctx context.Context, ev event) (*task.RunResult, error) {
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

		params, parseErr := parseTaskParams(ev.Params)
		if parseErr != nil {
			return &task.RunResult{Error: parseErr}, parseErr
		}

		result := runner.Run(ctx, ev.Task, ev.Schedule, params)
		return result, nil
	}
}

func parseTaskParams(params map[string]map[string]any) (map[string]task.ParamValue, error) {
	result := make(map[string]task.ParamValue)

	if params == nil {
		return result, nil
	}

	for k, v := range params {
		if len(v) == 1 {
			for _, val := range v {
				switch val := val.(type) {
				case string:
					result[k] = task.StringParam(val)
				default:
					return nil, fmt.Errorf("unsupported parameter type for %s: %T", k, val)
				}
			}
		}
	}

	return result, nil
}
