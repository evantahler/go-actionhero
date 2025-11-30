package config

// TasksConfig holds background task configuration
type TasksConfig struct {
	Enabled            bool
	TaskProcessors     int
	Queues             []string
	Timeout            int // Timeout in milliseconds
	StuckWorkerTimeout int // Stuck worker timeout in milliseconds
	RetryStuckJobs     bool
}

// DefaultTasksConfig returns default tasks configuration
func DefaultTasksConfig() TasksConfig {
	return TasksConfig{
		Enabled:            true,
		TaskProcessors:     1,
		Queues:             []string{"default"},
		Timeout:            10000, // 10 seconds
		StuckWorkerTimeout: 60000, // 60 seconds
		RetryStuckJobs:     false,
	}
}
