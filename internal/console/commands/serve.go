package command

import (
	"fiber-starter/internal/bootstrap"
	helpers "fiber-starter/internal/support"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  `Start the Fiber HTTP server with the configured host, port, and middleware stack.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := bootstrap.App(); err != nil {
			// Handle common startup errors (like port in use) gracefully
			helpers.HandleServerStartError(err, "")

			// Fallback for unexpected errors
			helpers.Fatal("server_bootstrap_failed", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
