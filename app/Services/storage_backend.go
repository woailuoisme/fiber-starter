package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/storage"
	redisStorage "github.com/gofiber/storage/redis/v3"
	"go.uber.org/zap"
)

func buildStorage(cfg *config.StorageConfig, redisCfg *config.RedisConfig) (storage.Storage, string, error) {
	driver := normalizeStorageDriver(cfg.Driver)

	switch driver {
	case "redis":
		return createRedisStorage(cfg, redisCfg), driver, nil
	case "garage":
		return createGarageStorage(cfg)
	case "s3":
		return createS3Storage(cfg)
	case "r2":
		return createR2Storage(cfg)
	case "oss":
		return createOSSStorage(cfg)
	default:
		return createDefaultStorage(cfg, redisCfg), driver, nil
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

	scheme := "http"
	if garageCfg.UseSSL {
		scheme = "https"
	}
	endpoint := fmt.Sprintf("%s://%s", scheme, garageCfg.Endpoint)

	s3Cfg := &config.S3StorageConfig{
		AccessKeyID:     garageCfg.AccessKeyID,
		SecretAccessKey: garageCfg.SecretAccessKey,
		Region:          garageCfg.Region,
		Bucket:          garageCfg.Bucket,
		Endpoint:        endpoint,
	}

	s3Storage, err := NewS3Storage(s3Cfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize garage storage (via s3 sdk): %w", err)
	}

	if err := s3Storage.EnsureBucket(); err != nil {
		return nil, "", fmt.Errorf("failed to ensure bucket: %w", err)
	}

	return s3Storage, "garage", nil
}

func createS3Storage(cfg *config.StorageConfig) (storage.Storage, string, error) {
	s3Storage, err := NewS3Storage(cfg.S3)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize s3 storage: %w", err)
	}
	return s3Storage, "s3", nil
}

func createR2Storage(cfg *config.StorageConfig) (storage.Storage, string, error) {
	s3Cfg, err := storageS3Config("r2", cfg.R2)
	if err != nil {
		return nil, "", err
	}
	s3Storage, err := NewS3Storage(s3Cfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize r2 storage: %w", err)
	}
	return s3Storage, "r2", nil
}

func createOSSStorage(cfg *config.StorageConfig) (storage.Storage, string, error) {
	s3Cfg, err := storageS3Config("oss", cfg.OSS)
	if err != nil {
		return nil, "", err
	}
	s3Storage, err := NewS3Storage(s3Cfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize oss storage: %w", err)
	}
	return s3Storage, "oss", nil
}

func createDefaultStorage(cfg *config.StorageConfig, redisCfg *config.RedisConfig) storage.Storage {
	helpers.Warn("Unknown storage driver, trying to use Redis storage as default", zap.String("driver", cfg.Driver))
	return createRedisStorage(cfg, redisCfg)
}

func NewS3Storage(cfg *config.S3StorageConfig) (*S3Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("s3 config cannot be empty")
	}

	awsCfg, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Storage{client: s3Client, bucket: cfg.Bucket}, nil
}

func loadAWSConfig(cfg *config.S3StorageConfig) (aws.Config, error) {
	return awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithCredentialsProvider(aws.NewCredentialsCache(
			aws.CredentialsProviderFunc(func(_ context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     cfg.AccessKeyID,
					SecretAccessKey: cfg.SecretAccessKey,
				}, nil
			}),
		)),
		awsConfig.WithRegion(cfg.Region),
	)
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

func storageS3Config(name string, cfg *config.S3StorageConfig) (*config.S3StorageConfig, error) {
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

func (s *S3Storage) Get(key string) ([]byte, error) {
	ctx := context.Background()
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	defer func() { _ = result.Body.Close() }()
	return io.ReadAll(result.Body)
}

func (s *S3Storage) Set(key string, value []byte, _ time.Duration) error {
	ctx := context.Background()
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(value),
		ContentType: aws.String("application/octet-stream"),
	})
	return err
}

func (s *S3Storage) Delete(key string) error {
	ctx := context.Background()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return err
}

func (s *S3Storage) Reset() error {
	ctx := context.Background()
	listResult, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket)})
	if err != nil {
		return err
	}
	for _, object := range listResult.Contents {
		if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: object.Key}); err != nil {
			return err
		}
	}
	return nil
}

func (s *S3Storage) EnsureBucket() error {
	ctx := context.Background()
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	if err == nil {
		return nil
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)})
	if err != nil {
		return fmt.Errorf("create bucket failed: %w", err)
	}
	return nil
}

func (s *S3Storage) Close() error { return nil }
