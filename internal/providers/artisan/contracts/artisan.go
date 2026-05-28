package contracts

import "github.com/spf13/cobra"

// CommandFactory builds a fresh Cobra root command for each Artisan call.
type CommandFactory func() *cobra.Command

// Result captures the observable output of an Artisan command call.
type Result struct {
	ExitCode    int
	Output      string
	ErrorOutput string
}

// CommandInfo describes a registered console command.
type CommandInfo struct {
	Name        string
	Description string
	GroupID     string
	Available   bool
}

// Artisan defines the programmatic console command runner contract.
type Artisan interface {
	Call(command string, args ...string) (Result, error)
	List() []CommandInfo
	Has(command string) bool
}
