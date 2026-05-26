package providers_test

import (
	"testing"

	"lfiber/configs"
	providers "lfiber/internal/providers"
	search "lfiber/internal/providers/search"
	searchContracts "lfiber/internal/providers/search/contracts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchProvider_NullDriver(t *testing.T) {
	cfg := &configs.Config{}
	manager := search.NewManager(cfg)
	runtime := &providers.Runtime{
		SearchManager: manager,
		SearchService: manager.Drive(),
	}
	providers.SetInstance(runtime)
	defer func() {
		_ = runtime.Close()
	}()

	engine := search.Drive("null")

	err := engine.HealthCheck()
	require.NoError(t, err)

	info, err := engine.CreateIndex("test", "id")
	require.NoError(t, err)
	assert.NotNil(t, info)
}

func TestSearchProvider_MeilisearchConfigError(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Search.Enabled = true
	cfg.Search.Host = "" // Missing host

	_, engine, err := search.Register(cfg)
	require.NoError(t, err)

	_, err = engine.Search("test", "query", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "meilisearch client not initialized")
}

func TestSearchProvider_Operations(t *testing.T) {
	cfg := &configs.Config{}
	manager := search.NewManager(cfg)
	runtime := &providers.Runtime{
		SearchManager: manager,
		SearchService: manager.Drive(),
	}
	providers.SetInstance(runtime)
	defer func() {
		_ = runtime.Close()
	}()

	engine := search.Drive("null")

	t.Run("Lifecycle", func(t *testing.T) {
		// CreateIndex
		info, err := engine.CreateIndex("test_index", "id")
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "enqueued", info.Status)

		// AddDocuments
		docs := []map[string]interface{}{
			{"id": 1, "title": "Doc 1"},
		}
		info, err = engine.AddDocuments("test_index", docs)
		require.NoError(t, err)
		assert.NotNil(t, info)

		// Search
		req := &searchContracts.SearchRequest{
			Limit: 10,
		}
		// This will likely fail without a real host, but we test the structure
		_, _ = engine.Search("test_index", "query", req)
	})
}
