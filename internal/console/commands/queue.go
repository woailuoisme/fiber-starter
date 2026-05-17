package command

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	Jobs "fiber-starter/internal/common/jobs"
	queue "fiber-starter/internal/providers/queue"
	helpers "fiber-starter/internal/support"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// ─── Queue Command Group ──────────────────────────────────────────────────────

var queueWorkCmd = &cobra.Command{
	Use:   "queue:work",
	Short: "Run queue worker (asynq)",
	Run: func(cmd *cobra.Command, _ []string) {
		if err := runQueueWorker(cmd); err != nil {
			os.Exit(1)
		}
	},
}

var queueStatusCmd = &cobra.Command{
	Use:   "queue:status",
	Short: "Show queue health and task counts",
	Run: func(_ *cobra.Command, _ []string) {
		if err := runQueueStatus(); err != nil {
			os.Exit(1)
		}
	},
}

var queueFailedCmd = &cobra.Command{
	Use:   "queue:failed",
	Short: "List all failed (archived) queue jobs",
	Run: func(_ *cobra.Command, _ []string) {
		if err := runQueueFailed(); err != nil {
			os.Exit(1)
		}
	},
}

var queueFlushCmd = &cobra.Command{
	Use:   "queue:flush [queue]",
	Short: "Flush all failed jobs from a specific queue",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		queueName := "default"
		if len(args) > 0 {
			queueName = args[0]
		}
		if err := runQueueFlush(queueName); err != nil {
			os.Exit(1)
		}
	},
}

var queueRetryCmd = &cobra.Command{
	Use:   "queue:retry [id]",
	Short: "Retry a failed queue job by ID",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if err := runQueueRetry(args[0]); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	queueWorkCmd.Flags().String("queue", "", "The queues to work on (comma-separated)")
	queueWorkCmd.Flags().Int("concurrency", 0, "The number of concurrent workers")

	rootCmd.AddCommand(queueWorkCmd, queueStatusCmd, queueFailedCmd, queueFlushCmd, queueRetryCmd)
}

// ─── Queue Operations ─────────────────────────────────────────────────────────

func runQueueWorker(cmd *cobra.Command) error {
	queuesStr, _ := cmd.Flags().GetString("queue")
	concurrency, _ := cmd.Flags().GetInt("concurrency")

	var queues []string
	if queuesStr != "" {
		queues = append(queues, helpers.SplitAndTrim(queuesStr, ",")...)
	}

	runtime, err := buildRuntime()
	if err != nil {
		helpers.Logger.Error("queue_worker_failed_to_build_runtime", zap.Error(err))
		return err
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	if concurrency > 0 {
		queue.Drive().SetConcurrency(concurrency)
	}

	Jobs.Register()

	errCh := make(chan error, 1)
	go func() {
		errCh <- queue.RunWorker(queues...)
	}()

	helpers.Logger.Info("queue_worker_started")

	quit := waitForInterrupt()
	select {
	case err := <-errCh:
		if err != nil {
			helpers.Logger.Error("queue_worker_exited", zap.Error(err))
			return err
		}
		helpers.Logger.Info("queue_worker_exited")
	case sig := <-quit:
		helpers.Logger.Info("queue_worker_shutdown_signal", zap.String("signal", sig.String()))
		_ = queue.StopWorker()
		if err := <-errCh; err != nil {
			helpers.Logger.Error("queue_worker_shutdown_error", zap.Error(err))
			return err
		}
		helpers.Logger.Info("queue_worker_stopped")
	}

	return nil
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

	statuses, err := runtime.QueueService.InspectQueues()
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
		_, _ = fmt.Fprintf(
			w, "%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%v\n",
			status.Name, status.Pending, status.Running, status.Succeeded,
			status.Failed, status.Scheduled, status.Retry, status.Archived, status.Paused,
		)
	}
	_ = w.Flush()

	return nil
}

func runQueueFailed() error {
	runtime, err := buildRuntime()
	if err != nil {
		helpers.Logger.Error("queue_failed_build_runtime", zap.Error(err))
		return err
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	jobs, err := queue.ListFailed(1, 100)
	if err != nil {
		helpers.Logger.Error("queue_failed_list_error", zap.Error(err))
		return err
	}

	if len(jobs) == 0 {
		fmt.Println("No failed jobs found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tQUEUE\tFAILED AT\tRETRIED\tERROR")
	for _, job := range jobs {
		_, _ = fmt.Fprintf(
			w, "%s\t%s\t%s\t%d/%d\t%s\n",
			job.ID, job.Queue,
			job.FailedAt.Format("2006-01-02 15:04:05"),
			job.Retried, job.MaxRetries,
			job.Error,
		)
	}
	_ = w.Flush()

	return nil
}

func runQueueFlush(queueName string) error {
	runtime, err := buildRuntime()
	if err != nil {
		helpers.Logger.Error("queue_flush_build_runtime", zap.Error(err))
		return err
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	if err := queue.Flush(queueName); err != nil {
		helpers.Logger.Error("queue_flush_failed", zap.String("queue", queueName), zap.Error(err))
		return err
	}

	fmt.Printf("Flushed all failed jobs from queue: %s\n", queueName)
	return nil
}

func runQueueRetry(id string) error {
	runtime, err := buildRuntime()
	if err != nil {
		helpers.Logger.Error("queue_retry_build_runtime", zap.Error(err))
		return err
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	if id == "all" {
		jobs, err := queue.ListFailed(1, 1000)
		if err != nil {
			helpers.Logger.Error("queue_retry_list_failed", zap.Error(err))
			return err
		}
		for _, job := range jobs {
			if err := queue.RetryFailed(job.ID); err != nil {
				helpers.Logger.Error("queue_retry_job_failed", zap.String("id", job.ID), zap.Error(err))
			} else {
				fmt.Printf("Retrying job %s...\n", job.ID)
			}
		}
		return nil
	}

	if err := queue.RetryFailed(id); err != nil {
		helpers.Logger.Error("failed_to_retry_job", zap.String("id", id), zap.Error(err))
		return err
	}

	fmt.Printf("Job %s has been pushed back to the queue\n", id)
	return nil
}
