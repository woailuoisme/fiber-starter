package schedule

import (
	"lfiber/configs"
	scheduleContracts "lfiber/internal/providers/schedule/contracts"
)

// RegisterSchedule initializes and returns the schedule manager and the default scheduler.
func RegisterSchedule(cfg *configs.Config) (scheduleContracts.Manager, scheduleContracts.Scheduler, error) {
	manager := NewManager(cfg)
	return manager, manager.Scheduler(), nil
}
