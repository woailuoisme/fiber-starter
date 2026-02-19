package command

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fiber-starter/internal/config"
	"fiber-starter/internal/db/seeders"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	database "fiber-starter/internal/db"
)

const (
	driverSQLite     = "sqlite"
	driverSQLite3    = "sqlite3"
	driverPostgres   = "postgres"
	driverPostgreSQL = "postgresql"
	dialectPostgres  = "psql"
	dialectSQLite    = "sqlite"
)

// initDB Initialize database connection
func initDB() error {
	_, _, err := initDBWithConfig()
	return err
}

func initDBWithConfig() (*sql.DB, *config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	conn, err := database.NewConnection(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database connection manager: %w", err)
	}

	db, err := conn.GetDB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, cfg, nil
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration management",
	Long:  `Manage database migrations, including running migrations, rolling back migrations, etc.`,
}

// migrateRunCmd represents the migrate:run command
var migrateRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all pending database migrations",
	Long: `Run all pending database migrations.
This command executes all migrations that have not yet been run, updating database structure to latest state.`,
	Run: func(_ *cobra.Command, _ []string) {
		runMigrations()
	},
}

// migrateRollbackCmd represents the migrate:rollback command
var migrateRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback last database migration",
	Long: `Rollback last executed database migration.
This command undoes last migration operation, restoring database to its pre-migration state.`,
	Run: func(_ *cobra.Command, _ []string) {
		rollbackMigrations()
	},
}

// migrateResetCmd represents the migrate:reset command
var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database (delete all tables and re-run migrations)",
	Long: `Reset database, delete all tables and re-run all migrations.
Warning: This operation will delete all data, use with caution!`,
	Run: func(_ *cobra.Command, _ []string) {
		resetDatabase()
	},
}

// migrateFreshCmd represents the migrate:fresh command
var migrateFreshCmd = &cobra.Command{
	Use:   "fresh",
	Short: "Drop all tables and re-run migrations and seed data",
	Long: `Drop all tables and re-run all migrations, then run seed data.
Warning: This operation will delete all data, use with caution!`,
	Run: func(_ *cobra.Command, _ []string) {
		freshDatabase()
	},
}

// migrateStatusCmd represents the migrate:status command
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Show current status of database migrations, including run and pending migrations.`,
	Run: func(_ *cobra.Command, _ []string) {
		showMigrationStatus()
	},
}

// seedCmd represents the seed command
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Database seed data management",
	Long:  `Manage database seed data, including running seed data, clearing seed data, etc.`,
}

// seedRunCmd represents the seed:run command
var seedRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all seed data",
	Long: `Run all seed data, populating database with initial data.
This command executes all seed data creation operations.`,
	Run: func(_ *cobra.Command, _ []string) {
		runSeeds()
	},
}

// seedRunRandomCmd represents the seed:run:random command
var seedRunRandomCmd = &cobra.Command{
	Use:   "random [count]",
	Short: "Generate specified number of random test data",
	Long: `Generate specified number of random test data.
If no count is specified, defaults to 10 records.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		count := 10 // default count
		if len(args) > 0 {
			_, _ = fmt.Sscanf(args[0], "%d", &count)
		}
		runRandomSeeds(count)
	},
}

// seedClearCmd represents the seed:clear command
var seedClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all seed data",
	Long: `Clear all seed data, deleting records created by seed data.
This operation deletes all seed data but does not delete table structures.`,
	Run: func(_ *cobra.Command, _ []string) {
		clearSeeds()
	},
}

// seedRefreshCmd represents the seed:refresh command
var seedRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh seed data (clear and re-run)",
	Long: `Refresh seed data, first clear all seed data, then re-run.
This command clears existing seed data and then repopulates it.`,
	Run: func(_ *cobra.Command, _ []string) {
		refreshSeeds()
	},
}

// dbCmd represents the database command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database operation management",
	Long:  `Provide various database-related operations, including migration, seed data management, etc.`,
}

// dbSetupCmd represents the db:setup command
var dbSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup database (run migrations and seed data)",
	Long: `Setup database, run all migrations and populate seed data.
This command executes migration and seed data operations in sequence, completing database initialization.`,
	Run: func(_ *cobra.Command, _ []string) {
		setupDatabase()
	},
}

