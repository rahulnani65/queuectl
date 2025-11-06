package pkg

import "time"

// JobState represents the lifecycle phase of a job in the queue.
// Valid values are: PENDING, PROCESSING, COMPLETED, FAILED, and DEAD.
type JobState string

const (
	StatePending    JobState = "PENDING"
	StateProcessing JobState = "PROCESSING"
	StateCompleted  JobState = "COMPLETED"
	StateFailed     JobState = "FAILED"
	StateDead       JobState = "DEAD"
)

// Job describes a unit of work to be executed by the queue.
// It stores command details, execution results, and scheduling metadata.
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

// CalculateBackoff returns an exponential backoff delay in seconds
// based on the provided base and the number of previous Attempts.
func (j *Job) CalculateBackoff(base int) int64 {
	result := 1
	for i := 0; i < j.Attempts; i++ {
		result *= base
	}
	return int64(result)
}
