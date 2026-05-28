package user

import (
	userJobs "lfiber/internal/features/user/jobs"
	queue "lfiber/internal/providers/queue"
)

// RegisterJobs registers user feature queue jobs.
func RegisterJobs(registry queue.Registry) {
	registry.Job(userJobs.NewWelcomeEmailJob("", ""))
	registry.Job(NewUserImportJob(""))
	registry.Job(NewUserExportJob(""))
}
