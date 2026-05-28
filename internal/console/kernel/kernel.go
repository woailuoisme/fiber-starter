package kernel

import (
	"lfiber/internal/features/monitoring"
	schedule "lfiber/internal/providers/schedule"
)

// Schedule defines the application's command schedule.
// This is where you register all your background tasks and their frequencies.
func Schedule() {
	monitoring.Schedule(schedule.DefaultRegistry())
}
