package command

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "fiber-starter",
	Short: "Fiber Starter application command-line tool",
	Long: `Fiber Starter is an application starter based on the Go Fiber framework.
This command-line tool provides various useful features including key generation, database management, and more.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exitWithError("Error executing command: %v", err)
	}
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is .env)")
}
