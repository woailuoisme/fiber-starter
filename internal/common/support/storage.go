package support

import (
	"fiber-starter/internal/providers/storage/Contracts"
)

// storage is the default storage disk instance
var storage Contracts.Disk

// manager is the full manager for multi-disk support
var manager Contracts.StorageManager

// InitStorage initializes the storage support with the manager
func InitStorage(m Contracts.StorageManager) {
	manager = m
	storage = m.Disk()
}

// Disk returns a specific storage disk
func Disk(name ...string) Contracts.Disk {
	if manager == nil {
		panic("storage manager not initialized")
	}
	return manager.Disk(name...)
}

// Get standard methods on the default disk
func Get(path string) ([]byte, error) {
	return storage.Get(path)
}

func Put(path string, contents []byte, options ...interface{}) error {
	return storage.Put(path, contents, options...)
}

func Exists(path string) (bool, error) {
	return storage.Exists(path)
}

func Delete(paths ...string) error {
	return storage.Delete(paths...)
}

func Url(path string) string {
	return storage.Url(path)
}
