package serve

import (
	"lfiber/internal/bootstrap"
	"lfiber/internal/support"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func New() *cobra.Command {
	return &cobra.Command{
		Use:     "serve",
		Short:   "Start the HTTP server",
		Long:    "Start the Fiber HTTP server with the configured host, port, and middleware stack.",
		GroupID: "app",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := bootstrap.App(); err != nil {
				support.HandleServerStartError(err, "")
				support.Fatal("server_bootstrap_failed", zap.Error(err))
			}
			return nil
		},
	}
}
