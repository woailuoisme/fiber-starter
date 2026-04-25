package services

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/storage"
	"github.com/klauspost/compress/zstd"
	"go.uber.org/zap"
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
	redisCfg    *config.RedisConfig
	compression CompressionType
	zstdEncoder *zstd.Encoder
	zstdDecoder *zstd.Decoder
	initMu      sync.Mutex
	initialized uint32
}

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewStorageService(cfg *config.StorageConfig, redisCfg *config.RedisConfig) (*StorageService, error) {
	return &StorageService{config: cfg, redisCfg: redisCfg}, nil
}

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
	helpers.Info("Storage service initialized", zap.String("driver", driver))

	return nil
}

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

func (s *StorageService) Delete(key string) error {
	return s.withStorage(func(store storage.Storage) error { return store.Delete(key) })
}

func (s *StorageService) Reset() error {
	return s.withStorage(func(store storage.Storage) error { return store.Reset() })
}

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

func (s *StorageService) CompressData(data []byte) ([]byte, error)   { return s.compressData(data) }
func (s *StorageService) DecompressData(data []byte) ([]byte, error) { return s.decompressData(data) }

func (s *StorageService) GetCompressionType() CompressionType { return s.compression }

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

func (s *StorageService) GetStorage() storage.Storage {
	_ = s.ensureInitialized()
	return s.storage
}

func (s *StorageService) SetWithDefaultTTL(key string, value []byte) error {
	return s.Set(key, value, time.Hour)
}

func (s *StorageService) GetString(key string) (string, error) {
	val, err := s.Get(key)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func (s *StorageService) SetString(key, value string, ttl time.Duration) error {
	return s.Set(key, []byte(value), ttl)
}

func (s *StorageService) SetStringWithDefaultTTL(key, value string) error {
	return s.SetWithDefaultTTL(key, []byte(value))
}

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

func (s *StorageService) SetExpire(key string, ttl time.Duration) error {
	val, err := s.Get(key)
	if err != nil {
		return err
	}
	return s.Set(key, val, ttl)
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
