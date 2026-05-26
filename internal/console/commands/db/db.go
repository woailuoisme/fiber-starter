package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"lfiber/configs"
	"lfiber/database/seeders"
	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/console/prompts"
	"lfiber/internal/console/ui"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{
		migrateCommand(), rollbackCommand(), resetCommand(), freshCommand(), setupCommand(), statusCommand(), seedCommand(), seedRandomCommand(),
	}
}

func migrateCommand() *cobra.Command {
	return &cobra.Command{Use: "db:migrate", Short: "Run all pending database migrations", GroupID: "database", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		ui.Info(cmd.OutOrStdout(), "Running database migrations...")
		if err := commandutil.RunAtlasForCurrentConnection("migrate", "apply"); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
		ui.Success(cmd.OutOrStdout(), "Database migration completed")
		return nil
	}}
}

func rollbackCommand() *cobra.Command {
	return &cobra.Command{Use: "db:rollback", Short: "Rollback last database migration", GroupID: "database", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		ui.Info(cmd.OutOrStdout(), "Rolling back database migrations...")
		if err := commandutil.RunAtlasForCurrentConnection("migrate", "down", "1"); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}
		ui.Success(cmd.OutOrStdout(), "Database migration rollback completed")
		return nil
	}}
}

func resetCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{Use: "db:reset", Short: "Reset database and re-run migrations", GroupID: "database", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		confirmed, err := prompts.ConfirmDestructive(cmd.InOrStdin(), cmd.OutOrStdout(), prompts.DestructiveOptions{
			Force:         force,
			NoInteraction: commandutil.NoInteraction(cmd),
			Message:       "This will delete database data and re-run migrations.",
		})
		if err != nil {
			return fmt.Errorf("confirm database reset: %w", err)
		}
		if !confirmed {
			prompts.Cancelled(cmd.OutOrStdout())
			return nil
		}
		if err := rebuildDatabase(cmd, "Resetting database...", false); err != nil {
			return err
		}
		ui.Success(cmd.OutOrStdout(), "Database reset completed")
		return nil
	}}
	cmd.Flags().BoolVar(&force, "force", false, "skip destructive action confirmation")
	return cmd
}

func freshCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{Use: "db:fresh", Short: "Drop all tables, migrate, and seed", GroupID: "database", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		confirmed, err := prompts.ConfirmDestructive(cmd.InOrStdin(), cmd.OutOrStdout(), prompts.DestructiveOptions{
			Force:         force,
			NoInteraction: commandutil.NoInteraction(cmd),
			Message:       "This will drop all tables, run migrations, and seed data.",
		})
		if err != nil {
			return fmt.Errorf("confirm database fresh: %w", err)
		}
		if !confirmed {
			prompts.Cancelled(cmd.OutOrStdout())
			return nil
		}
		if err := rebuildDatabase(cmd, "Dropping all tables and re-running migrations...", true); err != nil {
			return err
		}
		ui.Success(cmd.OutOrStdout(), "Database fresh completed")
		return nil
	}}
	cmd.Flags().BoolVar(&force, "force", false, "skip destructive action confirmation")
	return cmd
}

func setupCommand() *cobra.Command {
	return &cobra.Command{Use: "db:setup", Short: "Setup database with migrations and seed data", GroupID: "database", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if err := rebuildDatabase(cmd, "Setting up database...", true); err != nil {
			return err
		}
		ui.Success(cmd.OutOrStdout(), "Database setup completed")
		return nil
	}}
}

func statusCommand() *cobra.Command {
	return &cobra.Command{Use: "db:status", Short: "Show migration status", GroupID: "database", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		ui.Info(cmd.OutOrStdout(), "Checking migration status...")
		if err := commandutil.RunAtlasForCurrentConnection("migrate", "status"); err != nil {
			return fmt.Errorf("failed to get migration status: %w", err)
		}
		return nil
	}}
}

func seedCommand() *cobra.Command {
	return &cobra.Command{Use: "db:seed", Short: "Run all seed data", GroupID: "database", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if err := runSeedOperation(seeders.RunAllSeeders); err != nil {
			return fmt.Errorf("failed to run seeds: %w", err)
		}
		ui.Success(cmd.OutOrStdout(), "Seed data completed")
		return nil
	}}
}

func seedRandomCommand() *cobra.Command {
	var countOption int
	cmd := &cobra.Command{
		Use:     "db:seed-random [count]",
		Short:   "Generate random test data",
		GroupID: "database",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			count := countOption
			// If positional argument is provided, it takes precedence
			if len(args) > 0 {
				count = commandutil.ParsePositiveInt(args[0], countOption)
			}
			if err := runSeedOperation(func(db *sql.DB, dialect string) error { return seeders.RunRandomSeeders(db, dialect, count) }); err != nil {
				return fmt.Errorf("failed to run random seeds: %w", err)
			}
			ui.Success(cmd.OutOrStdout(), "Random seed data completed")
			return nil
		},
	}
	cmd.Flags().IntVarP(&countOption, "count", "n", 10, "number of random test items to seed")
	return cmd
}

func rebuildDatabase(cmd *cobra.Command, startMessage string, withSeed bool) error {
	ui.Info(cmd.OutOrStdout(), "%s", startMessage)
	connConfig, err := commandutil.GetDefaultConnection()
	if err != nil {
		return fmt.Errorf("failed to read database config: %w", err)
	}

	if commandutil.IsSQLiteDriver(connConfig.Driver) {
		if err := os.Remove(connConfig.Database); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove sqlite database: %w", err)
		}
	}
	if err := commandutil.RunAtlasForConnection(connConfig, "migrate", "apply"); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	if !withSeed {
		return nil
	}
	ui.Info(cmd.OutOrStdout(), "Running seed data...")
	return runSeedOperation(seeders.RunAllSeeders)
}

func runSeedOperation(fn func(db *sql.DB, dialect string) error) error {
	db, cfg, err := commandutil.InitDBWithConfig()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	return fn(db, seedDialectFromConfig(cfg))
}

func seedDialectFromConfig(cfg *configs.Config) string {
	if cfg == nil {
		return "psql"
	}
	defaultConn := cfg.Database.Default
	conn, ok := cfg.Database.Connections[defaultConn]
	if !ok {
		return "psql"
	}
	switch strings.ToLower(strings.TrimSpace(conn.Driver)) {
	case "sqlite", "sqlite3":
		return "sqlite"
	case "postgres", "postgresql":
		return "psql"
	default:
		return "psql"
	}
}
