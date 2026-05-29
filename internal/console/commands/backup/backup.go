package backup

import (
	"fmt"
	"text/tabwriter"

	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/console/prompts"
	"lfiber/internal/console/ui"
	backuppkg "lfiber/pkg/backup"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{runCommand(), restoreCommand(), listCommand()}
}

func runCommand() *cobra.Command {
	var opts backuppkg.RunOptions
	cmd := &cobra.Command{
		Use:     "backup:run",
		Short:   "Backup the configured database connection",
		GroupID: "database",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := commandutil.BuildRuntime()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(rt) }()
			if rt.Backup == nil {
				return fmt.Errorf("backup service is not available")
			}
			result, err := rt.Backup.Run(cmd.Context(), opts)
			if err != nil {
				return err
			}
			ui.Success(cmd.OutOrStdout(), "Backup stored: %s", result.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.ConnectionName, "connection", "", "database connection to backup")
	cmd.Flags().BoolVar(&opts.DisableNotifications, "disable-notifications", false, "disable backup notifications for this run")
	return cmd
}

func restoreCommand() *cobra.Command {
	var opts backuppkg.RestoreOptions
	var force bool
	cmd := &cobra.Command{
		Use:     "backup:restore <storage-path>",
		Short:   "Restore a database backup from storage",
		GroupID: "database",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			confirmed, err := prompts.ConfirmDestructive(cmd.InOrStdin(), cmd.OutOrStdout(), prompts.DestructiveOptions{
				Force:         force,
				NoInteraction: commandutil.NoInteraction(cmd),
				Message:       "This will restore database data from a backup.",
			})
			if err != nil {
				return fmt.Errorf("confirm database restore: %w", err)
			}
			if !confirmed {
				prompts.Cancelled(cmd.OutOrStdout())
				return nil
			}
			rt, err := commandutil.BuildRuntime()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(rt) }()
			if rt.Backup == nil {
				return fmt.Errorf("backup service is not available")
			}
			opts.Path = args[0]
			if err := rt.Backup.Restore(cmd.Context(), opts); err != nil {
				return err
			}
			ui.Success(cmd.OutOrStdout(), "Backup restored: %s", opts.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.ConnectionName, "connection", "", "database connection to restore into")
	cmd.Flags().BoolVar(&opts.DisableNotifications, "disable-notifications", false, "disable backup notifications for this run")
	cmd.Flags().BoolVar(&force, "force", false, "skip destructive action confirmation")
	return cmd
}

func listCommand() *cobra.Command {
	var opts backuppkg.ListOptions
	cmd := &cobra.Command{
		Use:     "backup:list",
		Short:   "List database backups in storage",
		GroupID: "database",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := commandutil.BuildRuntime()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(rt) }()
			if rt.Backup == nil {
				return fmt.Errorf("backup service is not available")
			}
			backups, err := rt.Backup.List(cmd.Context(), opts)
			if err != nil {
				return err
			}
			if len(backups) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No backups found")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "PATH\tCONNECTION\tSIZE\tMODIFIED")
			for _, backup := range backups {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", backup.Path, backup.Connection, backup.Size, backup.LastModified.Format("2006-01-02 15:04:05"))
			}
			return w.Flush()
		},
	}
	cmd.Flags().StringVar(&opts.ConnectionName, "connection", "", "database connection to list")
	return cmd
}
