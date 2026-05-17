package command

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	database "fiber-starter/app/Providers/Database"
	"fiber-starter/configs"
	"fiber-starter/database/seeders"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ─── Driver Constants ────────────────────────────────────────────────────────

const (
	driverSQLite     = "sqlite"
	driverSQLite3    = "sqlite3"
	driverPostgres   = "postgres"
	driverPostgreSQL = "postgresql"
	dialectPostgres  = "psql"
	dialectSQLite    = "sqlite"
)

// ─── DB Command Group ────────────────────────────────────────────────────────

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database operation management",
	Long:  `Provide various database-related operations, including migration, seed data management, etc.`,
}

var dbSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup database (run migrations and seed data)",
	Long: `Setup database, run all migrations and populate seed data.
This command executes migration and seed data operations in sequence, completing database initialization.`,
	Run: func(_ *cobra.Command, _ []string) { setupDatabase() },
}

// ─── Migrate Command Group ───────────────────────────────────────────────────

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Database migration management",
		Long:  `Manage database migrations, including running migrations, rolling back migrations, etc.`,
	}

	migrateRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run all pending database migrations",
		Long: `Run all pending database migrations.
This command executes all migrations that have not yet been run, updating database structure to latest state.`,
		Run: func(_ *cobra.Command, _ []string) { runMigrations() },
	}

	migrateRollbackCmd = &cobra.Command{
		Use:   "rollback",
		Short: "Rollback last database migration",
		Long: `Rollback last executed database migration.
This command undoes last migration operation, restoring database to its pre-migration state.`,
		Run: func(_ *cobra.Command, _ []string) { rollbackMigrations() },
	}

	migrateResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset database (delete all tables and re-run migrations)",
		Long: `Reset database, delete all tables and re-run all migrations.
Warning: This operation will delete all data, use with caution!`,
		Run: func(_ *cobra.Command, _ []string) { resetDatabase() },
	}

	migrateFreshCmd = &cobra.Command{
		Use:   "fresh",
		Short: "Drop all tables and re-run migrations and seed data",
		Long: `Drop all tables and re-run all migrations, then run seed data.
Warning: This operation will delete all data, use with caution!`,
		Run: func(_ *cobra.Command, _ []string) { freshDatabase() },
	}

	migrateStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  `Show current status of database migrations, including run and pending migrations.`,
		Run:   func(_ *cobra.Command, _ []string) { showMigrationStatus() },
	}
)

// ─── Seed Command Group ──────────────────────────────────────────────────────

