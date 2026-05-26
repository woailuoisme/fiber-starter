package cache

import (
	"fmt"

	"lfiber/internal/console/commands/commandutil"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{clearCommand(), forgetCommand(), hasCommand(), ttlCommand()}
}

func clearCommand() *cobra.Command {
	return &cobra.Command{Use: "cache:clear", Short: "Flush the default cache store", GroupID: "cache", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		if rt.Cache == nil {
			return fmt.Errorf("cache store is not available")
		}
		if err := rt.Cache.Flush(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cache cleared")
		return nil
	}}
}

func forgetCommand() *cobra.Command {
	return &cobra.Command{Use: "cache:forget <key>", Short: "Remove a key from the default cache store", GroupID: "cache", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		if rt.Cache == nil {
			return fmt.Errorf("cache store is not available")
		}
		if err := rt.Cache.Forget(args[0]); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Forgot cache key: %s\n", args[0])
		return nil
	}}
}

func hasCommand() *cobra.Command {
	return &cobra.Command{Use: "cache:has <key>", Short: "Check if a cache key exists", GroupID: "cache", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		if rt.Cache == nil {
			return fmt.Errorf("cache store is not available")
		}
		exists, err := rt.Cache.Has(args[0])
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%v\n", exists)
		return nil
	}}
}

func ttlCommand() *cobra.Command {
	return &cobra.Command{Use: "cache:ttl <key>", Short: "Show a cache key TTL", GroupID: "cache", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		rt, err := commandutil.BuildRuntime()
		if err != nil {
			return err
		}
		defer func() { _ = commandutil.CloseRuntime(rt) }()
		if rt.Cache == nil {
			return fmt.Errorf("cache store is not available")
		}
		ttl, err := rt.Cache.TTL(args[0])
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), ttl)
		return nil
	}}
}
