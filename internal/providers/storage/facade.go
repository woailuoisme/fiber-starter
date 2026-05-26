package storage

import (
	"errors"
	"time"

	"lfiber/internal/providers/storage/contracts"
	"lfiber/internal/support/appctx"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// manager returns the storage manager instance from the container.
func manager() contracts.StorageManager {
	if app := appctx.App(); app != nil {
		return app.StorageValue()
	}
	return nil
}

// disk returns the default storage disk from the container.
func disk() contracts.Disk {
	if m := manager(); m != nil {
		return m.Disk()
	}
	return nil
}

// GetDisk returns a specific storage disk (Facade method)
func GetDisk(name ...string) (contracts.Disk, error) {
	if m := manager(); m != nil {
		return m.Disk(name...), nil
	}
	return nil, ErrContainerNotInitialized
}

// Drive returns a specific storage disk
func Drive(name ...string) contracts.Disk {
	if m := manager(); m != nil {
		return m.Disk(name...)
	}
	return nil
}

// Global Facade methods for the default disk

func Get(path string) ([]byte, error) {
	if d := disk(); d != nil {
		return d.Get(path)
	}
	return nil, ErrContainerNotInitialized
}

func Put(path string, contents []byte, options ...interface{}) error {
	if d := disk(); d != nil {
		return d.Put(path, contents, options...)
	}
	return ErrContainerNotInitialized
}

func Delete(paths ...string) error {
	if d := disk(); d != nil {
		return d.Delete(paths...)
	}
	return ErrContainerNotInitialized
}

func Exists(path string) (bool, error) {
	if d := disk(); d != nil {
		return d.Exists(path)
	}
	return false, ErrContainerNotInitialized
}

func Missing(path string) (bool, error) {
	if d := disk(); d != nil {
		return d.Missing(path)
	}
	return false, ErrContainerNotInitialized
}

func Url(path string) string {
	if d := disk(); d != nil {
		return d.Url(path)
	}
	return ""
}

func TemporaryUrl(path string, expiration time.Duration) (string, error) {
	if d := disk(); d != nil {
		return d.TemporaryUrl(path, expiration)
	}
	return "", ErrContainerNotInitialized
}

func Size(path string) (int64, error) {
	if d := disk(); d != nil {
		return d.Size(path)
	}
	return 0, ErrContainerNotInitialized
}

func LastModified(path string) (time.Time, error) {
	if d := disk(); d != nil {
		return d.LastModified(path)
	}
	return time.Time{}, ErrContainerNotInitialized
}

func Copy(from, to string) error {
	if d := disk(); d != nil {
		return d.Copy(from, to)
	}
	return ErrContainerNotInitialized
}

func Move(from, to string) error {
	if d := disk(); d != nil {
		return d.Move(from, to)
	}
	return ErrContainerNotInitialized
}

func Prepend(path string, contents []byte) error {
	if d := disk(); d != nil {
		return d.Prepend(path, contents)
	}
	return ErrContainerNotInitialized
}

func Append(path string, contents []byte) error {
	if d := disk(); d != nil {
		return d.Append(path, contents)
	}
	return ErrContainerNotInitialized
}

func Files(directory string, recursive ...bool) ([]string, error) {
	if d := disk(); d != nil {
		return d.Files(directory, recursive...)
	}
	return nil, ErrContainerNotInitialized
}

func AllFiles(directory string) ([]string, error) {
	if d := disk(); d != nil {
		return d.AllFiles(directory)
	}
	return nil, ErrContainerNotInitialized
}

func Directories(directory string, recursive ...bool) ([]string, error) {
	if d := disk(); d != nil {
		return d.Directories(directory, recursive...)
	}
	return nil, ErrContainerNotInitialized
}

func AllDirectories(directory string) ([]string, error) {
	if d := disk(); d != nil {
		return d.AllDirectories(directory)
	}
	return nil, ErrContainerNotInitialized
}

func MakeDirectory(path string) error {
	if d := disk(); d != nil {
		return d.MakeDirectory(path)
	}
	return ErrContainerNotInitialized
}

func DeleteDirectory(path string) error {
	if d := disk(); d != nil {
		return d.DeleteDirectory(path)
	}
	return ErrContainerNotInitialized
}

func GetVisibility(path string) (string, error) {
	if d := disk(); d != nil {
		return d.GetVisibility(path)
	}
	return "", ErrContainerNotInitialized
}

func SetVisibility(path string, visibility string) error {
	if d := disk(); d != nil {
		return d.SetVisibility(path, visibility)
	}
	return ErrContainerNotInitialized
}
