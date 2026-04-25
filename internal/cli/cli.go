// Package command contains all command-line tool implementations
package command

import (
	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"
)

// CLI starts the command-line tool
func CLI() {
	// Initialize config (for command-line tool)
	if err := config.Init(); err != nil {
		exitWithError("Failed to initialize config: %v", err)
	}

	// Initialize logger
	if err := helpers.Init(); err != nil {
		exitWithError("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = helpers.Sync()
	}()

	// Execute command
	Execute()
}
