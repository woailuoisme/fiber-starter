package jobs

import (
	"lfiber/internal/features/monitoring"
	"lfiber/internal/features/user"
	queue "lfiber/internal/providers/queue"
	"lfiber/pkg/medialibrary"
)

// Register registers all application jobs with the queue system.
// This is typically called during application bootstrap or when starting a queue worker.
func Register() {
	registry := queue.DefaultRegistry()
	user.RegisterJobs(registry)
	monitoring.RegisterJobs(registry)
	medialibrary.RegisterJobs(registry)
}
