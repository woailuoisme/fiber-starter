package tests

import (
	"testing"

	command "fiber-starter/app/Console/Commands"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueWorkCommandRegistered(t *testing.T) {
	rootCmd := command.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"queue:work"})
	require.NoError(t, err)
	require.NotNil(t, cmd)
	assert.NotSame(t, rootCmd, cmd)
	assert.Equal(t, "queue:work", cmd.Use)
}
