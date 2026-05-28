package monitoring

import (
	monitoringJobs "lfiber/internal/features/monitoring/jobs"
	queue "lfiber/internal/providers/queue"
	schedule "lfiber/internal/providers/schedule"
)

// RegisterJobs registers monitoring feature queue jobs.
func RegisterJobs(registry queue.Registry) {
	registry.Job(monitoringJobs.NewCleanupTempFilesJob(""))
}

// Schedule registers monitoring feature scheduled tasks.
func Schedule(registry schedule.Registry) {
	registry.Job(monitoringJobs.NewCleanupTempFilesJob("/tmp")).EveryMinute()
}
