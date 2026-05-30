package schedule

import (
	"lfiber/configs"
	scheduleContracts "lfiber/internal/providers/schedule/contracts"
)

// RegisterSchedule initializes and returns the schedule manager and the default scheduler.
func RegisterSchedule(cfg *configs.Config) (scheduleContracts.Manager, scheduleContracts.Scheduler, error) {
	manager := NewManager(cfg)
	scheduler, err := manager.Scheduler()
	if err != nil {
		return nil, nil, err
	}
	return manager, scheduler, nil
}
