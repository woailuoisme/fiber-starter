package schedule

import (
	"errors"

	queueContracts "fiber-starter/app/Providers/Queue/Contracts"
	"fiber-starter/app/Providers/Schedule/Contracts"
	"fiber-starter/app/Support/appctx"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// service returns the default scheduler service instance from the container.
func service() Contracts.Scheduler {
	if app := appctx.App(); app != nil {
		return app.ScheduleServiceValue()
	}
	return nil
}

// manager returns the scheduler manager instance from the container.
func manager() Contracts.Manager {
	if app := appctx.App(); app != nil {
		return app.ScheduleManagerValue()
	}
	return nil
}

// Job registers a job to be scheduled using the default scheduler
func Job(job queueContracts.Job) *Contracts.Event {
	if s := service(); s != nil {
		return s.Job(job)
	}
	return nil
}

// Call registers a function to be scheduled using the default scheduler
func Call(fn func() error) *Contracts.Event {
	if s := service(); s != nil {
		return s.Call(fn)
	}
	return nil
}

// Command registers a console command to be scheduled using the default scheduler
func Command(command string, args ...string) *Contracts.Event {
	if s := service(); s != nil {
		return s.Command(command, args...)
	}
	return nil
}

// GetEvents returns all registered scheduled events
func GetEvents() []*Contracts.Event {
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
