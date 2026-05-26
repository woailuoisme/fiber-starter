package config

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"lfiber/configs"
	"lfiber/internal/support"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{showCommand()}
}

func showCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "config:show [key]",
		Short:   "Display loaded configuration",
		GroupID: "system",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, k, err := configs.LoadConfig()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				value := k.Get(args[0])
				encoded, err := json.MarshalIndent(value, "", "  ")
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), redact(args[0], string(encoded)))
				return nil
			}

			all := k.All()
			keys := make([]string, 0, len(all))
			for key := range all {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "KEY\tVALUE")
			for _, key := range keys {
				_, _ = fmt.Fprintf(w, "%s\t%v\n", key, redact(key, fmt.Sprint(all[key])))
			}
			return w.Flush()
		},
	}
}

func redact(key, value string) string {
	lowerKey := strings.ToLower(key)
	if strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "key") {
		return support.RedactionSentinel()
	}
	return support.RedactSensitive(value)
}
