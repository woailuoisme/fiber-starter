package schedule

import (
	"errors"

	queueContracts "lfiber/internal/providers/queue/contracts"
	"lfiber/internal/providers/schedule/contracts"
	"lfiber/internal/support/appctx"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// Registry is the feature-facing schedule registration surface.
type Registry interface {
	Job(job queueContracts.Job) *contracts.Event
	Call(fn func() error) *contracts.Event
	Command(command string, args ...string) *contracts.Event
}

type defaultRegistry struct{}

// DefaultRegistry returns a registry backed by the default scheduler service.
func DefaultRegistry() Registry {
	return defaultRegistry{}
}

func (defaultRegistry) Job(job queueContracts.Job) *contracts.Event {
	return Job(job)
}

func (defaultRegistry) Call(fn func() error) *contracts.Event {
	return Call(fn)
}

func (defaultRegistry) Command(command string, args ...string) *contracts.Event {
	return Command(command, args...)
}

// service returns the default scheduler service instance from the container.
func service() contracts.Scheduler {
	if app := appctx.App(); app != nil {
		return app.ScheduleServiceValue()
	}
	return nil
}

// manager returns the scheduler manager instance from the container.
func manager() contracts.Manager {
	if app := appctx.App(); app != nil {
		return app.ScheduleManagerValue()
	}
	return nil
}

// Job registers a job to be scheduled using the default scheduler
func Job(job queueContracts.Job) *contracts.Event {
	if s := service(); s != nil {
		return s.Job(job)
	}
	return nil
}

// Call registers a function to be scheduled using the default scheduler
func Call(fn func() error) *contracts.Event {
	if s := service(); s != nil {
		return s.Call(fn)
	}
	return nil
}

// Command registers a console command to be scheduled using the default scheduler
func Command(command string, args ...string) *contracts.Event {
	if s := service(); s != nil {
		return s.Command(command, args...)
	}
	return nil
}

// GetEvents returns all registered scheduled events
func GetEvents() []*contracts.Event {
	if s := service(); s != nil {
		return s.GetEvents()
	}
	return nil
}

// Run starts the default scheduler process
func Run() error {
	if s := service(); s != nil {
		return s.Run()
	}
	return ErrContainerNotInitialized
}

// Stop gracefully stops the default scheduler process
func Stop() error {
	if m := manager(); m != nil {
		return m.Close()
	}
	return ErrContainerNotInitialized
}
