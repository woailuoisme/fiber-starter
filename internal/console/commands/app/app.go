package app

import (
	"fmt"
	"runtime"
	"text/tabwriter"

	"lfiber/configs"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{aboutCommand()}
}

func aboutCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "app:about",
		Short:   "Display application information",
		GroupID: "app",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, _, err := configs.LoadConfig()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "KEY\tVALUE")
			_, _ = fmt.Fprintf(w, "Name\t%s\n", cfg.App.Name)
			_, _ = fmt.Fprintf(w, "Environment\t%s\n", cfg.App.Env)
			_, _ = fmt.Fprintf(w, "Debug\t%v\n", cfg.App.Debug)
			_, _ = fmt.Fprintf(w, "URL\t%s\n", cfg.App.URL)
			_, _ = fmt.Fprintf(w, "Host\t%s\n", cfg.App.Host)
			_, _ = fmt.Fprintf(w, "Port\t%s\n", cfg.App.Port)
			_, _ = fmt.Fprintf(w, "Go\t%s\n", runtime.Version())
			return w.Flush()
		},
	}
}
