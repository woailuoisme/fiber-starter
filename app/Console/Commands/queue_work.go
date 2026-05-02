package command

import (
	"os"

	providers "fiber-starter/app/Providers"
	Services "fiber-starter/app/Services"
	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

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
	container := providers.NewContainer()
	if err := container.RegisterProviders(); err != nil {
		helpers.Logger.Error("queue_worker_failed_to_init_container", zap.Error(err))
		return err
	}

	if err := container.Invoke(func(_ *config.Config) {}); err != nil {
		helpers.Logger.Error("queue_worker_failed_to_load_config", zap.Error(err))
		return err
	}

	if err := helpers.Init(); err != nil {
		helpers.Logger.Error("queue_worker_failed_to_init_logger", zap.Error(err))
		return err
	}
	defer func() {
		_ = helpers.Sync()
	}()

	var queue Services.QueueService
	if err := container.Invoke(func(q Services.QueueService) {
		queue = q
	}); err != nil {
		helpers.Logger.Error("queue_worker_failed_to_resolve_queue_service", zap.Error(err))
		return err
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