func getDefaultConnection() (config.DBConnection, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return config.DBConnection{}, fmt.Errorf("failed to load config: %w", err)
	}

	dbConfig := &cfg.Database
	defaultConn := dbConfig.Default
	connConfig, exists := dbConfig.Connections[defaultConn]
	if !exists {
		return config.DBConnection{}, fmt.Errorf("database connection config '%s' does not exist", defaultConn)
	}

	return connConfig, nil
}

func seedDialectFromConfig(cfg *config.Config) string {
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

func runAtlas(args ...string) error {
	cmd := exec.Command("atlas", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runMigrations Run database migrations
func runMigrations() {
	color.Cyan("Running database migrations...")

	connConfig, err := getDefaultConnection()
	if err != nil {
		color.Red("Failed to read database config: %v", err)
		os.Exit(1)
	}

	var envName string
	if isSQLiteDriver(connConfig.Driver) {
		envName = driverSQLite
	} else {
		envName = driverPostgres
	}

	if err := runAtlas("migrate", "apply", "--env", envName); err != nil {
		color.Red("Migration failed: %v", err)
		os.Exit(1)
	}

	color.Green("Database migration completed")
}

// rollbackMigrations Rollback database migrations
func rollbackMigrations() {
	color.Cyan("Rolling back database migrations...")

	connConfig, err := getDefaultConnection()
	if err != nil {
		color.Red("Failed to read database config: %v", err)
		os.Exit(1)
	}

	var envName string
	if isSQLiteDriver(connConfig.Driver) {
		envName = driverSQLite
	} else {
		envName = driverPostgres
	}

	if err := runAtlas("migrate", "down", "--env", envName, "1"); err != nil {
		color.Red("Rollback failed: %v", err)
		os.Exit(1)
	}

	color.Green("Database migration rollback completed")
}

// resetDatabase Reset database
func resetDatabase() {
	// Confirm operation
	fmt.Print("Warning: This will delete all data! Are you sure you want to continue? (y/N): ")
	var response string
	_, _ = fmt.Scanln(&response)

	if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
		color.Yellow("Operation cancelled")
		return
	}

	color.Cyan("Resetting database...")

	connConfig, err := getDefaultConnection()
	if err != nil {
		_, _ = color.New(color.FgRed).Printf("Failed to read database config: %v\n", err)
		os.Exit(1)
	}

	var envName string
	if isSQLiteDriver(connConfig.Driver) {
		envName = driverSQLite
		_ = os.Remove(connConfig.Database)
	} else {
		envName = driverPostgres
	}

	if err := runAtlas("migrate", "apply", "--env", envName); err != nil {
		_, _ = color.New(color.FgRed).Printf("Migration failed: %v\n", err)
		os.Exit(1)
	}

	color.Green("Database reset completed")
}

// freshDatabase Drop all tables and re-run migrations and seed data
func freshDatabase() {
	// Confirm operation
	fmt.Print("Warning: This will delete all data! Are you sure you want to continue? (y/N): ")
	var response string
	_, _ = fmt.Scanln(&response)

	if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
		color.Yellow("Operation cancelled")
		return
	}

	color.Cyan("Dropping database...")

	connConfig, err := getDefaultConnection()
	if err != nil {
		color.Red("Failed to read database config: %v", err)
		os.Exit(1)
	}

	var envName string
	if isSQLiteDriver(connConfig.Driver) {
		envName = driverSQLite
		_ = os.Remove(connConfig.Database)
	} else {
		envName = driverPostgres
	}

	if err := runAtlas("migrate", "apply", "--env", envName); err != nil {
		color.Red("Migration failed: %v", err)
		os.Exit(1)
	}

	color.Cyan("Running seed data...")
	db, cfg, err := initDBWithConfig()
	if err != nil {
		color.Red("Failed to initialize database connection: %v", err)
		os.Exit(1)
	}
	if err := seeders.RunAllSeeders(db, seedDialectFromConfig(cfg)); err != nil {
		color.Red("Failed to run seed data: %v", err)
		os.Exit(1)
	}

	color.Green("Database dropped and re-initialized completed")
}

// showMigrationStatus Show migration status
func showMigrationStatus() {
	color.Cyan("Checking migration status...")

	connConfig, err := getDefaultConnection()
	if err != nil {
		color.Red("Failed to read database config: %v", err)
		os.Exit(1)
	}

	var envName string
	if isSQLiteDriver(connConfig.Driver) {
		envName = driverSQLite
	} else {
		envName = driverPostgres
	}

	if err := runAtlas("migrate", "status", "--env", envName); err != nil {
		color.Red("Failed to get migration status: %v", err)
		os.Exit(1)
	}
}

// runSeeds Run seed data
func runSeeds() {
	db, cfg, err := initDBWithConfig()
	if err != nil {
		color.Red("Failed to initialize database connection: %v", err)
		os.Exit(1)
	}
	color.Cyan("Running seed data...")

	err = seeders.RunAllSeeders(db, seedDialectFromConfig(cfg))
	if err != nil {
		color.Red("Seed data run failed: %v", err)
		os.Exit(1)
	}

	color.Green("Seed data run completed")
}

// runRandomSeeds Generate random test data
func runRandomSeeds(count int) {
	db, cfg, err := initDBWithConfig()
	if err != nil {
		color.Red("Failed to initialize database connection: %v", err)
		os.Exit(1)
	}
	color.Cyan("Generating %d random test data records...", count)

	err = seeders.RunRandomSeeders(db, seedDialectFromConfig(cfg), count)
	if err != nil {
		color.Red("Failed to generate random data: %v", err)
		os.Exit(1)
	}

	color.Green("Successfully generated %d random test data records", count)
}

// clearSeeds Clear seed data
func clearSeeds() {
	db, cfg, err := initDBWithConfig()
	if err != nil {
		color.Red("Failed to initialize database connection: %v", err)
		os.Exit(1)
	}
	color.Cyan("Clearing seed data...")

	err = seeders.ClearAllSeeders(db, seedDialectFromConfig(cfg))
	if err != nil {
		color.Red("Failed to clear seed data: %v", err)
		os.Exit(1)
	}

	color.Green("Seed data clear completed")
}

// refreshSeeds Refresh seed data
func refreshSeeds() {
	db, cfg, err := initDBWithConfig()
	if err != nil {
		color.Red("Failed to initialize database connection: %v", err)
		os.Exit(1)
	}
	color.Cyan("Refreshing seed data...")

	err = seeders.RefreshAllSeeders(db, seedDialectFromConfig(cfg))
	if err != nil {
		color.Red("Failed to refresh seed data: %v", err)
		os.Exit(1)
	}

	color.Green("Seed data refresh completed")
}

// setupDatabase Setup database
func setupDatabase() {
	color.Cyan("Setting up database...")

	// Run migrations
	color.Yellow("Step 1/2: Running database migrations")
	connConfig, err := getDefaultConnection()
	if err != nil {
		color.Red("Failed to read database config: %v", err)
		os.Exit(1)
	}

	var envName string
	if isSQLiteDriver(connConfig.Driver) {
		envName = driverSQLite
	} else {
		envName = driverPostgres
	}

	if err := runAtlas("migrate", "apply", "--env", envName); err != nil {
		color.Red("Migration failed: %v", err)
		os.Exit(1)
	}

	// Run seed data
	color.Yellow("Step 2/2: Running seed data")
	db, cfg, err := initDBWithConfig()
	if err != nil {
		color.Red("Failed to initialize database connection: %v", err)
		os.Exit(1)
	}
	err = seeders.RunAllSeeders(db, seedDialectFromConfig(cfg))
	if err != nil {
		color.Red("Failed to run seed data: %v", err)
		os.Exit(1)
	}

	color.Green("Database setup completed")
}

// init Initialize commands
func init() {
	// Add migration subcommands
	migrateCmd.AddCommand(migrateRunCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateFreshCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	// Add seed data subcommands
	seedCmd.AddCommand(seedRunCmd)
	seedCmd.AddCommand(seedRunRandomCmd)
	seedCmd.AddCommand(seedClearCmd)
	seedCmd.AddCommand(seedRefreshCmd)

	// Add database subcommands
	dbCmd.AddCommand(migrateCmd)
	dbCmd.AddCommand(seedCmd)
	dbCmd.AddCommand(dbSetupCmd)

	// Add database commands to root command
	rootCmd.AddCommand(dbCmd)

	// Also add migration and seed commands directly to root command (backward compatibility)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(seedCmd)
}
