package providers

import (
	"testing"

	configs "lfiber/configs"
	config "lfiber/internal/providers/config"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigProvider(t *testing.T) {
	k := koanf.New(".")
	err := k.Load(confmap.Provider(map[string]interface{}{
		"app.name":                        "Test App",
		"app.debug":                       true,
		"database.connections.mysql.host": "localhost",
		"database.connections.mysql.port": 3306,
	}, "."), nil)
	require.NoError(t, err)

	repo := config.NewRepository(k)

	// Test Get
	assert.Equal(t, "Test App", repo.Get("app.name"))
	assert.Equal(t, true, repo.Get("app.debug"))
	assert.Equal(t, "localhost", repo.Get("database.connections.mysql.host"))
	assert.Equal(t, 3306, repo.Get("database.connections.mysql.port"))

	// Test Default Value
	assert.Equal(t, "Default", repo.Get("non.existent", "Default"))

	// Test Has
	assert.True(t, repo.Has("app.name"))
	assert.False(t, repo.Has("app.version"))

	// Test Typed Gets
	assert.Equal(t, "Test App", repo.GetString("app.name"))
	assert.True(t, repo.GetBool("app.debug"))
	assert.Equal(t, 3306, repo.GetInt("database.connections.mysql.port"))

	// Test Typed Default Values
	assert.Equal(t, "Fallback", repo.GetString("non.existent", "Fallback"))
	assert.Equal(t, 123, repo.GetInt("non.existent", 123))
	assert.False(t, repo.GetBool("non.existent", false))

	// Test Set
	repo.Set("app.version", "1.0.0")
	assert.Equal(t, "1.0.0", repo.GetString("app.version"))

	// Test Prepend
	repo.Set("app.tags", []interface{}{"web", "api"})
	repo.Prepend("app.tags", "framework")
	tags := repo.Get("app.tags").([]interface{})
	assert.Equal(t, "framework", tags[0])
	assert.Len(t, tags, 3)

	// Test Push
	repo.Push("app.tags", "starter")
	tags = repo.Get("app.tags").([]interface{})
	assert.Equal(t, "starter", tags[len(tags)-1])
	assert.Len(t, tags, 4)

	assert.InDelta(t, 0.0, repo.GetFloat64("non.existent.float"), 0.0001)
	assert.InDelta(t, 1.5, repo.GetFloat64("non.existent.float", 1.5), 0.0001)

	all := repo.All()
	assert.Contains(t, all, "app.name")
	assert.Contains(t, all, "app.debug")
	assert.Contains(t, all, "database.connections.mysql.host")

	var dbCfg configs.DatabaseConfig
	require.NoError(t, repo.LoadDatabaseConfig(&dbCfg))
	assert.Equal(t, "localhost", dbCfg.Connections["mysql"].Host)
}
