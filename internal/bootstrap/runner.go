package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"lfiber/configs"
	helpers "lfiber/internal/support"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

const gracefulShutdownTimeout = 15 * time.Second

// Serve starts the HTTP server and handles graceful shutdown
func Serve(app *fiber.App, cfg *configs.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listenAddr := net.JoinHostPort(cfg.App.Host, cfg.App.Port)
	baseURL := buildPublicURL(cfg.App.Host, cfg.App.Port)
	docsURL := baseURL + "/docs"

	helpers.Info(
		"server_listening",
		zap.String("listen_addr", listenAddr),
		zap.String("host", cfg.App.Host),
		zap.String("port", cfg.App.Port),
		zap.String("app_url", baseURL),
		zap.String("docs_url", docsURL),
	)

	listenErr := make(chan error, 1)
	go func() {
		listenErr <- app.Listen(listenAddr, fiber.ListenConfig{
			EnablePrefork:         cfg.App.Fiber.Prefork,
			DisableStartupMessage: true,
		})
	}()

	select {
	case err := <-listenErr:
		if err != nil {
			if isAddressInUse(err) {
				return fmt.Errorf("server_port_in_use: listen_addr=%s port=%s: %w", listenAddr, cfg.App.Port, err)
			}

			return fmt.Errorf("server_failed_to_start: listen_addr=%s: %w", listenAddr, err)
		}
		return nil
	case <-ctx.Done():
		helpers.Info("shutdown_signal_received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			helpers.Warn("server_shutdown_timed_out", zap.Duration("timeout", gracefulShutdownTimeout))
			return fmt.Errorf("server_shutdown_timed_out: timeout=%s: %w", gracefulShutdownTimeout, err)
		}

		helpers.Warn("server_shutdown_failed", zap.Error(err))
		return fmt.Errorf("server_shutdown_failed: %w", err)
	}

	return nil
}

func buildPublicURL(host, port string) string {
	host = strings.TrimSpace(host)
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "localhost"
	}

	return "http://" + net.JoinHostPort(host, port)
}

func isAddressInUse(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, syscall.EADDRINUSE) {
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), "address already in use")
}
