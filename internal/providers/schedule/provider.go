package schedule

import (
	"fiber-starter/configs"
	scheduleContracts "fiber-starter/internal/providers/schedule/Contracts"
)

// RegisterSchedule initializes and returns the schedule manager and the default scheduler.
func RegisterSchedule(cfg *configs.Config) (scheduleContracts.Manager, scheduleContracts.Scheduler, error) {
	manager := NewManager(cfg)
	return manager, manager.Scheduler(), nil
}
