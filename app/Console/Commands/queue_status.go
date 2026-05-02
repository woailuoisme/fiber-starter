package command

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	providers "fiber-starter/app/Providers"
	Services "fiber-starter/app/Services"
	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

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
	container := providers.NewContainer()
	if err := container.RegisterProviders(); err != nil {
		helpers.Logger.Error("queue_status_failed_to_init_container", zap.Error(err))
		return err
	}

	if err := container.Invoke(func(_ *config.Config) {}); err != nil {
		helpers.Logger.Error("queue_status_failed_to_load_config", zap.Error(err))
		return err
	}

	if err := helpers.Init(); err != nil {
		helpers.Logger.Error("queue_status_failed_to_init_logger", zap.Error(err))
		return err
	}
	defer func() {
		_ = helpers.Sync()
	}()

	var queue Services.QueueService
	if err := container.Invoke(func(q Services.QueueService) {
		queue = q
	}); err != nil {
		helpers.Logger.Error("queue_status_failed_to_resolve_queue_service", zap.Error(err))
		return err
	}

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
