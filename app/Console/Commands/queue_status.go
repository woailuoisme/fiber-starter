package command

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	helpers "fiber-starter/app/Support"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var queueStatusCmd = &cobra.Command{
	Use:   "queue:status",
	Short: "Show queue health and task counts",
	Run: func(_ *cobra.Command, _ []string) {
		if err := runQueueStatus(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(queueStatusCmd)
}

func runQueueStatus() error {
	runtime, err := buildRuntime()
	if err != nil {
		helpers.Logger.Error("queue_status_failed_to_build_runtime", zap.Error(err))
		return err
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	queue := runtime.QueueService

	statuses, err := queue.InspectQueues()
	if err != nil {
		helpers.Logger.Error("queue_status_failed", zap.Error(err))
		return err
	}

	if len(statuses) == 0 {
		fmt.Println("No queues found")
		return nil
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "QUEUE\tPENDING\tRUNNING\tSUCCEEDED\tFAILED\tSCHEDULED\tRETRY\tARCHIVED\tPAUSED")
	for _, status := range statuses {
		_, _ = fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%t\n",
			status.Name,
			status.Pending,
			status.Running,
			status.Succeeded,
			status.Failed,
			status.Scheduled,
			status.Retry,
			status.Archived,
			status.Paused,
		)
	}
	_ = w.Flush()

	return nil
}
