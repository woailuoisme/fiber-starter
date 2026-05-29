package media

import (
	"fmt"

	"lfiber/internal/console/commands/commandutil"
	"lfiber/pkg/medialibrary"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{regenerateCommand()}
}

func regenerateCommand() *cobra.Command {
	var opts medialibrary.RegenerateOptions
	cmd := &cobra.Command{
		Use:     "media:regenerate",
		Short:   "Regenerate derived media files",
		GroupID: "app",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := commandutil.BuildRuntime()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(rt) }()
			if rt.MediaLibrary == nil {
				return fmt.Errorf("media library is not available")
			}
			count, err := rt.MediaLibrary.Regenerate(cmd.Context(), opts)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Regenerated derived media for %d media item(s)\n", count)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.OnlyMissing, "only-missing", false, "only regenerate missing derived media")
	cmd.Flags().BoolVar(&opts.FailedOnly, "failed", false, "only regenerate failed derived media")
	cmd.Flags().StringVar(&opts.Collection, "collection", "", "limit regeneration to a media collection")
	return cmd
}
