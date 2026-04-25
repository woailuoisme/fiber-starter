package command

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
)

func exitWithError(format string, args ...any) {
	_, _ = color.New(color.FgRed).Printf(format+"\n", args...)
	os.Exit(1)
}

func waitForInterrupt() <-chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	return quit
}

func parsePositiveInt(value string, fallback int) int {
	var parsed int
	if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed); err == nil && parsed > 0 {
		return parsed
	}
	return fallback
}
