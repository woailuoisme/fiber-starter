package providers_test

import (
	"testing"
	"time"

	"lfiber/configs"
	cache "lfiber/internal/providers/cache"
	"lfiber/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheProvider(t *testing.T) {
	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)

	// Use memory for testing to avoid dependency on Redis
	cfg.Cache.Driver = "memory"

	t.Run("FullStoreOperations", func(t *testing.T) {
		_, store, err := cache.RegisterCache(cfg)
		require.NoError(t, err)
		require.NotNil(t, store)

		key := "cache_op_key"
		val := "cache_op_value"

		// Put
		err = store.Put(key, val, 1*time.Minute)
		require.NoError(t, err)

		// Wait for eventually consistent set if memory driver
		if m, ok := store.(interface{ Wait() }); ok {
			m.Wait()
		}

		// Has
		has, err := store.Has(key)
		require.NoError(t, err)
		assert.True(t, has)

		// Get
		got, err := store.Get(key)
		require.NoError(t, err)
		assert.Equal(t, val, got)

		// Pull
		pulled, err := store.Pull(key)
		require.NoError(t, err)
		assert.Equal(t, val, pulled)

		has, _ = store.Has(key)
		assert.False(t, has, "Pull should remove the item")

		// Add
		added, err := store.Add(key, val, 1*time.Minute)
		require.NoError(t, err)
		assert.True(t, added)

		if m, ok := store.(interface{ Wait() }); ok {
			m.Wait()
		}

		added, _ = store.Add(key, "new_val", 1*time.Minute)
		assert.False(t, added, "Add should fail if key exists")

		// Forever
		err = store.Forever("forever_key", "forever_val")
		require.NoError(t, err)

		if m, ok := store.(interface{ Wait() }); ok {
			m.Wait()
		}

		exists, _ := store.Has("forever_key")
		assert.True(t, exists)

		// Forget
		err = store.Forget("forever_key")
		require.NoError(t, err)
		exists, _ = store.Has("forever_key")
		assert.False(t, exists)

		// Flush
		store.Put("k1", "v1", 1*time.Minute)
		store.Put("k2", "v2", 1*time.Minute)
		err = store.Flush()
		require.NoError(t, err)
		h1, _ := store.Has("k1")
		h2, _ := store.Has("k2")
		assert.False(t, h1)
		assert.False(t, h2)
	})

	t.Run("JSONOperations", func(t *testing.T) {
		_, store, _ := cache.RegisterCache(cfg)

		type Data struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		key := "test_json_key"
		val := Data{Name: "Antigravity", Age: 1}

		err = store.Set(key, val, 1*time.Minute)
		require.NoError(t, err)

		// Wait for eventually consistent set if memory driver
		if m, ok := store.(interface{ Wait() }); ok {
			m.Wait()
		}

		var got Data
		err = store.GetJSON(key, &got)
		require.NoError(t, err)
		assert.Equal(t, val, got)
	})
}

func TestCacheManager(t *testing.T) {
	cfg := testkit.DefaultConfig() // Use a clean config

	t.Run("LazyLoading", func(t *testing.T) {
		// Mock config to use memory
		cfg.Cache.Driver = "memory"

		_, store, err := cache.RegisterCache(cfg)
		require.NoError(t, err)
		assert.NotNil(t, store)
	})
}
