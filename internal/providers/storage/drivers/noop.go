package drivers

import (
	"errors"
	"time"

	"lfiber/internal/providers/storage/contracts"
)

// NoopDisk 实现 contracts.Disk 接口。
// 用于在存储服务降级或禁用时作为安全的占位驱动，避免空指针崩溃。
type NoopDisk struct{}

var _ contracts.Disk = (*NoopDisk)(nil)

// NewNoopDisk 创建并返回一个 NoopDisk 实例。
func NewNoopDisk() *NoopDisk {
	return &NoopDisk{}
}

func (n *NoopDisk) Get(path string) ([]byte, error) {
	return nil, errors.New("file not found in fallback storage")
}

func (n *NoopDisk) Put(path string, contents []byte, options ...interface{}) error {
	return nil
}

func (n *NoopDisk) Exists(path string) (bool, error) {
	return false, nil
}

func (n *NoopDisk) Missing(path string) (bool, error) {
	return true, nil
}

func (n *NoopDisk) Url(path string) string {
	return ""
}

func (n *NoopDisk) TemporaryUrl(path string, expiration time.Duration) (string, error) {
	return "", nil
}

func (n *NoopDisk) Size(path string) (int64, error) {
	return 0, nil
}

func (n *NoopDisk) LastModified(path string) (time.Time, error) {
	return time.Time{}, nil
}

func (n *NoopDisk) Copy(from, to string) error {
	return nil
}

func (n *NoopDisk) Move(from, to string) error {
	return nil
}

func (n *NoopDisk) Prepend(path string, contents []byte) error {
	return nil
}

func (n *NoopDisk) Append(path string, contents []byte) error {
	return nil
}

func (n *NoopDisk) Delete(paths ...string) error {
	return nil
}

func (n *NoopDisk) Files(directory string, recursive ...bool) ([]string, error) {
	return []string{}, nil
}

func (n *NoopDisk) AllFiles(directory string) ([]string, error) {
	return []string{}, nil
}

func (n *NoopDisk) Directories(directory string, recursive ...bool) ([]string, error) {
	return []string{}, nil
}

func (n *NoopDisk) AllDirectories(directory string) ([]string, error) {
	return []string{}, nil
}

func (n *NoopDisk) MakeDirectory(path string) error {
	return nil
}

func (n *NoopDisk) DeleteDirectory(path string) error {
	return nil
}

func (n *NoopDisk) GetVisibility(path string) (string, error) {
	return "private", nil
}

func (n *NoopDisk) SetVisibility(path string, visibility string) error {
	return nil
}

func (n *NoopDisk) Reset() error {
	return nil
}

func (n *NoopDisk) HealthCheck() error {
	return nil
}

func (n *NoopDisk) Close() error {
	return nil
}
