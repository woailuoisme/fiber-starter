package command

import (
	helpers "fiber-starter/app/Support"
	"fiber-starter/config"
)

// CLI starts the command-line application.
func CLI() {
	if err := config.Init(); err != nil {
		exitWithError("Failed to initialize config: %v", err)
	}

	if err := helpers.Init(); err != nil {
		exitWithError("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = helpers.Sync()
	}()

	Execute()
}
