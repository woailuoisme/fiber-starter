package schedule

import (
	"fmt"
	"text/tabwriter"

	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/console/kernel"
	scheduleprovider "lfiber/internal/providers/schedule"
	"lfiber/internal/support"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{runCommand(), listCommand()}
}

func runCommand() *cobra.Command {
	return &cobra.Command{Use: "schedule:run", Short: "Run scheduled task scheduler", GroupID: "queue", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		support.Info("Starting scheduled task scheduler...")
		kernel.Schedule()
		errCh := make(chan error, 1)
		go func() { errCh <- scheduleprovider.Run() }()
		select {
		case err := <-errCh:
			return err
		case sig := <-commandutil.WaitForInterrupt():
			support.Logger.Info("schedule_shutdown_signal", zap.String("signal", sig.String()))
			_ = scheduleprovider.Stop()
			if err := <-errCh; err != nil {
				return err
			}
		}
		support.Info("Scheduled task scheduler stopped")
		return nil
	}}
}

func listCommand() *cobra.Command {
	return &cobra.Command{Use: "schedule:list", Short: "List all scheduled tasks", GroupID: "queue", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		kernel.Schedule()
		events := scheduleprovider.GetEvents()
		if len(events) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No scheduled tasks found")
			return nil
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "TASK\tEXPRESSION\tDESCRIPTION")
		for _, event := range events {
			desc := "-"
			if event.Expression == "" {
				desc = "Skipped (no expression)"
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", event.Job.TaskName(), event.Expression, desc)
		}
		return w.Flush()
	}}
}
