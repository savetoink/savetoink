package consts

// TaskConfig defines the configuration for a scheduled task.
type TaskConfig struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"` // cron expression
	Enabled  bool   `json:"enabled"`
}
