package bootstrap

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"fiber-starter/internal/app/providers"
	"fiber-starter/internal/config"
	database "fiber-starter/internal/db"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/services"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// runHTTPServer 负责启动、等待退出信号并清理资源。
func runHTTPServer(app *fiber.App, container *providers.Container, cfg *config.Config) {
	port := ":" + cfg.App.Port
	baseURL := strings.TrimRight(cfg.App.URL, "/")
	docsURL := baseURL + "/docs"

	helpers.Info(
		"server_listening",
		zap.String("port", cfg.App.Port),
		zap.String("app_url", baseURL),
		zap.String("docs_url", docsURL),
	)

	listenErr := make(chan error, 1)
	go func() {
		listenErr <- app.Listen(port, fiber.ListenConfig{
			EnablePrefork:         cfg.App.Fiber.Prefork,
			DisableStartupMessage: true,
		})
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-listenErr:
		if err != nil {
			helpers.Fatal("server_failed_to_start", zap.Error(err))
		}
		return
	case <-sigCh:
		helpers.Info("shutdown_signal_received")
	}

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- app.Shutdown()
	}()

	select {
	case err := <-shutdownDone:
		if err != nil {
			helpers.Warn("server_shutdown_failed", zap.Error(err))
		}
	case <-time.After(15 * time.Second):
		helpers.Warn("server_shutdown_timed_out")
	}

	_ = container.Invoke(func(conn *database.Connection, cache helpers.CacheService, queue services.QueueService, storage *services.StorageService) {
		_ = storage.Close()
		_ = queue.Close()
		_ = cache.Close()
		_ = conn.Close()
	})
	_ = helpers.Sync()
}
