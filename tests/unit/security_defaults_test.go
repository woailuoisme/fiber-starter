package tests

import (
	"os"
	"path/filepath"
	"testing"

	"lfiber/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityDefaults_AreSafeByDefault(t *testing.T) {
	root := testkit.RepoRoot(t)

	appYAML, err := os.ReadFile(filepath.Join(root, "configs/yml/app.yaml"))
	require.NoError(t, err)
	authYAML, err := os.ReadFile(filepath.Join(root, "configs/yml/auth.yaml"))
	require.NoError(t, err)
	envExample, err := os.ReadFile(filepath.Join(root, ".env.example"))
	require.NoError(t, err)
	servicesYAML, err := os.ReadFile(filepath.Join(root, "configs/yml/services.yaml"))
	require.NoError(t, err)

	assert.Contains(t, string(appYAML), "debug: ${APP_DEBUG:false}")
	assert.Contains(t, string(appYAML), "trust_proxy: false")
	assert.Contains(t, string(appYAML), "body_limit: 4194304")
	assert.Contains(t, string(authYAML), `secret: ${JWT_SECRET:""}`)
	assert.Contains(t, string(envExample), "APP_DEBUG=false")
	assert.Contains(t, string(envExample), "JWT_SECRET=")
	assert.NotContains(t, string(envExample), "change-me")
	assert.Contains(t, string(servicesYAML), "SERVICE_DATABASE_CRITICAL:true")
	assert.Contains(t, string(servicesYAML), "SERVICE_CACHE_CRITICAL:false")
}
