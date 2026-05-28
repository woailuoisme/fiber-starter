package providers_test

import (
	"fmt"
	"testing"

	artisan "lfiber/internal/providers/artisan"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtisanService_CallCapturesOutput(t *testing.T) {
	service, err := artisan.Register(testArtisanRoot)
	require.NoError(t, err)

	result, err := service.Call("demo:ok", "alpha")
	require.NoError(t, err)

	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "ok alpha\n", result.Output)
	assert.Empty(t, result.ErrorOutput)
}

func TestArtisanService_CallUnknownCommandReturnsResultAndError(t *testing.T) {
	service, err := artisan.Register(testArtisanRoot)
	require.NoError(t, err)

	result, err := service.Call("missing")
	require.Error(t, err)

	assert.Equal(t, 1, result.ExitCode)
	assert.Empty(t, result.Output)
}

func TestArtisanService_ListAndHasCommands(t *testing.T) {
	service, err := artisan.Register(testArtisanRoot)
	require.NoError(t, err)

	assert.True(t, service.Has("demo:ok"))
	assert.False(t, service.Has("missing"))

	commands := service.List()
	require.Len(t, commands, 1)
	assert.Equal(t, "demo:ok", commands[0].Name)
	assert.Equal(t, "Demo command", commands[0].Description)
	assert.Equal(t, "system", commands[0].GroupID)
	assert.True(t, commands[0].Available)
}

func TestArtisanService_MissingFactoryReturnsError(t *testing.T) {
	previous := artisan.DefaultCommandFactory()
	artisan.SetCommandFactory(nil)
	t.Cleanup(func() {
		artisan.SetCommandFactory(previous)
	})

	service, err := artisan.Register()
	require.NoError(t, err)

	result, err := service.Call("demo:ok")
	require.ErrorIs(t, err, artisan.ErrCommandFactoryNotConfigured)
	assert.Equal(t, 1, result.ExitCode)
	assert.Nil(t, service.List())
	assert.False(t, service.Has("demo:ok"))
}

func testArtisanRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "lfiber",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddGroup(&cobra.Group{ID: "system", Title: "System Commands"})
	root.AddCommand(&cobra.Command{
		Use:     "demo:ok <value>",
		Short:   "Demo command",
		GroupID: "system",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "ok %s\n", args[0])
			return err
		},
	})
	return root
}
