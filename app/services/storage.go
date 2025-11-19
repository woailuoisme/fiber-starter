package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"fiber-starter/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/storage"
	"github.com/gofiber/storage/bbolt"
	"github.com/gofiber/storage/memory"
	redisStorage "github.com/gofiber/storage/redis/v3"
	"github.com/klauspost/compress/zstd"
	minio "github.com/minio/minio-go/v7"
	minioCredentials "github.com/minio/minio-go/v7/pkg/credentials"
)

// CompressionType 压缩类型
type CompressionType int

const (
	CompressionNone CompressionType = iota
	CompressionGzip
	CompressionZstd
)

// StorageService 存储服务
type StorageService struct {
	storage     storage.Storage
	config      *config.StorageConfig
	compression CompressionType
	zstdEncoder *zstd.Encoder
	zstdDecoder *zstd.Decoder
}

// MinIOStorage MinIO存储适配器
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// S3Storage S3存储适配器
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewStorageService 创建新的存储服务实例
func NewStorageService(cfg *config.StorageConfig, redisCfg *config.RedisConfig) (*StorageService, error) {
	var store storage.Storage

	switch cfg.Driver {
	case "redis":
		// 使用Redis存储
		// 如果密码为空，则不需要密码部分
		var url string
		if redisCfg.Password == "" {
			url = fmt.Sprintf("redis://%s:%s/%d",
				redisCfg.Host, redisCfg.Port, redisCfg.DB)
		} else {
			url = fmt.Sprintf("redis://:%s@%s:%s/%d",
				redisCfg.Password, redisCfg.Host, redisCfg.Port, redisCfg.DB)
		}
		store = redisStorage.New(redisStorage.Config{
			URL:   url,
			Reset: cfg.Reset,
		})

	case "bbolt":
		// 使用BBolt存储
		store = bbolt.New(bbolt.Config{
			Database: cfg.Database,
			Bucket:   "fiber",
			Reset:    false, // 不自动重置
		})

	case "minio":
		// 使用MinIO存储
		minioStorage, err := NewMinIOStorage(cfg.MinIO)
		if err != nil {
			return nil, fmt.Errorf("初始化MinIO存储失败: %w", err)
		}
		store = minioStorage

	case "s3":
		// 使用S3存储
		s3Storage, err := NewS3Storage(cfg.S3)
		if err != nil {
			return nil, fmt.Errorf("初始化S3存储失败: %w", err)
		}
		store = s3Storage

	case "memory":
		// 使用内存存储（默认）
		store = memory.New(memory.Config{})

	default:
		// 默认使用内存存储
		log.Printf("未知的存储驱动 '%s'，使用内存存储作为默认", cfg.Driver)
		store = memory.New(memory.Config{})
	}

	log.Printf("存储服务已初始化，驱动类型: %s", cfg.Driver)

	return &StorageService{
		storage: store,
		config:  cfg,
	}, nil
}

// Get 获取存储值
func (s *StorageService) Get(key string) ([]byte, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("存储未初始化")
	}

	value, err := s.storage.Get(key)
	if err != nil {
		return nil, err
	}

	// 如果启用了压缩，需要解压数据
	if s.compression != CompressionNone {
		decompressed, err := s.decompressData(value)
		if err != nil {
			return nil, fmt.Errorf("解压数据失败: %w", err)
		}
		value = decompressed
	}

	return value, nil
}

// Set 设置存储值
func (s *StorageService) Set(key string, value []byte, ttl time.Duration) error {
	// 如果启用了压缩，先压缩数据
	if s.compression != CompressionNone {
		compressed, err := s.compressData(value)
		if err != nil {
			return fmt.Errorf("压缩数据失败: %w", err)
		}
		value = compressed
	}

	return s.storage.Set(key, value, ttl)
}

// Delete 删除存储值
func (s *StorageService) Delete(key string) error {
	return s.storage.Delete(key)
}

// Reset 重置存储
func (s *StorageService) Reset() error {
	return s.storage.Reset()
}

