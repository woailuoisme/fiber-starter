package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
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

// MinIOStorage MinIO存储适配器 (已废弃，使用S3Storage)
// type MinIOStorage struct {
// 	client *minio.Client
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

	var store storage.Storage
	var err error

	switch s.config.Driver {
	case "redis":
		store = createRedisStorage(s.config, s.redisCfg)
	case "minio":
		store, err = createMinIOStorage(s.config)
	case "s3":
		store, err = createS3Storage(s.config)
	case "r2":
		store, err = createR2Storage(s.config)
	case "oss":
		store, err = createOSSStorage(s.config)
	default:
		store = createDefaultStorage(s.config, s.redisCfg)
	}

	if err != nil {
		return err
	}

	s.storage = store
	atomic.StoreUint32(&s.initialized, 1)
	helpers.Logger.Info("Storage service initialized", zap.String("driver", s.config.Driver))

	return nil
}

func createRedisStorage(cfg *config.StorageConfig, redisCfg *config.RedisConfig) storage.Storage {
	var url string
	if redisCfg.Password == "" {
		url = fmt.Sprintf("redis://%s:%s/%d",
			redisCfg.Host, redisCfg.Port, redisCfg.DB)
	} else {
		url = fmt.Sprintf("redis://:%s@%s:%s/%d",
			redisCfg.Password, redisCfg.Host, redisCfg.Port, redisCfg.DB)
	}
	return redisStorage.New(redisStorage.Config{
		URL:   url,
		Reset: cfg.Reset,
	})
}

func createMinIOStorage(cfg *config.StorageConfig) (storage.Storage, error) {
	if cfg.MinIO == nil {
		return nil, fmt.Errorf("minio config cannot be empty")
	}

	scheme := "http"
	if cfg.MinIO.UseSSL {
		scheme = "https"
	}
	endpoint := fmt.Sprintf("%s://%s", scheme, cfg.MinIO.Endpoint)

	s3Cfg := &config.S3StorageConfig{
		AccessKeyID:     cfg.MinIO.AccessKeyID,
		SecretAccessKey: cfg.MinIO.SecretAccessKey,
		Region:          cfg.MinIO.Region,
		Bucket:          cfg.MinIO.Bucket,
		Endpoint:        endpoint,
	}

	s3Storage, err := NewS3Storage(s3Cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio storage (via s3 sdk): %w", err)
	}

	// 确保存储桶存在
	if err := s3Storage.EnsureBucket(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket: %w", err)
	}

	return s3Storage, nil
}

func createS3Storage(cfg *config.StorageConfig) (storage.Storage, error) {
	s3Storage, err := NewS3Storage(cfg.S3)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3 storage: %w", err)
	}
	return s3Storage, nil
}

func createR2Storage(cfg *config.StorageConfig) (storage.Storage, error) {
	if cfg.R2 == nil {
		return nil, fmt.Errorf("r2 config cannot be empty")
	}
	s3Storage, err := NewS3Storage(cfg.R2)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize r2 storage: %w", err)
	}
	return s3Storage, nil
}

func createOSSStorage(cfg *config.StorageConfig) (storage.Storage, error) {
	if cfg.OSS == nil {
		return nil, fmt.Errorf("oss config cannot be empty")
	}
	s3Storage, err := NewS3Storage(cfg.OSS)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize oss storage: %w", err)
	}
	return s3Storage, nil
}

func createDefaultStorage(cfg *config.StorageConfig, redisCfg *config.RedisConfig) storage.Storage {
	helpers.Logger.Warn("Unknown storage driver, trying to use Redis storage as default", zap.String("driver", cfg.Driver))
	return createRedisStorage(cfg, redisCfg)
}

// Get 获取存储值
func (s *StorageService) Get(key string) ([]byte, error) {
	if err := s.ensureInitialized(); err != nil {
		return nil, err
	}

	if s.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	value, err := s.storage.Get(key)
	if err != nil {
		return nil, err
	}

	// 如果启用了压缩，需要解压数据
	if s.compression != CompressionNone {
		decompressed, err := s.decompressData(value)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress data: %w", err)
		}
		value = decompressed
	}

	return value, nil
}

