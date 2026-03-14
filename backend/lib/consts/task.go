package consts

// TaskConfig defines the configuration for a scheduled task.
type TaskConfig struct {
	Name     string            `json:"name"`
	Schedule string            `json:"schedule"`
	Enabled  bool              `json:"enabled"`
	Params   map[string]string `json:"params,omitempty"`
}
