package tests

import (
	"testing"

	command "fiber-starter/app/Console/Commands"
)

func TestQueueWorkCommandRegistered(t *testing.T) {
	rootCmd := command.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"queue:work"})
	if err != nil {
		t.Fatalf("Failed to find command: %v", err)
	}
	if cmd == nil || cmd == rootCmd {
		t.Fatalf("queue:work command not registered")
	}
	if cmd.Use != "queue:work" {
		t.Fatalf("Command Use mismatch: got=%q want=%q", cmd.Use, "queue:work")
	}
}
