package contracts

import (
	"context"
)

// Job represents a background task that can be queued
type Job interface {
	// Handle executes the job logic
	Handle(ctx context.Context) error

	// TaskName returns the unique identifier for this job type
	TaskName() string

	// QueueName returns the name of the queue this job should run on
	QueueName() string
}
