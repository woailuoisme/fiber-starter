package kernel

import (
	monitoringJobs "lfiber/internal/features/monitoring/jobs"
	schedule "lfiber/internal/providers/schedule"
)

// Schedule defines the application's command schedule.
// This is where you register all your background tasks and their frequencies.
func Schedule() {
	// Example: Cleanup temporary files every minute
	schedule.Job(monitoringJobs.NewCleanupTempFilesJob("/tmp")).EveryMinute()

	// Example: Run a custom function every 5 minutes
	schedule.Call(func() error {
		// Log something or perform a task
		return nil
	}).EveryFiveMinutes().Name("log_cleanup")

	// Example: Run a console command daily
	schedule.Command("cache:clear").Daily().Name("cache_refresh")
}
