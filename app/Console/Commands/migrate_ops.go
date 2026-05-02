package command

import (
	"database/sql"
	"os"

	"fiber-starter/database/seeders"

	"github.com/fatih/color"
)

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

	rebuildDatabase("Dropping database...", true)
	color.Green("Database dropped and re-initialized completed")
}

func showMigrationStatus() {
	color.Cyan("Checking migration status...")
	if err := runAtlasForCurrentConnection("migrate", "status"); err != nil {
		color.Red("Failed to get migration status: %v", err)
		os.Exit(1)
	}
}

func runSeeds() {
	color.Cyan("Running seed data...")
	runSeedOperation("Seed data run failed", seeders.RunAllSeeders)

	color.Green("Seed data run completed")
}

func runRandomSeeds(count int) {
	color.Cyan("Generating %d random test data records...", count)
	runSeedOperation("Failed to generate random data", func(db *sql.DB, dialect string) error {
		return seeders.RunRandomSeeders(db, dialect, count)
	})

	color.Green("Successfully generated %d random test data records", count)
}

func clearSeeds() {
	color.Cyan("Clearing seed data...")
	runSeedOperation("Failed to clear seed data", seeders.ClearAllSeeders)

	color.Green("Seed data clear completed")
}

func refreshSeeds() {
	color.Cyan("Refreshing seed data...")
	runSeedOperation("Failed to refresh seed data", seeders.RefreshAllSeeders)

	color.Green("Seed data refresh completed")
}

func setupDatabase() {
	color.Cyan("Setting up database...")
	color.Yellow("Step 1/2: Running database migrations")
	if err := runAtlasForCurrentConnection("migrate", "apply"); err != nil {
		color.Red("Migration failed: %v", err)
		os.Exit(1)
	}

	color.Yellow("Step 2/2: Running seed data")
	runSeedOperation("Failed to run seed data", seeders.RunAllSeeders)

	color.Green("Database setup completed")
}

func init() {
	migrateCmd.AddCommand(migrateRunCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateFreshCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	seedCmd.AddCommand(seedRunCmd)
	seedCmd.AddCommand(seedRunRandomCmd)
	seedCmd.AddCommand(seedClearCmd)
	seedCmd.AddCommand(seedRefreshCmd)

	dbCmd.AddCommand(migrateCmd)
	dbCmd.AddCommand(seedCmd)
	dbCmd.AddCommand(dbSetupCmd)

	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(seedCmd)
}
