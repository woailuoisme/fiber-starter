package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/storage"
	redisStorage "github.com/gofiber/storage/redis/v3"
	"github.com/klauspost/compress/zstd"
	"go.uber.org/zap"
)

// CompressionType 压缩类型
type CompressionType int

const (
	// CompressionNone no compression
	CompressionNone CompressionType = iota
	// CompressionGzip gzip compression
	CompressionGzip
	// CompressionZstd zstd compression
	CompressionZstd
)

// StorageService 存储服务
type StorageService struct {
	storage     storage.Storage
	config      *config.StorageConfig
	redisCfg    *config.RedisConfig
	compression CompressionType
	zstdEncoder *zstd.Encoder
	zstdDecoder *zstd.Decoder
	initMu      sync.Mutex
	initialized uint32 // 0: false, 1: true
}

// GarageStorage Garage存储适配器 (已废弃，使用S3Storage)
// type GarageStorage struct {
// 	client *s3.Client
// 	bucket string
// }

// S3Storage S3存储适配器
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewStorageService 创建新的存储服务实例
func NewStorageService(cfg *config.StorageConfig, redisCfg *config.RedisConfig) (*StorageService, error) {
	// 延迟初始化，仅保存配置
	return &StorageService{
		config:   cfg,
		redisCfg: redisCfg,
	}, nil
}

// ensureInitialized 确保存储服务已初始化
func (s *StorageService) ensureInitialized() error {
	if atomic.LoadUint32(&s.initialized) == 1 {
		return nil
	}

	s.initMu.Lock()
	defer s.initMu.Unlock()

	if s.initialized == 1 {
		return nil
	}

	store, driver, err := buildStorage(s.config, s.redisCfg)

	if err != nil {
		return err
	}

	s.storage = store
	atomic.StoreUint32(&s.initialized, 1)
	helpers.Logger.Info("Storage service initialized", zap.String("driver", driver))

	return nil
}

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
	helpers.Logger.Warn("Unknown storage driver, trying to use Redis storage as default", zap.String("driver", cfg.Driver))
	return createRedisStorage(cfg, redisCfg)
}

// Get 获取存储值
func (s *StorageService) Get(key string) ([]byte, error) {
	var value []byte
	err := s.withStorage(func(store storage.Storage) error {
		rawValue, err := store.Get(key)
		if err != nil {
			return err
		}

		if s.compression == CompressionNone {
			value = rawValue
			return nil
		}

		decompressed, err := s.decompressData(rawValue)
		if err != nil {
			return fmt.Errorf("failed to decompress data: %w", err)
		}
		value = decompressed
		return nil
	})
	return value, err
}

// Set 设置存储值
func (s *StorageService) Set(key string, value []byte, ttl time.Duration) error {
	return s.withStorage(func(store storage.Storage) error {
		if s.compression != CompressionNone {
			compressed, err := s.compressData(value)
			if err != nil {
				return fmt.Errorf("failed to compress data: %w", err)
			}
			value = compressed
		}
		return store.Set(key, value, ttl)
	})
}

// Delete 删除存储值
func (s *StorageService) Delete(key string) error {
	return s.withStorage(func(store storage.Storage) error {
		return store.Delete(key)
	})
}

// Reset 重置存储
func (s *StorageService) Reset() error {
	return s.withStorage(func(store storage.Storage) error {
		return store.Reset()
	})
}

// Close 关闭存储服务
func (s *StorageService) Close() error {
	if atomic.LoadUint32(&s.initialized) == 0 {
		return nil
	}
	s.closeCompression()
	if closer, ok := s.storage.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// SetCompression 设置压缩类型
func (s *StorageService) SetCompression(compressionType CompressionType) error {
	s.compression = compressionType

	switch compressionType {
	case CompressionZstd:
		return s.initZstd()
	case CompressionGzip:
		return nil
	case CompressionNone:
		s.closeCompression()
		return nil
	default:
		return fmt.Errorf("unsupported compression type: %d", compressionType)
	}
}

// CompressData 压缩数据（公开方法用于测试）
func (s *StorageService) CompressData(data []byte) ([]byte, error) {
	return s.compressData(data)
}

// DecompressData 解压数据（公开方法用于测试）
func (s *StorageService) DecompressData(data []byte) ([]byte, error) {
	return s.decompressData(data)
}

// compressData 压缩数据
func (s *StorageService) compressData(data []byte) ([]byte, error) {
	switch s.compression {
	case CompressionNone:
		return data, nil

	case CompressionGzip:
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		_, err := writer.Write(data)
		if err != nil {
			return nil, fmt.Errorf("gzip compression failed: %w", err)
		}
		err = writer.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}
		return buf.Bytes(), nil

	case CompressionZstd:
		if s.zstdEncoder == nil {
			return nil, fmt.Errorf("zstd encoder not initialized")
		}
		return s.zstdEncoder.EncodeAll(data, nil), nil

	default:
		return nil, fmt.Errorf("unsupported compression type: %d", s.compression)
	}
}

// decompressData 解压数据
func (s *StorageService) decompressData(data []byte) ([]byte, error) {
	switch s.compression {
	case CompressionNone:
		return data, nil

	case CompressionGzip:
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		// 使用匿名函数确保正确关闭并忽略错误
		defer func() {
			_ = reader.Close()
		}()

		result, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("gzip decompression failed: %w", err)
		}
		return result, nil

	case CompressionZstd:
		if s.zstdDecoder == nil {
			return nil, fmt.Errorf("zstd decoder not initialized")
		}
		return s.zstdDecoder.DecodeAll(data, nil)

	default:
		return nil, fmt.Errorf("unsupported compression type: %d", s.compression)
	}
}

