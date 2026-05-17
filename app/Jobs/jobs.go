package Jobs

import (
	queue "fiber-starter/app/Providers/Queue"
)

// Register registers all application jobs with the queue system.
// This is typically called during application bootstrap or when starting a queue worker.
func Register() {
	queue.Register(&WelcomeEmailJob{})
	queue.Register(&CleanupTempFilesJob{})
}
