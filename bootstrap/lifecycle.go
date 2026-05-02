package bootstrap

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func runHTTPServer(app *fiber.App, cfg *config.Config) error {
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-listenErr:
		if err != nil {
			if isAddressInUse(err) {
				return fmt.Errorf("server_port_in_use: listen_addr=%s port=%s: %w", listenAddr, cfg.App.Port, err)
			}

			return fmt.Errorf("server_failed_to_start: listen_addr=%s: %w", listenAddr, err)
		}
		return nil
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
