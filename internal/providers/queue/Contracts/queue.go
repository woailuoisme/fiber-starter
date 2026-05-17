package contracts

import (
	"time"
)

// QueueStatus represents the current state of a queue
type QueueStatus struct {
	Name      string
	Pending   int
	Running   int
	Succeeded int
	Failed    int
	Scheduled int
	Retry     int
	Archived  int
	Paused    bool
}

// FailedJob represents a task that failed and was archived
type FailedJob struct {
	ID         string
	Payload    string
	Queue      string
	FailedAt   time.Time
	Error      string
	Retried    int
	MaxRetries int
}

// Queue defines the contract for queue drivers
type Queue interface {
	// Push adds a job to the queue for immediate execution
	Push(job Job) error

	// Size returns the number of jobs in a specific queue
	Size(queue ...string) (int64, error)

	// PushOn adds a job to a specific queue
	PushOn(queue string, job Job) error

	// Later adds a job to the queue to be executed after a delay
	Later(delay time.Duration, job Job) error

	// LaterOn adds a job to a specific queue to be executed after a delay
	LaterOn(queue string, delay time.Duration, job Job) error

	// Bulk pushes a batch of jobs onto the queue
	Bulk(jobs []Job, queue ...string) error

	// ProcessAt adds a job to the queue to be executed at a specific time
	ProcessAt(at time.Time, job Job) error

	// Register registers a job handler for the worker
	Register(job Job)

	// StartWorker starts the background worker process (non-blocking)
	StartWorker(queue ...string) error

	// RunWorker starts the background worker process (blocking)
	RunWorker(queue ...string) error

	// StopWorker gracefully stops the background worker process
	StopWorker() error

	// InspectQueues returns status information for all queues
	InspectQueues() ([]QueueStatus, error)

	// ListFailed returns a list of failed (archived) jobs
	ListFailed(page, pageSize int) ([]FailedJob, error)

	// RetryFailed moves a failed job back to the pending state
	RetryFailed(id string) error

	// DeleteFailed permanently removes a failed job
	DeleteFailed(id string) error

	// Flush removes all jobs from a specific queue and state
	Flush(queue string) error

	// HealthCheck verifies the queue connection is alive
	HealthCheck() error

	// Close closes the queue connection
	Close() error

	// SetConcurrency sets the number of concurrent workers
	SetConcurrency(n int)

	// GetConcurrency returns the current concurrency setting
	GetConcurrency() int
}
