package tests

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	command "fiber-starter/internal/console/commands"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTSecretCommand_ReplacesEnvSecret(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("APP_NAME=Fiber Starter\nJWT_SECRET=old-secret\nAPP_DEBUG=false\n"), 0o600))

	previousDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() {
		_ = os.Chdir(previousDir)
	})

	root := command.GetRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"jwt:secret", "--env", ".env"})
	t.Cleanup(func() {
		root.SetArgs(nil)
		root.SetOut(os.Stdout)
		root.SetErr(os.Stderr)
	})

	require.NoError(t, root.Execute())

	updated, err := os.ReadFile(envPath)
	require.NoError(t, err)
	lines := strings.Split(string(updated), "\n")
	require.Len(t, lines, 4)
	assert.Equal(t, "APP_NAME=Fiber Starter", lines[0])
	assert.Equal(t, "APP_DEBUG=false", lines[2])
	assert.Contains(t, out.String(), "JWT_SECRET updated in .env")

	secret := strings.TrimPrefix(lines[1], "JWT_SECRET=")
	assert.NotEqual(t, "old-secret", secret)
	decoded, err := base64.StdEncoding.DecodeString(secret)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
}
