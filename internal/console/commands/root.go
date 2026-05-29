package command

import (
	"fmt"
	"os"
	"strings"

	appcmd "lfiber/internal/console/commands/app"
	authcmd "lfiber/internal/console/commands/auth"
	backupcmd "lfiber/internal/console/commands/backup"
	cachecmd "lfiber/internal/console/commands/cache"
	configcmd "lfiber/internal/console/commands/config"
	dbcmd "lfiber/internal/console/commands/db"
	hashcmd "lfiber/internal/console/commands/hash"
	jwtcmd "lfiber/internal/console/commands/jwt"
	mediacmd "lfiber/internal/console/commands/media"
	queuecmd "lfiber/internal/console/commands/queue"
	routecmd "lfiber/internal/console/commands/route"
	schedulecmd "lfiber/internal/console/commands/schedule"
	servecmd "lfiber/internal/console/commands/serve"
	"lfiber/internal/console/ui"
	artisan "lfiber/internal/providers/artisan"

	"github.com/spf13/cobra"
)

func init() {
	artisan.SetCommandFactory(NewRootCommand)
}

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
		&cobra.Group{ID: "database", Title: "Database Commands"},
		&cobra.Group{ID: "cache", Title: "Cache Commands"},
		&cobra.Group{ID: "auth", Title: "Authentication & Security"},
		&cobra.Group{ID: "queue", Title: "Queue Commands"},
		&cobra.Group{ID: "schedule", Title: "Task Scheduling"},
		&cobra.Group{ID: "system", Title: "System & Configuration"},
	)

	// Register serve command and ensure it belongs to 'app' group
	serveCmd := servecmd.New()
	serveCmd.GroupID = "app"
	root.AddCommand(serveCmd)

	root.AddCommand(appcmd.Commands()...)
	root.AddCommand(routecmd.Commands()...)
	root.AddCommand(configcmd.Commands()...)
	root.AddCommand(cachecmd.Commands()...)
	root.AddCommand(authcmd.Commands()...)
	root.AddCommand(backupcmd.Commands()...)
	root.AddCommand(dbcmd.Commands()...)
	root.AddCommand(hashcmd.Commands()...)
	root.AddCommand(jwtcmd.Commands()...)
	root.AddCommand(mediacmd.Commands()...)
	root.AddCommand(queuecmd.Commands()...)
	root.AddCommand(schedulecmd.Commands()...)

	// Configure lipgloss colorized help function
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		// Color styles resolved from ui package
		yellowStyle := ui.YellowStyle(out)
		cyanStyle := ui.CyanStyle(out)
		greenStyle := ui.GreenStyle(out)
		grayStyle := ui.GrayStyle(out)

		if cmd.Long != "" {
			_, _ = fmt.Fprintln(out, grayStyle.Render(cmd.Long))
		} else if cmd.Short != "" {
			_, _ = fmt.Fprintln(out, grayStyle.Render(cmd.Short))
		}
		_, _ = fmt.Fprintln(out)

		_, _ = fmt.Fprintln(out, yellowStyle.Render("Usage:"))
		// Highlight lfiber in green and the rest [flags] / [command] in gray
		useLine := cmd.UseLine()
		parts := strings.SplitN(useLine, " ", 2)
		if len(parts) == 2 {
			_, _ = fmt.Fprintf(out, "  %s %s\n", greenStyle.Render(parts[0]), grayStyle.Render(parts[1]))
		} else {
			_, _ = fmt.Fprintf(out, "  %s\n", greenStyle.Render(useLine))
		}

		if len(cmd.Groups()) > 0 {
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, yellowStyle.Render("Available Commands Groups:"))
			for _, g := range cmd.Groups() {
				// Show group title
				_, _ = fmt.Fprintf(out, "  %s\n", cyanStyle.Render(g.Title))

				// Find all commands belonging to this group
				for _, subCmd := range cmd.Commands() {
					if subCmd.GroupID == g.ID && subCmd.IsAvailableCommand() {
						padding := 16
						if len(subCmd.Name()) > padding {
							padding = len(subCmd.Name()) + 2
						}
						formattedName := fmt.Sprintf("    %-*s", padding, subCmd.Name())
						_, _ = fmt.Fprintf(out, "%s  %s\n", greenStyle.Render(formattedName), grayStyle.Render(subCmd.Short))
					}
				}
			}
		}

		// Additional Commands (without Group ID)
		hasAdditional := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.GroupID == "" && subCmd.IsAvailableCommand() {
				hasAdditional = true
				break
			}
		}
		if hasAdditional {
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, yellowStyle.Render("Additional Commands:"))
			for _, subCmd := range cmd.Commands() {
				if subCmd.GroupID == "" && subCmd.IsAvailableCommand() {
					padding := 16
					formattedName := fmt.Sprintf("  %-*s", padding, subCmd.Name())
					_, _ = fmt.Fprintf(out, "%s  %s\n", greenStyle.Render(formattedName), grayStyle.Render(subCmd.Short))
				}
			}
		}

		if cmd.HasAvailableLocalFlags() {
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, yellowStyle.Render("Flags:"))
			_, _ = fmt.Fprintln(out, cmd.LocalFlags().FlagUsages())
		}

		if cmd.HasAvailableInheritedFlags() {
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, yellowStyle.Render("Global Flags:"))
			_, _ = fmt.Fprintln(out, cmd.InheritedFlags().FlagUsages())
		}

		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintf(out, "Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
	})

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
