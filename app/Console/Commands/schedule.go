package command

import (
	kernel "fiber-starter/app/Console/Kernel"
	helpers "fiber-starter/app/Support"

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
	if err := initDB(); err != nil {
		panic(err)
	}

	helpers.Info("Starting scheduled task scheduler...")
	scheduler := kernel.NewKernel()
	scheduler.Start()

	<-waitForInterrupt()

	helpers.Info("Stopping scheduled task scheduler...")
	scheduler.Stop()
	helpers.Info("Scheduled task scheduler stopped")
}
