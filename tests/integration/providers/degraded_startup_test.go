package providers_test

import (
	"testing"

	"fiber-starter/configs"
	providers "fiber-starter/internal/providers"
	"fiber-starter/internal/support/health"
	"fiber-starter/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviders_DegradedOptionalDependencyDoesNotFailRuntime(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	cfg.Cache.Enabled = true
	cfg.Mail.Enabled = false
	cfg.Queue.Enabled = false
	cfg.Search.Enabled = false
	cfg.Storage.Enabled = false
	cfg.Services.Dependencies = map[string]configs.ServiceDependencyConfig{
		"database": {Critical: true},
		"cache":    {Critical: false},
	}

	rt := &providers.Runtime{
		Config:     cfg,
		Connection: &testkit.StubConnection{},
		Degraded: map[string]string{
			"cache": "redis password=<redacted> unavailable",
		},
	}

	results, overall := health.NewAggregator(rt).Check()
	require.Equal(t, health.OverallDegraded, overall)
	require.Contains(t, results, "cache")
	assert.Equal(t, health.StatusDegraded, results["cache"].Status)
	assert.False(t, results["cache"].Critical)
	assert.NoError(t, rt.Close())
}
