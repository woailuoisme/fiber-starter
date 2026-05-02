package services

import (
	"testing"

	"fiber-starter/config"
)

func TestNormalizeStorageDriver(t *testing.T) {
	t.Parallel()

	if got := normalizeStorageDriver(" MINIO "); got != "garage" {
		t.Fatalf("expected minio to normalize to garage, got %q", got)
	}

	if got := normalizeStorageDriver("r2"); got != "r2" {
		t.Fatalf("expected r2 to stay unchanged, got %q", got)
	}
}

func TestNormalizeS3Endpoint(t *testing.T) {
	t.Parallel()

	if got := normalizeS3Endpoint("localhost:9000", "http", true); got != "http://localhost:9000" {
		t.Fatalf("expected http endpoint, got %q", got)
	}

	if got := normalizeS3Endpoint("https://example.com", "http", true); got != "https://example.com" {
		t.Fatalf("expected pre-schemed endpoint to stay unchanged, got %q", got)
	}

	if got := normalizeS3Endpoint("localhost:9000", "http", false); got != "localhost:9000" {
		t.Fatalf("expected normalization to be skipped, got %q", got)
	}
}

func TestBuildStorageSupportsS3CompatibleProviders(t *testing.T) {
	t.Parallel()

	redisCfg := &config.RedisConfig{
		Host: "localhost",
		Port: "6379",
		DB:   0,
	}

	cases := []struct {
		name       string
		cfg        *config.StorageConfig
		wantDriver string
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
			wantDriver: "garage",
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
			wantDriver: "garage",
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
			wantDriver: "s3",
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
			wantDriver: "r2",
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
			wantDriver: "oss",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store, driver, err := buildStorage(tc.cfg, redisCfg)
			if err != nil {
				t.Fatalf("buildStorage returned error: %v", err)
			}
			if driver != tc.wantDriver {
				t.Fatalf("expected driver %q, got %q", tc.wantDriver, driver)
			}
			if store == nil {
				t.Fatal("expected storage backend to be initialized")
			}
			if err := store.Close(); err != nil {
				t.Fatalf("close returned error: %v", err)
			}
		})
	}
}

func TestBuildStorageRejectsIncompleteS3Configs(t *testing.T) {
	t.Parallel()

	redisCfg := &config.RedisConfig{
		Host: "localhost",
		Port: "6379",
		DB:   0,
	}

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
			t.Parallel()

			store, driver, err := buildStorage(tc.cfg, redisCfg)
			if err == nil {
				if store != nil {
					_ = store.Close()
				}
				t.Fatalf("expected buildStorage to fail for driver %q", driver)
			}
		})
	}
}
