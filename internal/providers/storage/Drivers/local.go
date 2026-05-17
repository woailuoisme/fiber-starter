package drivers

import (
	"os"
	"path/filepath"
	"time"

	helpers "fiber-starter/internal/support"

	"go.uber.org/zap"
)

type LocalDriver struct {
	root string
	url  string
}

func NewLocalDriver(root string, url string) *LocalDriver {
	// Ensure root directory exists
	// Using 0750 for security compliance as suggested by gosec
	if err := os.MkdirAll(root, 0o750); err != nil {
		helpers.Error("Failed to create local storage directory",
			zap.String("path", root),
			zap.Error(err))
	} else {
		helpers.Info("Local storage driver initialized",
			zap.String("root", root))
	}

	return &LocalDriver{
		root: root,
		url:  url,
	}
}

func (d *LocalDriver) Get(path string) ([]byte, error) {
	fullPath := filepath.Join(d.root, path)
	// #nosec G304
	return os.ReadFile(fullPath)
}

func (d *LocalDriver) Put(path string, contents []byte, options ...interface{}) error {
	fullPath := filepath.Join(d.root, path)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	// Using 0600 for security compliance
	return os.WriteFile(fullPath, contents, 0o600)
}

func (d *LocalDriver) Exists(path string) (bool, error) {
	fullPath := filepath.Join(d.root, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (d *LocalDriver) Missing(path string) (bool, error) {
	exists, err := d.Exists(path)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (d *LocalDriver) Url(path string) string {
	return filepath.Join(d.url, path)
}

func (d *LocalDriver) TemporaryUrl(path string, expiration time.Duration) (string, error) {
	// For local driver, we just return the standard URL as there's no native signed URL support
	return d.Url(path), nil
}

func (d *LocalDriver) Size(path string) (int64, error) {
	fullPath := filepath.Join(d.root, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (d *LocalDriver) LastModified(path string) (time.Time, error) {
	fullPath := filepath.Join(d.root, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func (d *LocalDriver) Copy(from, to string) error {
	content, err := d.Get(from)
	if err != nil {
		return err
	}
	return d.Put(to, content)
}

func (d *LocalDriver) Move(from, to string) error {
	fullFrom := filepath.Join(d.root, from)
	fullTo := filepath.Join(d.root, to)

	dir := filepath.Dir(fullTo)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	return os.Rename(fullFrom, fullTo)
}

func (d *LocalDriver) Prepend(path string, contents []byte) error {
	existing, _ := d.Get(path)
	return d.Put(path, append(contents, existing...))
}

func (d *LocalDriver) Append(path string, contents []byte) error {
	existing, _ := d.Get(path)
	return d.Put(path, append(existing, contents...))
}

func (d *LocalDriver) Delete(paths ...string) error {
	for _, path := range paths {
		fullPath := filepath.Join(d.root, path)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (d *LocalDriver) Files(directory string, recursive ...bool) ([]string, error) {
	isRecursive := false
	if len(recursive) > 0 {
		isRecursive = recursive[0]
	}

	var files []string
	fullDir := filepath.Join(d.root, directory)

	err := filepath.Walk(fullDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == fullDir {
			return nil
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(d.root, path)
			files = append(files, rel)
		}
		if !isRecursive && info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})

	return files, err
}

func (d *LocalDriver) AllFiles(directory string) ([]string, error) {
	return d.Files(directory, true)
}

func (d *LocalDriver) Directories(directory string, recursive ...bool) ([]string, error) {
	isRecursive := false
	if len(recursive) > 0 {
		isRecursive = recursive[0]
	}

	var dirs []string
	fullDir := filepath.Join(d.root, directory)

	err := filepath.Walk(fullDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == fullDir {
			return nil
		}
		if info.IsDir() {
			rel, _ := filepath.Rel(d.root, path)
			dirs = append(dirs, rel)
			if !isRecursive {
				return filepath.SkipDir
			}
		}
		return nil
	})

	return dirs, err
}

func (d *LocalDriver) AllDirectories(directory string) ([]string, error) {
	return d.Directories(directory, true)
}

func (d *LocalDriver) MakeDirectory(path string) error {
	fullPath := filepath.Join(d.root, path)
	return os.MkdirAll(fullPath, 0o750)
}

func (d *LocalDriver) DeleteDirectory(path string) error {
	fullPath := filepath.Join(d.root, path)
	return os.RemoveAll(fullPath)
}

func (d *LocalDriver) GetVisibility(path string) (string, error) {
	fullPath := filepath.Join(d.root, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}
	mode := info.Mode()
	if mode&0o044 != 0 {
		return "public", nil
	}
	return "private", nil
}

func (d *LocalDriver) SetVisibility(path string, visibility string) error {
	fullPath := filepath.Join(d.root, path)
	var mode os.FileMode = 0o600 // private
	if visibility == "public" {
		mode = 0o644
	}
	return os.Chmod(fullPath, mode)
}

func (d *LocalDriver) Reset() error {
	return os.RemoveAll(d.root)
}

func (d *LocalDriver) HealthCheck() error {
	_, err := os.Stat(d.root)
	return err
}

func (d *LocalDriver) Close() error {
	return nil
}
