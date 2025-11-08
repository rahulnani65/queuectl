package pkg

import "time"

// JobState represents job state
type JobState string

const (
	StatePending    JobState = "PENDING"
	StateProcessing JobState = "PROCESSING"
	StateCompleted  JobState = "COMPLETED"
	StateFailed     JobState = "FAILED"
	StateDead       JobState = "DEAD"
)

// Job represents a queued job
type Job struct {
	ID           string
	Command      string
	State        JobState
	Attempts     int
	MaxRetries   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ScheduledAt  *time.Time
	ErrorMessage string
	ExitCode     *int
	Output       string
}

// CalculateBackoff calculates exponential backoff delay
func (j *Job) CalculateBackoff(base int) int64 {
	result := 1
	for i := 0; i < j.Attempts; i++ {
		result *= base
	}
	return int64(result)
}
