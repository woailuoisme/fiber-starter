package providers_test

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"lfiber/configs"
	command "lfiber/internal/console/commands"
	providers "lfiber/internal/providers"
	search "lfiber/pkg/search"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestSearchProvider_NullDriver(t *testing.T) {
	cfg := &configs.Config{}
	manager := search.NewManager(cfg)
	runtime := &providers.Runtime{
		SearchManager: manager,
		SearchService: manager.Drive("null"),
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
		SearchService: manager.Drive("null"),
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
		docs := []map[string]any{
			{"id": 1, "title": "Doc 1"},
		}
		info, err = engine.AddDocuments("test_index", docs)
		require.NoError(t, err)
		assert.NotNil(t, info)

		// Search
		req := &search.SearchRequest{
			Limit: 10,
		}
		_, _ = engine.Search("test_index", "query", req)
	})
}

type mockSearchableModel struct {
	ID    int    `bun:"id,pk,autoincrement"`
	Title string `bun:"title"`
}

func (m *mockSearchableModel) SearchableId() string {
	return fmt.Sprintf("%d", m.ID)
}

func (m *mockSearchableModel) SearchableIndex() string {
	return "mocks"
}

func (m *mockSearchableModel) ToSearchableArray() map[string]any {
	return map[string]any{
		"title": m.Title,
	}
}

func TestSearchProvider_ScoutCLI(t *testing.T) {
	// 1. 初始化 Runtime
	dbPath := filepath.Join(t.TempDir(), "test_search_scout.sqlite")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_SQLITE_DATABASE", dbPath)
	t.Setenv("SEARCH_DRIVER", "null")

	rt, err := providers.Build()
	require.NoError(t, err)
	defer func() { _ = rt.Close() }()

	db, err := rt.Connection.BunDB()
	require.NoError(t, err)
	ctx := context.Background()

	// 创表和写测试数据
	_, err = db.NewCreateTable().Model((*mockSearchableModel)(nil)).Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewInsert().Model(&mockSearchableModel{ID: 1, Title: "Test Document"}).Exec(ctx)
	require.NoError(t, err)

	// 2. 注册到 Search Registry
	search.RegisterModel("mocks", &mockSearchableModel{}, func(ctx context.Context) (*bun.SelectQuery, error) {
		return db.NewSelect().Model((*mockSearchableModel)(nil)), nil
	})

	// 3. 运行 scout:import 命令行测试
	rootCmd := command.NewRootCommand()
	rootCmd.SetArgs([]string{"scout:import", "mocks"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err = rootCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Successfully imported model mocks")

	// 4. 运行 scout:flush 命令行测试
	rootCmd = command.NewRootCommand()
	rootCmd.SetArgs([]string{"scout:flush", "mocks"})

	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err = rootCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Successfully flushed all documents of model mocks")

	// 5. 运行 scout:delete-index 命令行测试
	rootCmd = command.NewRootCommand()
	rootCmd.SetArgs([]string{"scout:delete-index", "mocks"})

	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err = rootCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Successfully deleted search index of model mocks")
}
