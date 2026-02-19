// Package command contains all command-line tool implementations
package command

import (
	"fmt"
	"os"

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"
)

// CLI starts the command-line tool
func CLI() {
	// Initialize config (for command-line tool)
	if err := config.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := helpers.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = helpers.Sync()
	}()

	// Execute command
	Execute()
}
