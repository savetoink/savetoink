package task

import "github.com/shaftoe/savetoink/backend/lib/consts"

// SchedulerConfig defines the configuration for the background scheduler.
type SchedulerConfig struct {
	Tasks []consts.TaskConfig
}
