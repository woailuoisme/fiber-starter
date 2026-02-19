package schedule

import (
	"fiber-starter/internal/platform/helpers"
	"go.uber.org/zap"
)

// ExampleTask Example task
func ExampleTask() {
	helpers.Info("Executing example task")
	// Write your task logic here
}

// CleanupTask Cleanup task example
func CleanupTask() {
	helpers.Info("Executing cleanup task")
	// Clean up temporary files, expired data, etc.
}

// SendEmailTask Send email task example
func SendEmailTask() {
	helpers.Info("Executing send email task")
	// Send scheduled emails
}

// BackupDatabaseTask Database backup task example
func BackupDatabaseTask() {
	helpers.Info("Executing database backup task")
	// Backup database
}

// GenerateReportTask Generate report task example
func GenerateReportTask() {
	helpers.Info("Executing generate report task", zap.String("type", "daily"))
	// Generate daily reports, weekly reports, etc.
}
