package tests

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lfiber/configs"
	command "lfiber/internal/console/commands"
	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/providers"
	cacheDrivers "lfiber/internal/providers/cache/drivers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := command.NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestJWTGenerateCommand_ReplacesEnvSecret(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("APP_NAME=lfiber\nJWT_SECRET=old-secret\nAPP_DEBUG=false\n"), 0o600))

	out, err := executeCommand(t, "jwt:generate", "--env", envPath)
	require.NoError(t, err)

	updated, err := os.ReadFile(envPath)
	require.NoError(t, err)
	lines := strings.Split(string(updated), "\n")
	require.Len(t, lines, 4)
	assert.Equal(t, "APP_NAME=lfiber", lines[0])
	assert.Equal(t, "APP_DEBUG=false", lines[2])
	assert.Contains(t, out, "JWT_SECRET updated in "+envPath)

	secret := strings.TrimPrefix(lines[1], "JWT_SECRET=")
	assert.NotEqual(t, "old-secret", secret)
	decoded, err := base64.StdEncoding.DecodeString(secret)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
}

func TestOldCommands_AreNotRegistered(t *testing.T) {
	oldCommands := [][]string{
		{"jwt:secret"},
		{"routes"},
		{"migrate", "run"},
		{"seed", "run"},
	}
	for _, args := range oldCommands {
		_, err := executeCommand(t, args...)
		require.Error(t, err, strings.Join(args, " "))
		assert.Contains(t, err.Error(), "unknown command")
	}
}

func TestHashCommands_MakeAndCheck(t *testing.T) {
	t.Setenv("HASH_DRIVER", "bcrypt")
	out, err := executeCommand(t, "hash:make", "secret-value")
	require.NoError(t, err)
	hashed := strings.TrimSpace(out)
	require.NotEmpty(t, hashed)

	out, err = executeCommand(t, "hash:check", "secret-value", hashed)
	require.NoError(t, err)
	assert.Contains(t, out, "Hash verified")

	_, err = executeCommand(t, "hash:check", "wrong-value", hashed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash verification failed")
}

func TestConfigShow_RedactsSensitiveValues(t *testing.T) {
	t.Setenv("JWT_SECRET", "super-secret-value")
	out, err := executeCommand(t, "config:show", "jwt.secret")
	require.NoError(t, err)
	assert.Contains(t, out, "<redacted>")
	assert.NotContains(t, out, "super-secret-value")
}

func TestCacheCommands_UseRuntimeCacheStore(t *testing.T) {
	store := cacheDrivers.NewMemoryStore("test:")
	require.NoError(t, store.Set("present", "1", 0))
	store.Wait()

	previous := commandutil.RuntimeBuilder
	commandutil.RuntimeBuilder = func() (*providers.Runtime, error) {
		return &providers.Runtime{Cache: store}, nil
	}
	t.Cleanup(func() {
		commandutil.RuntimeBuilder = previous
	})

	out, err := executeCommand(t, "cache:has", "present")
	require.NoError(t, err)
	assert.Equal(t, "true\n", out)

	out, err = executeCommand(t, "cache:forget", "present")
	require.NoError(t, err)
	assert.Contains(t, out, "Forgot cache key: present")

	out, err = executeCommand(t, "cache:clear")
	require.NoError(t, err)
	assert.Contains(t, out, "Cache cleared")
}

func TestDBSeedRandom_RejectsTooManyArgs(t *testing.T) {
	_, err := executeCommand(t, "db:seed-random", "10", "20")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts at most 1 arg")
}

func TestDBReset_CancelsWhenNoInteraction(t *testing.T) {
	previous := commandutil.AtlasRunner
	commandutil.AtlasRunner = func(args ...string) error {
		t.Fatalf("atlas should not run when destructive command is cancelled: %v", args)
		return nil
	}
	t.Cleanup(func() {
		commandutil.AtlasRunner = previous
	})

	out, err := executeCommand(t, "--no-interaction", "db:reset")
	require.NoError(t, err)
	assert.Contains(t, out, "Operation cancelled")
}

func TestDBReset_ForceSkipsConfirmation(t *testing.T) {
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_DATABASE", filepath.Join(t.TempDir(), "database.sqlite"))

	var captured [][]string
	previous := commandutil.AtlasRunner
	commandutil.AtlasRunner = func(args ...string) error {
		captured = append(captured, append([]string{}, args...))
		return nil
	}
	t.Cleanup(func() {
		commandutil.AtlasRunner = previous
	})

	out, err := executeCommand(t, "db:reset", "--force")
	require.NoError(t, err)
	assert.Contains(t, out, "Resetting database...")
	assert.Contains(t, out, "Database reset completed")
	require.Equal(t, [][]string{{"migrate", "apply", "--env", "sqlite"}}, captured)
}

func TestRootCommand_HasLaravelStyleCommands(t *testing.T) {
	out, err := executeCommand(t, "--help")
	require.NoError(t, err)
	for _, name := range []string{"app:about", "jwt:generate", "config:show", "cache:clear", "auth:about", "db:migrate", "hash:make", "schedule:run", "queue:work"} {
		assert.Contains(t, out, name)
	}
	assert.Contains(t, out, "--no-interaction")
}

func TestConfigShow_AllowsDefaultsWithoutEnvFile(t *testing.T) {
	_, _, err := configs.LoadConfig()
	require.NoError(t, err)
}