// GetCompressionType 获取当前压缩类型
func (s *StorageService) GetCompressionType() CompressionType {
	return s.compression
}

// GetCompressionTypeName Get compression type name
func (s *StorageService) GetCompressionTypeName() string {
	switch s.compression {
	case CompressionNone:
		return "No compression"
	case CompressionGzip:
		return "Gzip"
	case CompressionZstd:
		return "Zstd"
	default:
		return "Unknown"
	}
}

// GetStorage 获取底层存储实例
func (s *StorageService) GetStorage() storage.Storage {
	_ = s.ensureInitialized()
	return s.storage
}

// SetWithDefaultTTL 使用默认TTL设置存储值
func (s *StorageService) SetWithDefaultTTL(key string, value []byte) error {
	// 默认TTL为1小时
	return s.Set(key, value, time.Hour)
}

// GetString 获取字符串值
func (s *StorageService) GetString(key string) (string, error) {
	val, err := s.Get(key)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// SetString 设置字符串值
func (s *StorageService) SetString(key, value string, ttl time.Duration) error {
	return s.Set(key, []byte(value), ttl)
}

// SetStringWithDefaultTTL 使用默认TTL设置字符串值
func (s *StorageService) SetStringWithDefaultTTL(key, value string) error {
	return s.SetWithDefaultTTL(key, []byte(value))
}

// Exists 检查键是否存在
func (s *StorageService) Exists(key string) (bool, error) {
	val, err := s.Get(key)
	if err != nil {
		if isStorageNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return val != nil, nil
}

// SetExpire 设置键的过期时间
func (s *StorageService) SetExpire(key string, ttl time.Duration) error {
	val, err := s.Get(key)
	if err != nil {
		return err
	}
	return s.Set(key, val, ttl)
}

// NewGarageStorage 创建Garage存储实例 (已废弃，直接使用S3兼容模式)
// func NewGarageStorage(cfg *config.GarageStorageConfig) (*GarageStorage, error) {
// 	return nil, fmt.Errorf("NewGarageStorage is deprecated")
// }

// NewS3Storage 创建S3存储实例
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

	return &S3Storage{
		client: s3Client,
		bucket: cfg.Bucket,
	}, nil
}

func (s *StorageService) withStorage(fn func(storage.Storage) error) error {
	if err := s.ensureInitialized(); err != nil {
		return err
	}
	if s.storage == nil {
		return fmt.Errorf("storage not initialized")
	}
	return fn(s.storage)
}

func (s *StorageService) initZstd() error {
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return fmt.Errorf("failed to create zstd encoder: %w", err)
	}

	decoder, err := zstd.NewReader(nil)
	if err != nil {
		_ = encoder.Close()
		return fmt.Errorf("failed to create zstd decoder: %w", err)
	}

	s.closeCompression()
	s.zstdEncoder = encoder
	s.zstdDecoder = decoder
	return nil
}

func (s *StorageService) closeCompression() {
	if s.zstdEncoder != nil {
		_ = s.zstdEncoder.Close()
		s.zstdEncoder = nil
	}
	if s.zstdDecoder != nil {
		s.zstdDecoder.Close()
		s.zstdDecoder = nil
	}
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

// Get 获取存储值 (Garage实现 - 已废弃)
// func (m *GarageStorage) Get(key string) ([]byte, error) {
// 	return nil, fmt.Errorf("deprecated")
// }

// Set 设置存储值 (Garage实现 - 已废弃)
// func (m *GarageStorage) Set(key string, value []byte, ttl time.Duration) error {
// 	return fmt.Errorf("deprecated")
// }

// Delete 删除存储值 (Garage实现 - 已废弃)
// func (m *GarageStorage) Delete(key string) error {
// 	return fmt.Errorf("deprecated")
// }

// Reset 重置存储 (Garage实现 - 已废弃)
// func (m *GarageStorage) Reset() error {
// 	return fmt.Errorf("deprecated")
// }

// Close 关闭存储连接 (Garage实现 - 已废弃)
// func (m *GarageStorage) Close() error {
// 	return nil
// }

// Get 获取存储值 (S3实现)
func (s *S3Storage) Get(key string) ([]byte, error) {
	ctx := context.Background()
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = result.Body.Close()
	}()

	return io.ReadAll(result.Body)
}

// Set 设置存储值 (S3实现)
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

// Delete 删除存储值 (S3实现)
func (s *S3Storage) Delete(key string) error {
	ctx := context.Background()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// Reset 重置存储 (S3实现)
func (s *S3Storage) Reset() error {
	ctx := context.Background()

	// 列出所有对象
	listResult, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return err
	}

	// 删除所有对象
	for _, object := range listResult.Contents {
		_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    object.Key,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// EnsureBucket 确保存储桶存在 (S3实现)
func (s *S3Storage) EnsureBucket() error {
	ctx := context.Background()
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}

	// 如果存储桶不存在，则创建
	// 注意：这里我们简单地假设所有错误都意味着存储桶不存在或无法访问
	// 在生产环境中，应该更仔细地检查错误类型
	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("create bucket failed: %w", err)
	}
	return nil
}

// Close 关闭存储连接 (S3实现)
func (s *S3Storage) Close() error {
	// S3客户端不需要显式关闭
	return nil
}
