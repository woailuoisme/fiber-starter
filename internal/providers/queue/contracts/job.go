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

// JobMeta provides reusable queue metadata for jobs.
type JobMeta struct {
	taskName  string
	queueName string
}

// NewJobMeta creates queue metadata for a job.
func NewJobMeta(taskName, queueName string) JobMeta {
	if queueName == "" {
		queueName = "default"
	}
	return JobMeta{taskName: taskName, queueName: queueName}
}

// TaskName returns the unique identifier for this job type.
func (m JobMeta) TaskName() string {
	return m.taskName
}

// QueueName returns the queue this job should run on.
func (m JobMeta) QueueName() string {
	if m.queueName == "" {
		return "default"
	}
	return m.queueName
}
