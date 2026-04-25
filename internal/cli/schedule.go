package command

import (
	"fiber-starter/internal/platform/helpers"
	schedule "fiber-starter/internal/scheduler"

	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule:run",
	Short: "Run scheduled task scheduler",
	Long:  "Start scheduled task scheduler and execute all registered scheduled tasks",
	Run: func(_ *cobra.Command, _ []string) {
		runSchedule()
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}

func runSchedule() {
	// Initialize database connection (includes config and logger initialization)
	if err := initDB(); err != nil {
		panic(err)
	}

	helpers.Info("Starting scheduled task scheduler...")

	// Create scheduled task kernel
	kernel := schedule.NewKernel()

	// Start scheduled tasks
	kernel.Start()

	// Wait for interrupt signal
	<-waitForInterrupt()

	helpers.Info("Stopping scheduled task scheduler...")
	kernel.Stop()
	helpers.Info("Scheduled task scheduler stopped")
}
