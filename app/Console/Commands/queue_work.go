package command

import (
	"os"

	helpers "fiber-starter/app/Support"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var queueWorkCmd = &cobra.Command{
	Use:   "queue:work",
	Short: "Run queue worker (asynq)",
	Run: func(_ *cobra.Command, _ []string) {
		if err := runQueueWorker(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(queueWorkCmd)
}

func runQueueWorker() error {
	runtime, err := buildRuntime()
	if err != nil {
		helpers.Logger.Error("queue_worker_failed_to_build_runtime", zap.Error(err))
		return err
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	queue := runtime.QueueService

	errCh := make(chan error, 1)
	go func() {
		errCh <- queue.RunWorker()
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
