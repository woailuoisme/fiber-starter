package hash

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"lfiber/configs"
	hashprovider "lfiber/internal/providers/hash"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{makeCommand(), checkCommand(), infoCommand()}
}

func makeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "hash:make <value>",
		Short:   "Hash a value using the configured hash driver",
		GroupID: "system",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hasher, err := configuredHasher()
			if err != nil {
				return err
			}
			hashed, err := hasher.Make(args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), hashed)
			return nil
		},
	}
}

func checkCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "hash:check <value> <hash>",
		Short:   "Check a value against a hash",
		GroupID: "system",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			hasher, err := configuredHasher()
			if err != nil {
				return err
			}
			if hasher.Check(args[0], args[1]) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Hash verified")
				return nil
			}
			return fmt.Errorf("hash verification failed")
		},
	}
}

func infoCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "hash:info <hash>",
		Short:   "Display hash metadata",
		GroupID: "system",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hasher, err := configuredHasher()
			if err != nil {
				return err
			}
			info := hasher.Info(args[0])
			keys := make([]string, 0, len(info))
			for key := range info {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "KEY\tVALUE")
			for _, key := range keys {
				_, _ = fmt.Fprintf(w, "%s\t%v\n", key, info[key])
			}
			return w.Flush()
		},
	}
}

func configuredHasher() (*hashprovider.Manager, error) {
	cfg, _, err := configs.LoadConfig()
	if err != nil {
		return nil, err
	}
	if cfg.Hash.Driver == "" {
		cfg.Hash.Driver = "bcrypt"
	}
	if cfg.Hash.Bcrypt.Rounds == 0 {
		cfg.Hash.Bcrypt.Rounds = 10
	}
	return hashprovider.NewHashManager(cfg), nil
}
