package providers_test

import (
	"testing"
	"time"

	"fiber-starter/configs"
	storageDrivers "fiber-starter/internal/providers/storage/drivers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3Driver_BasicOperationsAndUnsupportedPaths(t *testing.T) {
	cfg := &configs.Config{
		Storage: configs.StorageConfig{
			S3: &configs.S3StorageConfig{
				Bucket:          "bucket",
				Endpoint:        "http://127.0.0.1:1",
				Region:          "us-east-1",
				AccessKeyID:     "access",
				SecretAccessKey: "secret",
			},
		},
	}

	driver := storageDrivers.NewS3Driver(cfg, "s3")
	require.NotNil(t, driver)

	assert.Equal(t, "/storage/path.txt", driver.Url("path.txt"))
	tmpURL, err := driver.TemporaryUrl("path.txt", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, "/storage/path.txt", tmpURL)
	files, err := driver.Files("dir")
	require.Error(t, err)
	assert.Nil(t, files)
	files, err = driver.AllFiles("dir")
	require.Error(t, err)
	assert.Nil(t, files)
	dirs, err := driver.Directories("dir")
	require.Error(t, err)
	assert.Nil(t, dirs)
	dirs, err = driver.AllDirectories("dir")
	require.Error(t, err)
	assert.Nil(t, dirs)
	require.NoError(t, driver.MakeDirectory("dir"))
	require.NoError(t, driver.SetVisibility("path.txt", "private"))
	visibility, err := driver.GetVisibility("path.txt")
	require.NoError(t, err)
	assert.Equal(t, "public", visibility)
	modified, err := driver.LastModified("path.txt")
	require.NoError(t, err)
	assert.False(t, modified.IsZero())
	require.Error(t, driver.DeleteDirectory("dir"))

	_, err = driver.Get("missing")
	require.Error(t, err)
	require.Error(t, driver.Put("key", []byte("value")))
	exists, err := driver.Exists("missing")
	require.NoError(t, err)
	assert.False(t, exists)
	missing, err := driver.Missing("missing")
	require.NoError(t, err)
	assert.True(t, missing)
	_, err = driver.Size("missing")
	require.Error(t, err)
	require.Error(t, driver.Copy("missing", "copy"))
	require.Error(t, driver.Move("missing", "move"))
	require.Error(t, driver.Prepend("missing", []byte("pre")))
	require.Error(t, driver.Append("missing", []byte("post")))
	require.Error(t, driver.Delete("missing"))
	require.Error(t, driver.Reset())
	require.NoError(t, driver.Close())
}

func TestS3Driver_GarageConfig(t *testing.T) {
	t.Run("UseSSL", func(t *testing.T) {
		cfg := &configs.Config{
			Storage: configs.StorageConfig{
				Garage: &configs.GarageStorageConfig{
					Bucket:          "bucket",
					Endpoint:        "garage.local",
					Region:          "us-east-1",
					AccessKeyID:     "access",
					SecretAccessKey: "secret",
					UseSSL:          true,
				},
			},
		}

		driver := storageDrivers.NewS3Driver(cfg, "garage")
		require.NotNil(t, driver)
		assert.Equal(t, "/storage/file.txt", driver.Url("file.txt"))
	})

	t.Run("NoSSL", func(t *testing.T) {
		cfg := &configs.Config{
			Storage: configs.StorageConfig{
				Garage: &configs.GarageStorageConfig{
					Bucket:          "bucket",
					Endpoint:        "garage.local",
					Region:          "us-east-1",
					AccessKeyID:     "access",
					SecretAccessKey: "secret",
					UseSSL:          false,
				},
			},
		}

		driver := storageDrivers.NewS3Driver(cfg, "garage")
		require.NotNil(t, driver)
		assert.Equal(t, "/storage/file.txt", driver.Url("file.txt"))
	})
}
