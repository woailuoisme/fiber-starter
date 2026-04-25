package command

import "github.com/spf13/cobra"

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
	Run: func(_ *cobra.Command, _ []string) {
		setupDatabase()
	},
}
