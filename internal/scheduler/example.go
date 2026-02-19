// Package schedule defines application scheduled tasks
package schedule

import (
	"time"

	"fiber-starter/internal/platform/helpers"

	"go.uber.org/zap"
)

// TestScheduleTask Test scheduled task (runs every 10 seconds)
func TestScheduleTask() {
	now := time.Now().Format("2006-01-02 15:04:05")
	helpers.Info("🎯 Scheduled task test", zap.String("time", now))
}

// If you want to quickly test if scheduled tasks work, add in kernel.go's Schedule() method:
//
// // Test: run every 10 seconds
// k.Cron("*/10 * * * * *", TestScheduleTask)
//
// Then run: make schedule
// You will see log output every 10 seconds