// Close 关闭存储服务
func (s *StorageService) Close() error {
	if s.zstdEncoder != nil {
		s.zstdEncoder.Close()
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
			return fmt.Errorf("创建Zstd编码器失败: %w", err)
		}
		s.zstdEncoder = encoder

		// 初始化Zstd解码器
		decoder, err := zstd.NewReader(nil)
		if err != nil {
			return fmt.Errorf("创建Zstd解码器失败: %w", err)
		}
		s.zstdDecoder = decoder

	case CompressionGzip:
		// Gzip不需要预初始化

	case CompressionNone:
		// 无压缩

	default:
		return fmt.Errorf("不支持的压缩类型: %d", compressionType)
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
			return nil, fmt.Errorf("Gzip压缩失败: %w", err)
		}
		err = writer.Close()
		if err != nil {
			return nil, fmt.Errorf("关闭Gzip写入器失败: %w", err)
		}
		return buf.Bytes(), nil

	case CompressionZstd:
		if s.zstdEncoder == nil {
			return nil, fmt.Errorf("Zstd编码器未初始化")
		}
		return s.zstdEncoder.EncodeAll(data, nil), nil

	default:
		return nil, fmt.Errorf("不支持的压缩类型: %d", s.compression)
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
			return nil, fmt.Errorf("创建Gzip读取器失败: %w", err)
		}
		defer reader.Close()

		result, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("Gzip解压失败: %w", err)
		}
		return result, nil

	case CompressionZstd:
		if s.zstdDecoder == nil {
			return nil, fmt.Errorf("Zstd解码器未初始化")
		}
		return s.zstdDecoder.DecodeAll(data, nil)

	default:
		return nil, fmt.Errorf("不支持的压缩类型: %d", s.compression)
	}
}

// GetCompressionType 获取当前压缩类型
func (s *StorageService) GetCompressionType() CompressionType {
	return s.compression
}

// GetCompressionTypeName 获取压缩类型名称
func (s *StorageService) GetCompressionTypeName() string {
	switch s.compression {
	case CompressionNone:
		return "无压缩"
	case CompressionGzip:
		return "Gzip"
	case CompressionZstd:
		return "Zstd"
	default:
		return "未知"
	}
}

// GetStorage 获取底层存储实例
func (s *StorageService) GetStorage() storage.Storage {
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

// NewMinIOStorage 创建MinIO存储实例
func NewMinIOStorage(cfg *config.MinIOStorageConfig) (*MinIOStorage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("MinIO配置不能为空")
	}

	// 创建MinIO客户端
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  minioCredentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建MinIO客户端失败: %w", err)
	}

	// 检查存储桶是否存在，如果不存在则创建
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("检查存储桶失败: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region})
		if err != nil {
			return nil, fmt.Errorf("创建存储桶失败: %w", err)
		}
		log.Printf("MinIO存储桶 '%s' 已创建", cfg.Bucket)
	}

	return &MinIOStorage{
		client: minioClient,
		bucket: cfg.Bucket,
	}, nil
}

// NewS3Storage 创建S3存储实例
func NewS3Storage(cfg *config.S3StorageConfig) (*S3Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("S3配置不能为空")
	}

	// 创建AWS配置
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithCredentialsProvider(aws.NewCredentialsCache(
			aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     cfg.AccessKeyID,
					SecretAccessKey: cfg.SecretAccessKey,
				}, nil
			}),
		)),
		awsConfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("创建AWS配置失败: %w", err)
	}

	// 创建S3客户端
	var s3Client *s3.Client
	if cfg.Endpoint != "" {
		// 使用自定义端点（兼容其他S3服务）
		s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	} else {
		s3Client = s3.NewFromConfig(awsCfg)
	}

	return &S3Storage{
		client: s3Client,
		bucket: cfg.Bucket,
	}, nil
}

// Get 获取存储值 (MinIO实现)
func (m *MinIOStorage) Get(key string) ([]byte, error) {
	ctx := context.Background()
	obj, err := m.client.GetObject(ctx, m.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, stat.Size)
	_, err = obj.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Set 设置存储值 (MinIO实现)
func (m *MinIOStorage) Set(key string, value []byte, ttl time.Duration) error {
	ctx := context.Background()
	_, err := m.client.PutObject(ctx, m.bucket, key, bytes.NewReader(value), int64(len(value)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	return err
}

// Delete 删除存储值 (MinIO实现)
func (m *MinIOStorage) Delete(key string) error {
	ctx := context.Background()
	return m.client.RemoveObject(ctx, m.bucket, key, minio.RemoveObjectOptions{})
}

// Reset 重置存储 (MinIO实现)
func (m *MinIOStorage) Reset() error {
	ctx := context.Background()
	objectsCh := make(chan minio.ObjectInfo)

	// 列出所有对象
	go func() {
		defer close(objectsCh)
		for object := range m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{}) {
			if object.Err != nil {
				return
			}
			objectsCh <- object
		}
	}()

	// 删除所有对象
	errorCh := m.client.RemoveObjects(ctx, m.bucket, objectsCh, minio.RemoveObjectsOptions{})
	for err := range errorCh {
		if err.Err != nil {
			return err.Err
		}
	}

	return nil
}

// Close 关闭存储连接 (MinIO实现)
func (m *MinIOStorage) Close() error {
	// MinIO客户端不需要显式关闭
	return nil
}

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
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// Set 设置存储值 (S3实现)
func (s *S3Storage) Set(key string, value []byte, ttl time.Duration) error {
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

// Close 关闭存储连接 (S3实现)
func (s *S3Storage) Close() error {
	// S3客户端不需要显式关闭
	return nil
}
