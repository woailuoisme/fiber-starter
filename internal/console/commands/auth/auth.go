package auth

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"lfiber/configs"
	"lfiber/internal/support"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{aboutCommand()}
}

func aboutCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "auth:about",
		Short:   "Display authentication configuration summary",
		GroupID: "system",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, _, err := configs.LoadConfig()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "KEY\tVALUE")
			_, _ = fmt.Fprintf(w, "Default Guard\t%s\n", cfg.Auth.Default)
			_, _ = fmt.Fprintf(w, "JWT Issuer\t%s\n", cfg.JWT.Issuer)
			_, _ = fmt.Fprintf(w, "JWT Secret\t%s\n", support.RedactSensitive("secret="+cfg.JWT.Secret))

			guards := make([]string, 0, len(cfg.Auth.Guards))
			for name := range cfg.Auth.Guards {
				guards = append(guards, name)
			}
			sort.Strings(guards)
			for _, name := range guards {
				guard := cfg.Auth.Guards[name]
				_, _ = fmt.Fprintf(w, "Guard %s\t%s / %s\n", name, guard.Driver, guard.Provider)
			}
			return w.Flush()
		},
	}
}
