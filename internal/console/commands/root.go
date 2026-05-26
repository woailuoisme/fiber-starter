package command

import (
	"fmt"
	"os"

	appcmd "lfiber/internal/console/commands/app"
	authcmd "lfiber/internal/console/commands/auth"
	cachecmd "lfiber/internal/console/commands/cache"
	configcmd "lfiber/internal/console/commands/config"
	dbcmd "lfiber/internal/console/commands/db"
	hashcmd "lfiber/internal/console/commands/hash"
	jwtcmd "lfiber/internal/console/commands/jwt"
	queuecmd "lfiber/internal/console/commands/queue"
	routecmd "lfiber/internal/console/commands/route"
	schedulecmd "lfiber/internal/console/commands/schedule"
	servecmd "lfiber/internal/console/commands/serve"
	"lfiber/internal/console/ui"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "lfiber",
		Short:         "lfiber application command-line tool",
		Long:          "lfiber is an application starter based on the Go Fiber framework.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	root.PersistentFlags().StringP("config", "c", "", "config file (default is .env)")
	root.PersistentFlags().Bool("no-interaction", false, "disable interactive prompts")
	root.AddGroup(
		&cobra.Group{ID: "app", Title: "Application Commands"},
		&cobra.Group{ID: "system", Title: "System Commands"},
		&cobra.Group{ID: "database", Title: "Database Commands"},
		&cobra.Group{ID: "queue", Title: "Queue Commands"},
	)

	root.AddCommand(servecmd.New())
	root.AddCommand(appcmd.Commands()...)
	root.AddCommand(routecmd.Commands()...)
	root.AddCommand(configcmd.Commands()...)
	root.AddCommand(cachecmd.Commands()...)
	root.AddCommand(authcmd.Commands()...)
	root.AddCommand(dbcmd.Commands()...)
	root.AddCommand(hashcmd.Commands()...)
	root.AddCommand(jwtcmd.Commands()...)
	root.AddCommand(queuecmd.Commands()...)
	root.AddCommand(schedulecmd.Commands()...)

	return root
}

func Execute() error {
	return NewRootCommand().Execute()
}

func CLI() {
	if err := Execute(); err != nil {
		ui.Error(os.Stderr, "Error executing command: %v", err)
		os.Exit(1)
	}
}

func GetRootCmd() *cobra.Command {
	return NewRootCommand()
}

func ExecuteArgs(args ...string) error {
	root := NewRootCommand()
	root.SetArgs(args)
	return root.Execute()
}

func Usage(root *cobra.Command) string {
	return fmt.Sprintf("%s\n", root.UseLine())
}
