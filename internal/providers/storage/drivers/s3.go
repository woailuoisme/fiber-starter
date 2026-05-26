package drivers

import (
	"fmt"
	"strings"
	"time"

	"lfiber/configs"
	helpers "lfiber/internal/support"

	"github.com/gofiber/storage/s3"
	"go.uber.org/zap"
)

// S3Driver implements the Disk interface for S3-compatible storage
type S3Driver struct {
	storage *s3.Storage
	config  *configs.StorageConfig
	bucket  string
}

func NewS3Driver(cfg *configs.Config, name string) *S3Driver {
	var s3Cfg s3.Config
	var err error

	switch name {
	case "garage":
		s3Cfg, err = buildS3ConfigFromGarage(cfg.Storage.Garage, cfg.Storage.Reset)
	case "s3":
		s3Cfg, err = buildS3Config(cfg.Storage.S3, cfg.Storage.Reset)
	case "r2":
		s3Cfg, err = buildS3Config(cfg.Storage.R2, cfg.Storage.Reset)
	case "oss":
		s3Cfg, err = buildS3Config(cfg.Storage.OSS, cfg.Storage.Reset)
	}

	if err != nil {
		helpers.LogError("Failed to build S3 config", zap.Error(err), zap.String("driver", name))
		return nil
	}

	store := s3.New(s3Cfg)
	helpers.Info("S3 storage driver initialized", zap.String("driver", name), zap.String("bucket", s3Cfg.Bucket))

	return &S3Driver{
		storage: store,
		config:  &cfg.Storage,
		bucket:  s3Cfg.Bucket,
	}
}

func (d *S3Driver) Get(path string) ([]byte, error) {
	return d.storage.Get(path)
}

func (d *S3Driver) Put(path string, contents []byte, options ...interface{}) error {
	var exp time.Duration
	// Handle potential TTL in options if needed, or other options
	return d.storage.Set(path, contents, exp)
}

func (d *S3Driver) Exists(path string) (bool, error) {
	val, err := d.Get(path)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return false, nil
		}
		return false, err
	}
	return val != nil, nil
}

func (d *S3Driver) Missing(path string) (bool, error) {
	exists, err := d.Exists(path)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (d *S3Driver) Url(path string) string {
	// In a real app, this would return the signed or public URL
	return fmt.Sprintf("/storage/%s", path)
}

func (d *S3Driver) TemporaryUrl(path string, expiration time.Duration) (string, error) {
	// Signed URL implementation would go here
	return d.Url(path), nil
}

func (d *S3Driver) Size(path string) (int64, error) {
	val, err := d.Get(path)
	if err != nil {
		return 0, err
	}
	return int64(len(val)), nil
}

func (d *S3Driver) LastModified(path string) (time.Time, error) {
	return time.Now(), nil
}

func (d *S3Driver) Copy(from, to string) error {
	content, err := d.Get(from)
	if err != nil {
		return err
	}
	return d.Put(to, content)
}

func (d *S3Driver) Move(from, to string) error {
	if err := d.Copy(from, to); err != nil {
		return err
	}
	return d.Delete(from)
}

func (d *S3Driver) Prepend(path string, contents []byte) error {
	existing, _ := d.Get(path)
	return d.Put(path, append(contents, existing...))
}

func (d *S3Driver) Append(path string, contents []byte) error {
	existing, _ := d.Get(path)
	return d.Put(path, append(existing, contents...))
}

func (d *S3Driver) Delete(paths ...string) error {
	for _, path := range paths {
		if err := d.storage.Delete(path); err != nil {
			return err
		}
	}
	return nil
}

func (d *S3Driver) Files(directory string, recursive ...bool) ([]string, error) {
	return nil, fmt.Errorf("listing files is not supported by current S3 driver")
}

func (d *S3Driver) AllFiles(directory string) ([]string, error) {
	return d.Files(directory, true)
}

func (d *S3Driver) Directories(directory string, recursive ...bool) ([]string, error) {
	return nil, fmt.Errorf("listing directories is not supported by current S3 driver")
}

func (d *S3Driver) AllDirectories(directory string) ([]string, error) {
	return d.Directories(directory, true)
}

func (d *S3Driver) MakeDirectory(path string) error {
	return nil // S3 is flat, no need to create directories
}

func (d *S3Driver) DeleteDirectory(path string) error {
	return fmt.Errorf("deleting directories is not supported by current S3 driver")
}

func (d *S3Driver) GetVisibility(path string) (string, error) {
	return "public", nil
}

func (d *S3Driver) SetVisibility(path string, visibility string) error {
	return nil
}

func (d *S3Driver) Reset() error {
	return d.storage.Reset()
}

func (d *S3Driver) HealthCheck() error {
	if d.storage == nil {
		return fmt.Errorf("s3 storage not initialized")
	}
	// Try a dummy Get to check connection
	_, err := d.storage.Get("__health_check__")
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}
	return nil
}

func (d *S3Driver) Close() error {
	return d.storage.Close()
}

// Helpers moved from storage_backend.go

func buildS3Config(cfg *configs.S3StorageConfig, reset bool) (s3.Config, error) {
	if cfg == nil {
		return s3.Config{}, fmt.Errorf("s3 config cannot be empty")
	}

	return s3.Config{
		Bucket:   cfg.Bucket,
		Endpoint: cfg.Endpoint,
		Region:   cfg.Region,
		Reset:    reset,
		Credentials: s3.Credentials{
			AccessKey:       cfg.AccessKeyID,
			SecretAccessKey: cfg.SecretAccessKey,
		},
	}, nil
}

func buildS3ConfigFromGarage(cfg *configs.GarageStorageConfig, reset bool) (s3.Config, error) {
	if cfg == nil {
		return s3.Config{}, fmt.Errorf("garage config cannot be empty")
	}

	endpoint := cfg.Endpoint
	if cfg.UseSSL {
		if !strings.HasPrefix(endpoint, "http") {
			endpoint = "https://" + endpoint
		}
	} else {
		if !strings.HasPrefix(endpoint, "http") {
			endpoint = "http://" + endpoint
		}
	}

	return s3.Config{
		Bucket:   cfg.Bucket,
		Endpoint: endpoint,
		Region:   cfg.Region,
		Reset:    reset,
		Credentials: s3.Credentials{
			AccessKey:       cfg.AccessKeyID,
			SecretAccessKey: cfg.SecretAccessKey,
		},
	}, nil
}
