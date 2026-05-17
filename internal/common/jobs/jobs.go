package jobs

import (
	monitoringJobs "fiber-starter/internal/features/monitoring/jobs"
	userJobs "fiber-starter/internal/features/user/jobs"
	queue "fiber-starter/internal/providers/queue"
)

// Register registers all application jobs with the queue system.
// This is typically called during application bootstrap or when starting a queue worker.
func Register() {
	queue.Register(&userJobs.WelcomeEmailJob{})
	queue.Register(&monitoringJobs.CleanupTempFilesJob{})
}
