package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fiber-starter",
	Short: "Fiber Starter application command-line tool",
	Long: `Fiber Starter is an application starter based on the Go Fiber framework.
This command-line tool provides various useful features including key generation, database management, and more.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error executing command: '%s'", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

// GetRootCmd returns the root command
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is .env)")
}