// Set 设置存储值
func (s *StorageService) Set(key string, value []byte, ttl time.Duration) error {
	if err := s.ensureInitialized(); err != nil {
		return err
	}

	// 如果启用了压缩，先压缩数据
	if s.compression != CompressionNone {
		compressed, err := s.compressData(value)
		if err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
		value = compressed
	}

	return s.storage.Set(key, value, ttl)
}

// Delete 删除存储值
func (s *StorageService) Delete(key string) error {
	if err := s.ensureInitialized(); err != nil {
		return err
	}
	return s.storage.Delete(key)
}

// Reset 重置存储
func (s *StorageService) Reset() error {
	if err := s.ensureInitialized(); err != nil {
		return err
	}
	return s.storage.Reset()
}

// Close 关闭存储服务
func (s *StorageService) Close() error {
	if atomic.LoadUint32(&s.initialized) == 0 {
		return nil
	}
	if s.zstdEncoder != nil {
		_ = s.zstdEncoder.Close()
	}
	if s.zstdDecoder != nil {
		s.zstdDecoder.Close()
	}
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
		// 初始化Zstd编码器
		encoder, err := zstd.NewWriter(nil)
		if err != nil {
			return fmt.Errorf("failed to create zstd encoder: %w", err)
		}
		s.zstdEncoder = encoder

		// 初始化Zstd解码器
		decoder, err := zstd.NewReader(nil)
		if err != nil {
			return fmt.Errorf("failed to create zstd decoder: %w", err)
		}
		s.zstdDecoder = decoder

	case CompressionGzip:
		// Gzip不需要预初始化

	case CompressionNone:
		// 无压缩

	default:
		return fmt.Errorf("unsupported compression type: %d", compressionType)
	}

	return nil
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
		// 检查是否是"未找到"错误
		if err.Error() == "key not found" || err.Error() == "not found" {
			return false, nil
		}
		return false, err
	}
	return val != nil, nil
}

// SetExpire 设置键的过期时间
func (s *StorageService) SetExpire(key string, ttl time.Duration) error {
	// 获取当前值
	val, err := s.Get(key)
	if err != nil {
		return err
	}
	// 重新设置带过期时间的值
	return s.Set(key, val, ttl)
}

// NewMinIOStorage 创建MinIO存储实例 (已废弃，直接使用S3兼容模式)
// func NewMinIOStorage(cfg *config.MinIOStorageConfig) (*MinIOStorage, error) {
// 	return nil, fmt.Errorf("NewMinIOStorage is deprecated")
// }

// NewS3Storage 创建S3存储实例
func NewS3Storage(cfg *config.S3StorageConfig) (*S3Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("s3 config cannot be empty")
	}

	// 创建AWS配置选项
	configOptions := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithCredentialsProvider(aws.NewCredentialsCache(
			aws.CredentialsProviderFunc(func(_ context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     cfg.AccessKeyID,
					SecretAccessKey: cfg.SecretAccessKey,
				}, nil
			}),
		)),
		awsConfig.WithRegion(cfg.Region),
	}

	// 创建AWS配置
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(), configOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws config: %w", err)
	}

	// 创建S3客户端
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

// Get 获取存储值 (MinIO实现 - 已废弃)
// func (m *MinIOStorage) Get(key string) ([]byte, error) {
// 	return nil, fmt.Errorf("deprecated")
// }

// Set 设置存储值 (MinIO实现 - 已废弃)
// func (m *MinIOStorage) Set(key string, value []byte, ttl time.Duration) error {
// 	return fmt.Errorf("deprecated")
// }

// Delete 删除存储值 (MinIO实现 - 已废弃)
// func (m *MinIOStorage) Delete(key string) error {
// 	return fmt.Errorf("deprecated")
// }

// Reset 重置存储 (MinIO实现 - 已废弃)
// func (m *MinIOStorage) Reset() error {
// 	return fmt.Errorf("deprecated")
// }

// Close 关闭存储连接 (MinIO实现 - 已废弃)
// func (m *MinIOStorage) Close() error {
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
