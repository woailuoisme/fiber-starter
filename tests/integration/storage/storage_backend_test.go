package tests

import (
	"testing"
	"time"

	services "fiber-starter/app/Services"
	"fiber-starter/config"
	"fiber-starter/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageServiceInitialization_S3CompatibleDrivers(t *testing.T) {
	cases := []struct {
		name     string
		cfg      *config.StorageConfig
		redisCfg *config.RedisConfig
	}{
		{
			name: "garage",
			cfg: &config.StorageConfig{
				Driver: "garage",
				Garage: &config.GarageStorageConfig{
					Endpoint:        "garage.local:3900",
					AccessKeyID:     "garage-access",
					SecretAccessKey: "garage-secret",
					UseSSL:          false,
					Bucket:          "garage-bucket",
					Region:          "us-east-1",
				},
			},
			redisCfg: testkit.DefaultRedisConfig(),
		},
		{
			name: "minio alias",
			cfg: &config.StorageConfig{
				Driver: "minio",
				MinIO: &config.GarageStorageConfig{
					Endpoint:        "minio.local:9000",
					AccessKeyID:     "minio-access",
					SecretAccessKey: "minio-secret",
					UseSSL:          false,
					Bucket:          "minio-bucket",
					Region:          "us-east-1",
				},
			},
			redisCfg: testkit.DefaultRedisConfig(),
		},
		{
			name: "generic s3",
			cfg: &config.StorageConfig{
				Driver: "s3",
				S3: &config.S3StorageConfig{
					AccessKeyID:     "s3-access",
					SecretAccessKey: "s3-secret",
					Region:          "us-east-1",
					Bucket:          "s3-bucket",
					Endpoint:        "https://s3.amazonaws.com",
				},
			},
			redisCfg: testkit.DefaultRedisConfig(),
		},
		{
			name: "r2",
			cfg: &config.StorageConfig{
				Driver: "r2",
				R2: &config.S3StorageConfig{
					AccessKeyID:     "r2-access",
					SecretAccessKey: "r2-secret",
					Region:          "auto",
					Bucket:          "r2-bucket",
					Endpoint:        "account-id.r2.cloudflarestorage.com",
				},
			},
			redisCfg: testkit.DefaultRedisConfig(),
		},
		{
			name: "oss",
			cfg: &config.StorageConfig{
				Driver: "oss",
				OSS: &config.S3StorageConfig{
					AccessKeyID:     "oss-access",
					SecretAccessKey: "oss-secret",
					Region:          "cn-hangzhou",
					Bucket:          "oss-bucket",
					Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
				},
			},
			redisCfg: testkit.DefaultRedisConfig(),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, err := services.NewStorageService(tc.cfg, tc.redisCfg)
			require.NoError(t, err)
			assert.NotNil(t, svc.GetStorage())
			require.NoError(t, svc.Close())
		})
	}
}

func TestStorageServiceRejectsIncompleteS3Configs(t *testing.T) {
	cases := []struct {
		name string
		cfg  *config.StorageConfig
	}{
		{
			name: "garage missing endpoint",
			cfg: &config.StorageConfig{
				Driver: "garage",
				Garage: &config.GarageStorageConfig{
					Bucket: "garage-bucket",
					Region: "us-east-1",
				},
			},
		},
		{
			name: "r2 missing bucket",
			cfg: &config.StorageConfig{
				Driver: "r2",
				R2: &config.S3StorageConfig{
					Region:   "auto",
					Endpoint: "account-id.r2.cloudflarestorage.com",
				},
			},
		},
		{
			name: "oss missing region",
			cfg: &config.StorageConfig{
				Driver: "oss",
				OSS: &config.S3StorageConfig{
					Bucket:   "oss-bucket",
					Endpoint: "oss-cn-hangzhou.aliyuncs.com",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, err := services.NewStorageService(tc.cfg, testkit.DefaultRedisConfig())
			require.NoError(t, err)
			err = svc.Set("probe", []byte("value"), time.Second)
			require.Error(t, err)
			_ = svc.Close()
		})
	}
}
