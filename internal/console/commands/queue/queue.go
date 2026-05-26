package queue

import (
	"fmt"
	"sort"
	"text/tabwriter"

	jobs "lfiber/internal/common/jobs"
	"lfiber/internal/console/commands/commandutil"
	queueprovider "lfiber/internal/providers/queue"
	"lfiber/internal/support"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Commands() []*cobra.Command {
	work := workCommand()
	return []*cobra.Command{work, statusCommand(), failedCommand(), flushCommand(), retryCommand()}
}

func workCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "queue:work", Short: "Run queue worker", GroupID: "queue", Args: cobra.NoArgs, RunE: runWork}
	cmd.Flags().String("queue", "", "queues to work on (comma-separated)")
	cmd.Flags().Int("concurrency", 0, "number of concurrent workers")
	return cmd
}

func statusCommand() *cobra.Command {
	return &cobra.Command{Use: "queue:status", Short: "Show queue health and task counts", GroupID: "queue", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		if rt.QueueService == nil {
			return fmt.Errorf("queue service is not available")
		}
		statuses, err := rt.QueueService.InspectQueues()
		if err != nil {
			return err
		}
		if len(statuses) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No queues found")
			return nil
		}
		sort.Slice(statuses, func(i, j int) bool { return statuses[i].Name < statuses[j].Name })
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "QUEUE\tPENDING\tRUNNING\tSUCCEEDED\tFAILED\tSCHEDULED\tRETRY\tARCHIVED\tPAUSED")
		for _, status := range statuses {
			_, _ = fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%v\n", status.Name, status.Pending, status.Running, status.Succeeded, status.Failed, status.Scheduled, status.Retry, status.Archived, status.Paused)
		}
		return w.Flush()
	}}
}

func failedCommand() *cobra.Command {
	return &cobra.Command{Use: "queue:failed", Short: "List failed queue jobs", GroupID: "queue", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		jobs, err := queueprovider.ListFailed(1, 100)
		if err != nil {
			return err
		}
		if len(jobs) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No failed jobs found")
			return nil
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tQUEUE\tFAILED AT\tRETRIED\tERROR")
		for _, job := range jobs {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d/%d\t%s\n", job.ID, job.Queue, job.FailedAt.Format("2006-01-02 15:04:05"), job.Retried, job.MaxRetries, job.Error)
		}
		return w.Flush()
	}}
}

func flushCommand() *cobra.Command {
	return &cobra.Command{Use: "queue:flush [queue]", Short: "Flush failed jobs from a queue", GroupID: "queue", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		queueName := "default"
		if len(args) > 0 {
			queueName = args[0]
		}
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		if err := queueprovider.Flush(queueName); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Flushed all failed jobs from queue: %s\n", queueName)
		return nil
	}}
}

func retryCommand() *cobra.Command {
	return &cobra.Command{Use: "queue:retry <id|all>", Short: "Retry a failed queue job", GroupID: "queue", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		id := args[0]
		if id == "all" {
			jobs, err := queueprovider.ListFailed(1, 1000)
			if err != nil {
				return err
			}
			for _, job := range jobs {
				if err := queueprovider.RetryFailed(job.ID); err != nil {
					support.Logger.Error("queue_retry_job_failed", zap.String("id", job.ID), zap.Error(err))
					continue
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Retrying job %s...\n", job.ID)
			}
			return nil
		}
		if err := queueprovider.RetryFailed(id); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Job %s has been pushed back to the queue\n", id)
		return nil
	}}
}

func runWork(cmd *cobra.Command, _ []string) error {
	queuesStr, _ := cmd.Flags().GetString("queue")
	concurrency, _ := cmd.Flags().GetInt("concurrency")
	queues := []string{}
	if queuesStr != "" {
		queues = append(queues, support.SplitAndTrim(queuesStr, ",")...)
	}

	rt, err := commandutil.BuildRuntime()
	if err != nil {
		return err
	}
	defer func() { _ = commandutil.CloseRuntime(rt) }()
	if concurrency > 0 {
		queueprovider.Drive().SetConcurrency(concurrency)
	}
	jobs.Register()

	errCh := make(chan error, 1)
	go func() { errCh <- queueprovider.RunWorker(queues...) }()
	support.Logger.Info("queue_worker_started")

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case sig := <-commandutil.WaitForInterrupt():
		support.Logger.Info("queue_worker_shutdown_signal", zap.String("signal", sig.String()))
		_ = queueprovider.StopWorker()
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}
