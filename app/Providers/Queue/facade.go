package queue

import (
	"errors"
	"time"

	"fiber-starter/app/Providers/Queue/Contracts"
	"fiber-starter/app/Support/appctx"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// service returns the default queue service instance from the container.
func service() Contracts.Queue {
	if app := appctx.App(); app != nil {
		return app.QueueServiceValue()
	}
	return nil
}

// manager returns the queue manager instance from the container.
func manager() Contracts.Manager {
	if app := appctx.App(); app != nil {
		return app.QueueManagerValue()
	}
	return nil
}

// Drive returns a specific queue driver instance
func Drive(name ...string) Contracts.Queue {
	if m := manager(); m != nil {
		return m.Drive(name...)
	}
	return nil
}

// Push adds a job to the default queue for immediate execution
func Push(job Contracts.Job) error {
	if s := service(); s != nil {
		return s.Push(job)
	}
	return ErrContainerNotInitialized
}

// Size returns the number of jobs in a specific queue
func Size(queueName ...string) (int64, error) {
	if s := service(); s != nil {
		return s.Size(queueName...)
	}
	return 0, ErrContainerNotInitialized
}

// PushOn adds a job to a specific queue
func PushOn(queueName string, job Contracts.Job) error {
	if s := service(); s != nil {
		return s.PushOn(queueName, job)
	}
	return ErrContainerNotInitialized
}

// Later adds a job to the default queue to be executed after a delay
func Later(delay time.Duration, job Contracts.Job) error {
	if s := service(); s != nil {
		return s.Later(delay, job)
	}
	return ErrContainerNotInitialized
}

// LaterOn adds a job to a specific queue to be executed after a delay
func LaterOn(queueName string, delay time.Duration, job Contracts.Job) error {
	if s := service(); s != nil {
		return s.LaterOn(queueName, delay, job)
	}
	return ErrContainerNotInitialized
}

// Bulk pushes a batch of jobs onto the queue
func Bulk(jobs []Contracts.Job, queueName ...string) error {
	if s := service(); s != nil {
		return s.Bulk(jobs, queueName...)
	}
	return ErrContainerNotInitialized
}

// ProcessAt adds a job to the default queue to be executed at a specific time
func ProcessAt(at time.Time, job Contracts.Job) error {
	if s := service(); s != nil {
		return s.ProcessAt(at, job)
	}
	return ErrContainerNotInitialized
}

// Register registers a job handler for the worker in the default queue
func Register(job Contracts.Job) {
	if s := service(); s != nil {
		s.Register(job)
	}
}

// StartWorker starts the background worker process using the default driver (non-blocking)
func StartWorker(queues ...string) error {
	if s := service(); s != nil {
		return s.StartWorker(queues...)
	}
	return ErrContainerNotInitialized
}

// RunWorker starts the background worker process using the default driver (blocking)
func RunWorker(queues ...string) error {
	if s := service(); s != nil {
		return s.RunWorker(queues...)
	}
	return ErrContainerNotInitialized
}

// StopWorker gracefully stops the background worker process using the default driver
func StopWorker() error {
	if s := service(); s != nil {
		return s.StopWorker()
	}
	return ErrContainerNotInitialized
}

// InspectQueues returns status information for all queues
func InspectQueues() ([]Contracts.QueueStatus, error) {
	if s := service(); s != nil {
		return s.InspectQueues()
	}
	return nil, ErrContainerNotInitialized
}

// ListFailed returns a list of failed (archived) jobs
func ListFailed(page, pageSize int) ([]Contracts.FailedJob, error) {
	if s := service(); s != nil {
		return s.ListFailed(page, pageSize)
	}
	return nil, ErrContainerNotInitialized
}

// RetryFailed moves a failed job back to the pending state
func RetryFailed(id string) error {
	if s := service(); s != nil {
		return s.RetryFailed(id)
	}
	return ErrContainerNotInitialized
}

// DeleteFailed permanently removes a failed job
func DeleteFailed(id string) error {
	if s := service(); s != nil {
		return s.DeleteFailed(id)
	}
	return ErrContainerNotInitialized
}

// Flush removes all archived jobs from a specific queue
func Flush(queueName string) error {
	if s := service(); s != nil {
		return s.Flush(queueName)
	}
	return ErrContainerNotInitialized
}

// SetConcurrency sets the number of concurrent workers
func SetConcurrency(n int) {
	if s := service(); s != nil {
		s.SetConcurrency(n)
	}
}

// GetConcurrency returns the current concurrency setting
func GetConcurrency() int {
	if s := service(); s != nil {
		return s.GetConcurrency()
	}
	return 0
}

// Close closes the default queue connection
func Close() error {
	if m := manager(); m != nil {
		return m.Close()
	}
	return ErrContainerNotInitialized
}
