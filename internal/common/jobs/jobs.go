package jobs

import (
	monitoringJobs "lfiber/internal/features/monitoring/jobs"
	"lfiber/internal/features/user"
	userJobs "lfiber/internal/features/user/jobs"
	queue "lfiber/internal/providers/queue"
)

// Register registers all application jobs with the queue system.
// This is typically called during application bootstrap or when starting a queue worker.
func Register() {
	queue.Register(&userJobs.WelcomeEmailJob{})
	queue.Register(&monitoringJobs.CleanupTempFilesJob{})
	queue.Register(&user.UserImportJob{})
	queue.Register(&user.UserExportJob{})
}
