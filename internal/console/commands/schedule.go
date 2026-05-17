package command

import (
	"fmt"
	"os"
	"text/tabwriter"

	kernel "fiber-starter/internal/console/kernel"
	schedule "fiber-starter/internal/providers/schedule"
	helpers "fiber-starter/internal/support"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// ─── Schedule Command Group ───────────────────────────────────────────────────

var scheduleCmd = &cobra.Command{
	Use:   "schedule:run",
	Short: "Run scheduled task scheduler",
	Long:  "Start scheduled task scheduler and execute all registered scheduled tasks",
	Run:   func(_ *cobra.Command, _ []string) { runSchedule() },
}

var scheduleListCmd = &cobra.Command{
	Use:   "schedule:list",
	Short: "List all scheduled tasks",
	Run: func(_ *cobra.Command, _ []string) {
		if err := runScheduleList(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd, scheduleListCmd)
}

// ─── Schedule Operations ──────────────────────────────────────────────────────

func runSchedule() {
	runtime, err := buildRuntime()
	if err != nil {
		helpers.Logger.Error("scheduler_failed_to_build_runtime", zap.Error(err))
		return
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	helpers.Info("Starting scheduled task scheduler...")

	kernel.Schedule()

	go func() {
		if err := schedule.Run(); err != nil {
			helpers.Error("Scheduler runtime error", zap.Error(err))
		}
	}()

	<-waitForInterrupt()

	helpers.Info("Stopping scheduled task scheduler...")
	_ = schedule.Stop()
	helpers.Info("Scheduled task scheduler stopped")
}

func runScheduleList() error {
	runtime, err := buildRuntime()
	if err != nil {
		helpers.Logger.Error("schedule_list_failed_to_build_runtime", zap.Error(err))
		return err
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	kernel.Schedule()

	events := schedule.GetEvents()
	if len(events) == 0 {
		fmt.Println("No scheduled tasks found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "TASK\tEXPRESSION\tDESCRIPTION")
	for _, event := range events {
		desc := "-"
		if event.Expression == "" {
			desc = "Skipped (no expression)"
		}
		_, _ = fmt.Fprintf(
			w, "%s\t%s\t%s\n",
			event.Job.TaskName(),
			event.Expression,
			desc,
		)
	}
	_ = w.Flush()

	return nil
}
