package schedule

import (
	scheduleContracts "fiber-starter/app/Providers/Schedule/Contracts"
	"fiber-starter/configs"
)

// RegisterSchedule initializes and returns the schedule manager and the default scheduler.
func RegisterSchedule(cfg *configs.Config) (scheduleContracts.Manager, scheduleContracts.Scheduler, error) {
	manager := NewManager(cfg)
	return manager, manager.Scheduler(), nil
}
