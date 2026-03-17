package consts

// TaskConfig defines the configuration for a scheduled task.
type TaskConfig struct {
	Task string `json:"task"`

	// Backup
	Schedule string `json:"schedule"`

	// Restore
	BackupName string `json:"backup_name,omitempty"`
	Overwrite  bool   `json:"overwrite,omitempty"`
}
