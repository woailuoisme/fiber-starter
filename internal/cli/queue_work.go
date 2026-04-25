package command

import (
	"os"

	"fiber-starter/internal/app/providers"
	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/services"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var queueWorkCmd = &cobra.Command{
	Use:   "queue:work",
	Short: "Run queue worker (asynq)",
	Run: func(_ *cobra.Command, _ []string) {
		runQueueWorker()
	},
}

func init() {
	rootCmd.AddCommand(queueWorkCmd)
}

func runQueueWorker() {
	container := providers.NewContainer()
	if err := container.RegisterProviders(); err != nil {
		helpers.Logger.Error("queue_worker_failed_to_init_container", zap.Error(err))
		os.Exit(1)
	}

	if err := container.Invoke(func(_ *config.Config) {}); err != nil {
		helpers.Logger.Error("queue_worker_failed_to_load_config", zap.Error(err))
		os.Exit(1)
	}

	if err := helpers.Init(); err != nil {
		helpers.Logger.Error("queue_worker_failed_to_init_logger", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		_ = helpers.Sync()
	}()

	var queue services.QueueService
	if err := container.Invoke(func(q services.QueueService) {
		queue = q
	}); err != nil {
		helpers.Logger.Error("queue_worker_failed_to_resolve_queue_service", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		_ = queue.Close()
	}()

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
			os.Exit(1)
		}
		helpers.Logger.Info("queue_worker_exited")
	case sig := <-quit:
		helpers.Logger.Info("queue_worker_shutdown_signal", zap.String("signal", sig.String()))
		_ = queue.StopWorker()
		if err := <-errCh; err != nil {
			helpers.Logger.Error("queue_worker_shutdown_error", zap.Error(err))
			os.Exit(1)
		}
		helpers.Logger.Info("queue_worker_stopped")
	}
}
