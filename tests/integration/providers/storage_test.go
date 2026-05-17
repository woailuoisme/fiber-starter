package providers_test

import (
	"testing"

	providers "fiber-starter/app/Providers"
	storage "fiber-starter/app/Providers/Storage"
	storageContracts "fiber-starter/app/Providers/Storage/Contracts"
	"fiber-starter/configs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageProvider(t *testing.T) {
	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)

	t.Run("LocalDisk", func(t *testing.T) {
		// Override local root for testing
		tmpRoot := t.TempDir()
		cfg.Storage.Local = &configs.LocalStorageConfig{Root: tmpRoot, URL: "/storage"}

		runtime := &providers.Runtime{
			Storage: storage.NewManager(cfg),
		}
		providers.SetInstance(runtime)
		defer func() {
			_ = runtime.Close()
		}()

		disk := storage.Drive("local")
		require.NotNil(t, disk)

		testStorageOperations(t, disk)
	})

	t.Run("PublicDisk", func(t *testing.T) {
		tmpRoot := t.TempDir()
		cfg.Storage.Public = &configs.LocalStorageConfig{Root: tmpRoot, URL: "/storage"}

		runtime := &providers.Runtime{
			Storage: storage.NewManager(cfg),
		}
		providers.SetInstance(runtime)
		defer func() {
			_ = runtime.Close()
		}()

		disk := storage.Drive("public")
		require.NotNil(t, disk)

		testStorageOperations(t, disk)

		// Test URL generation
		url := disk.Url("test.png")
		assert.Contains(t, url, "/storage")
		assert.Contains(t, url, "test.png")
	})
}

func testStorageOperations(t *testing.T, disk storageContracts.Disk) {
	t.Helper()

	path := "test/file.txt"
	content := []byte("hello world")

	// Put
	err := disk.Put(path, content)
	require.NoError(t, err)

	// Missing
	missing, err := disk.Missing("non_existent.txt")
	require.NoError(t, err)
	assert.True(t, missing)

	// Copy
	copyPath := "test/copy.txt"
	err = disk.Copy(path, copyPath)
	require.NoError(t, err)
	exists, _ := disk.Exists(copyPath)
	assert.True(t, exists)

	// Move
	movePath := "test/moved.txt"
	err = disk.Move(copyPath, movePath)
	require.NoError(t, err)
	exists, _ = disk.Exists(movePath)
	assert.True(t, exists)
	exists, _ = disk.Exists(copyPath)
	assert.False(t, exists)

	// Prepend/Append
	err = disk.Put("test/append.txt", []byte("middle"))
	require.NoError(t, err)
	disk.Prepend("test/append.txt", []byte("start_"))
	disk.Append("test/append.txt", []byte("_end"))
	got, _ := disk.Get("test/append.txt")
	assert.Equal(t, []byte("start_middle_end"), got)

	// Directory Operations
	dirPath := "test/new_dir"
	err = disk.MakeDirectory(dirPath)
	require.NoError(t, err)

	err = disk.Put(dirPath+"/file.txt", []byte("in dir"))
	require.NoError(t, err)

	files, err := disk.Files("test")
	require.NoError(t, err)
	assert.NotEmpty(t, files)

	// Visibility
	err = disk.SetVisibility(path, "private")
	require.NoError(t, err)
	vis, err := disk.GetVisibility(path)
	require.NoError(t, err)
	assert.Equal(t, "private", vis)

	// Delete Directory
	err = disk.DeleteDirectory(dirPath)
	require.NoError(t, err)
	exists, _ = disk.Exists(dirPath + "/file.txt")
	assert.False(t, exists)

	// Final Cleanup
	err = disk.Delete(path, "test/append.txt", movePath)
	require.NoError(t, err)
}
