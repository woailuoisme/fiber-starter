package services

import (
	"fmt"
	"strings"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/gofiber/storage"
	redisStorage "github.com/gofiber/storage/redis/v3"
	fiberS3 "github.com/gofiber/storage/s3"
	"go.uber.org/zap"
)

func buildStorage(cfg *config.StorageConfig, redisCfg *config.RedisConfig) (storage.Storage, string, error) {
	driver := normalizeStorageDriver(cfg.Driver)

	switch driver {
	case "redis":
		return createRedisStorage(cfg, redisCfg), "redis", nil
	case "garage":
		return createGarageStorage(cfg)
	case "s3":
		return createS3Storage(cfg)
	case "r2":
		return createR2Storage(cfg)
	case "oss":
		return createOSSStorage(cfg)
	default:
		return createDefaultStorage(cfg, redisCfg), "redis", nil
	}
}

func normalizeStorageDriver(driver string) string {
	driver = strings.ToLower(strings.TrimSpace(driver))
	if driver == "minio" {
		return "garage"
	}
	return driver
}

func createRedisStorage(cfg *config.StorageConfig, redisCfg *config.RedisConfig) storage.Storage {
	url := redisURL(redisCfg)
	return redisStorage.New(redisStorage.Config{
		URL:   url,
		Reset: cfg.Reset,
	})
}

func createGarageStorage(cfg *config.StorageConfig) (storage.Storage, string, error) {
	garageCfg, err := storageCompatConfig("garage", cfg.Garage, cfg.MinIO)
	if err != nil {
		return nil, "", err
	}

	s3Cfg, err := buildS3ConfigFromGarage("garage", garageCfg, cfg.Reset)
	if err != nil {
		return nil, "", err
	}

	store, err := newFiberS3Storage("garage", s3Cfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize garage storage: %w", err)
	}

	return store, "garage", nil
}

func createS3Storage(cfg *config.StorageConfig) (storage.Storage, string, error) {
	s3Cfg, err := buildS3Config("s3", cfg.S3, cfg.Reset, false)
	if err != nil {
		return nil, "", err
	}

	store, err := newFiberS3Storage("s3", s3Cfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize s3 storage: %w", err)
	}

	return store, "s3", nil
}

func createR2Storage(cfg *config.StorageConfig) (storage.Storage, string, error) {
	s3Cfg, err := buildS3Config("r2", cfg.R2, cfg.Reset, true)
	if err != nil {
		return nil, "", err
	}

	store, err := newFiberS3Storage("r2", s3Cfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize r2 storage: %w", err)
	}

	return store, "r2", nil
}

func createOSSStorage(cfg *config.StorageConfig) (storage.Storage, string, error) {
	s3Cfg, err := buildS3Config("oss", cfg.OSS, cfg.Reset, true)
	if err != nil {
		return nil, "", err
	}

	store, err := newFiberS3Storage("oss", s3Cfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize oss storage: %w", err)
	}

	return store, "oss", nil
}

func createDefaultStorage(cfg *config.StorageConfig, redisCfg *config.RedisConfig) storage.Storage {
	helpers.Warn("Unknown storage driver, trying to use Redis storage as default", zap.String("driver", cfg.Driver))
	return createRedisStorage(cfg, redisCfg)
}