var (
	seedCmd = &cobra.Command{
		Use:   "seed",
		Short: "Database seed data management",
		Long:  `Manage database seed data, including running seed data, clearing seed data, etc.`,
	}

	seedRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run all seed data",
		Long: `Run all seed data, populating database with initial data.
This command executes all seed data creation operations.`,
		Run: func(_ *cobra.Command, _ []string) { runSeeds() },
	}

	seedRunRandomCmd = &cobra.Command{
		Use:   "random [count]",
		Short: "Generate specified number of random test data",
		Long:  `Generate specified number of random test data. If no count is specified, defaults to 10 records.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			count := 10
			if len(args) > 0 {
				count = parsePositiveInt(args[0], count)
			}
			runRandomSeeds(count)
		},
	}

	seedClearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear all seed data",
		Long:  `Clear all seed data, deleting records created by seed data. This operation deletes all seed data but does not delete table structures.`,
		Run:   func(_ *cobra.Command, _ []string) { clearSeeds() },
	}

	seedRefreshCmd = &cobra.Command{
		Use:   "refresh",
		Short: "Refresh seed data (clear and re-run)",
		Long:  `Refresh seed data, first clear all seed data, then re-run. This command clears existing seed data and then repopulates it.`,
		Run:   func(_ *cobra.Command, _ []string) { refreshSeeds() },
	}
)

func init() {
	// db group
	dbCmd.AddCommand(dbSetupCmd)
	rootCmd.AddCommand(dbCmd)

	// migrate group
	migrateCmd.AddCommand(migrateRunCmd, migrateRollbackCmd, migrateResetCmd, migrateFreshCmd, migrateStatusCmd)
	rootCmd.AddCommand(migrateCmd)

	// seed group
	seedCmd.AddCommand(seedRunCmd, seedRunRandomCmd, seedClearCmd, seedRefreshCmd)
	rootCmd.AddCommand(seedCmd)
}

// ─── Migration Operations ────────────────────────────────────────────────────

func runMigrations() {
	color.Cyan("Running database migrations...")
	if err := runAtlasForCurrentConnection("migrate", "apply"); err != nil {
		color.Red("Migration failed: %v", err)
		os.Exit(1)
	}
	color.Green("Database migration completed")
}

func rollbackMigrations() {
	color.Cyan("Rolling back database migrations...")
	if err := runAtlasForCurrentConnection("migrate", "down", "1"); err != nil {
		color.Red("Rollback failed: %v", err)
		os.Exit(1)
	}
	color.Green("Database migration rollback completed")
}

func resetDatabase() {
	if !confirmDestructiveAction() {
		color.Yellow("Operation cancelled")
		return
	}
	rebuildDatabase("Resetting database...", false)
	color.Green("Database reset completed")
}

func freshDatabase() {
	if !confirmDestructiveAction() {
		color.Yellow("Operation cancelled")
		return
	}
	rebuildDatabase("Dropping all tables and re-running migrations...", true)
	color.Green("Database fresh completed")
}

func setupDatabase() {
	rebuildDatabase("Setting up database...", true)
	color.Green("Database setup completed")
}

func showMigrationStatus() {
	color.Cyan("Checking migration status...")
	if err := runAtlasForCurrentConnection("migrate", "status"); err != nil {
		color.Red("Failed to get migration status: %v", err)
		os.Exit(1)
	}
}

// ─── Seed Operations ─────────────────────────────────────────────────────────

func runSeeds() {
	runSeedOperation("Failed to run seeds", seeders.RunAllSeeders)
	color.Green("Seed data completed")
}

func runRandomSeeds(count int) {
	runSeedOperation("Failed to run random seeds", func(db *sql.DB, dialect string) error {
		return seeders.RunRandomSeeders(db, dialect, count)
	})
	color.Green("Random seed data completed")
}

func clearSeeds() {
	runSeedOperation("Failed to clear seeds", seeders.ClearAllSeeders)
	color.Green("Seed data cleared")
}

func refreshSeeds() {
	runSeedOperation("Failed to refresh seeds", func(db *sql.DB, dialect string) error {
		if err := seeders.ClearAllSeeders(db, dialect); err != nil {
			return err
		}
		return seeders.RunAllSeeders(db, dialect)
	})
	color.Green("Seed data refreshed")
}

// ─── Internal Helpers ─────────────────────────────────────────────────────────

func initDBWithConfig() (*sql.DB, *configs.Config, error) {
	cfg, _, err := configs.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	conn := database.NewConnection(cfg)

	db, err := conn.GetDB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, cfg, nil
}

func getDefaultConnection() (configs.DBConnection, error) {
	cfg, _, err := configs.LoadConfig()
	if err != nil {
		return configs.DBConnection{}, fmt.Errorf("failed to load config: %w", err)
	}

	dbConfig := &cfg.Database
	defaultConn := dbConfig.Default
	connConfig, exists := dbConfig.Connections[defaultConn]
	if !exists {
		return configs.DBConnection{}, fmt.Errorf("database connection config '%s' does not exist", defaultConn)
	}

	return connConfig, nil
}

func migrationEnvName(driver string) string {
	if isSQLiteDriver(driver) {
		return driverSQLite
	}
	return driverPostgres
}

func runAtlasForConnection(connConfig configs.DBConnection, args ...string) error {
	fullArgs := append([]string{}, args...)
	fullArgs = append(fullArgs, "--env", migrationEnvName(connConfig.Driver))
	return runAtlas(fullArgs...)
}

func runAtlasForCurrentConnection(args ...string) error {
	connConfig, err := getDefaultConnection()
	if err != nil {
		return err
	}
	return runAtlasForConnection(connConfig, args...)
}

func runAtlas(args ...string) error {
	// #nosec G204 -- atlas is an expected local developer tool invoked with controlled args.
	cmd := exec.Command("atlas", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func rebuildDatabase(startMessage string, withSeed bool) {
	color.Cyan(startMessage)

	connConfig, err := getDefaultConnection()
	if err != nil {
		color.Red("Failed to read database config: %v", err)
		os.Exit(1)
	}

	if isSQLiteDriver(connConfig.Driver) {
		_ = os.Remove(connConfig.Database)
	}

	if err := runAtlasForConnection(connConfig, "migrate", "apply"); err != nil {
		color.Red("Migration failed: %v", err)
		os.Exit(1)
	}

	if !withSeed {
		return
	}

	color.Cyan("Running seed data...")
	runSeedOperation("Failed to run seed data", seeders.RunAllSeeders)
}

func runSeedOperation(action string, fn func(db *sql.DB, dialect string) error) {
	db, cfg, err := initDBWithConfig()
	if err != nil {
		color.Red("Failed to initialize database connection: %v", err)
		os.Exit(1)
	}

	if err := fn(db, seedDialectFromConfig(cfg)); err != nil {
		color.Red("%s: %v", action, err)
		os.Exit(1)
	}
}

func seedDialectFromConfig(cfg *configs.Config) string {
	if cfg == nil {
		return dialectPostgres
	}

	defaultConn := cfg.Database.Default
	conn, ok := cfg.Database.Connections[defaultConn]
	if !ok {
		return dialectPostgres
	}

	switch strings.ToLower(strings.TrimSpace(conn.Driver)) {
	case driverSQLite, driverSQLite3:
		return dialectSQLite
	case driverPostgres, driverPostgreSQL:
		return dialectPostgres
	default:
		return dialectPostgres
	}
}

func isSQLiteDriver(driver string) bool {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case driverSQLite, driverSQLite3:
		return true
	default:
		return false
	}
}

func confirmDestructiveAction() bool {
	fmt.Print("Warning: This will delete all data! Are you sure you want to continue? (y/N): ")
	var response string
	_, _ = fmt.Scanln(&response)
	return strings.EqualFold(response, "y") || strings.EqualFold(response, "yes")
}