func buildS3Config(name string, cfg *config.S3StorageConfig, reset bool, normalizeEndpoint bool) (fiberS3.Config, error) {
	if cfg == nil {
		return fiberS3.Config{}, fmt.Errorf("%s config cannot be empty", name)
	}

	bucket := strings.TrimSpace(cfg.Bucket)
	region := strings.TrimSpace(cfg.Region)
	endpoint := strings.TrimSpace(cfg.Endpoint)

	if bucket == "" {
		return fiberS3.Config{}, fmt.Errorf("%s bucket cannot be empty", name)
	}
	if region == "" {
		return fiberS3.Config{}, fmt.Errorf("%s region cannot be empty", name)
	}
	if normalizeEndpoint && endpoint == "" {
		return fiberS3.Config{}, fmt.Errorf("%s endpoint cannot be empty", name)
	}

	endpoint = normalizeS3Endpoint(endpoint, "https", normalizeEndpoint)
	credentials, err := buildS3Credentials(name, cfg.AccessKeyID, cfg.SecretAccessKey)
	if err != nil {
		return fiberS3.Config{}, err
	}

	return fiberS3.Config{
		Bucket:         bucket,
		Endpoint:       endpoint,
		Region:         region,
		Reset:          reset,
		Credentials:    credentials,
		MaxAttempts:    3,
		RequestTimeout: 0,
	}, nil
}

func buildS3ConfigFromGarage(name string, cfg *config.GarageStorageConfig, reset bool) (fiberS3.Config, error) {
	if cfg == nil {
		return fiberS3.Config{}, fmt.Errorf("%s config cannot be empty", name)
	}

	bucket := strings.TrimSpace(cfg.Bucket)
	region := strings.TrimSpace(cfg.Region)
	endpoint := strings.TrimSpace(cfg.Endpoint)

	if bucket == "" {
		return fiberS3.Config{}, fmt.Errorf("%s bucket cannot be empty", name)
	}
	if region == "" {
		return fiberS3.Config{}, fmt.Errorf("%s region cannot be empty", name)
	}
	if endpoint == "" {
		return fiberS3.Config{}, fmt.Errorf("%s endpoint cannot be empty", name)
	}

	endpoint = normalizeS3Endpoint(endpoint, map[bool]string{true: "https", false: "http"}[cfg.UseSSL], true)
	credentials, err := buildS3Credentials(name, cfg.AccessKeyID, cfg.SecretAccessKey)
	if err != nil {
		return fiberS3.Config{}, err
	}

	return fiberS3.Config{
		Bucket:      bucket,
		Endpoint:    endpoint,
		Region:      region,
		Reset:       reset,
		Credentials: credentials,
		MaxAttempts: 3,
	}, nil
}

func buildS3Credentials(name, accessKeyID, secretAccessKey string) (fiberS3.Credentials, error) {
	accessKeyID = strings.TrimSpace(accessKeyID)
	secretAccessKey = strings.TrimSpace(secretAccessKey)

	switch {
	case accessKeyID == "" && secretAccessKey == "":
		return fiberS3.Credentials{}, nil
	case accessKeyID == "" || secretAccessKey == "":
		return fiberS3.Credentials{}, fmt.Errorf("%s credentials require both access key and secret access key", name)
	default:
		return fiberS3.Credentials{
			AccessKey:       accessKeyID,
			SecretAccessKey: secretAccessKey,
		}, nil
	}
}

func normalizeS3Endpoint(endpoint, scheme string, normalize bool) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" || !normalize {
		return endpoint
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	if scheme == "" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, endpoint)
}

func newFiberS3Storage(name string, cfg fiberS3.Config) (store storage.Storage, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%s storage init panicked: %v", name, recovered)
		}
	}()

	store = fiberS3.New(cfg)
	return store, nil
}

func storageCompatConfig(name string, primary, fallback *config.GarageStorageConfig) (*config.GarageStorageConfig, error) {
	cfg := primary
	if cfg == nil {
		cfg = fallback
	}
	if cfg == nil {
		return nil, fmt.Errorf("%s config cannot be empty", name)
	}
	return cfg, nil
}

func redisURL(redisCfg *config.RedisConfig) string {
	if redisCfg.Password == "" {
		return fmt.Sprintf("redis://%s:%s/%d", redisCfg.Host, redisCfg.Port, redisCfg.DB)
	}
	return fmt.Sprintf("redis://:%s@%s:%s/%d", redisCfg.Password, redisCfg.Host, redisCfg.Port, redisCfg.DB)
}

func isStorageNotFoundError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "key not found") || strings.Contains(msg, "not found")
}
