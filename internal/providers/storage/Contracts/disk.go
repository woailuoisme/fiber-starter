package contracts

import (
	"time"
)

// Disk defines the interface for file storage (similar to Laravel's Storage::disk())
type Disk interface {
	// Get retrieves the contents of a file
	Get(path string) ([]byte, error)

	// Put stores the contents of a file
	Put(path string, contents []byte, options ...interface{}) error

	// Exists checks if a file exists
	Exists(path string) (bool, error)

	// Missing checks if a file is missing
	Missing(path string) (bool, error)

	// Url returns the public URL for a file
	Url(path string) string

	// TemporaryUrl returns a temporary URL for a file
	TemporaryUrl(path string, expiration time.Duration) (string, error)

	// Size returns the size of a file in bytes
	Size(path string) (int64, error)

	// LastModified returns the last modified time of a file
	LastModified(path string) (time.Time, error)

	// Copy copies a file to a new location
	Copy(from, to string) error

	// Move moves a file to a new location
	Move(from, to string) error

	// Prepend prepends content to a file
	Prepend(path string, contents []byte) error

	// Append appends content to a file
	Append(path string, contents []byte) error

	// Delete removes a file or an array of files
	Delete(paths ...string) error

	// Files returns an array of files in a directory
	Files(directory string, recursive ...bool) ([]string, error)

	// AllFiles returns an array of all files in a directory (recursive)
	AllFiles(directory string) ([]string, error)

	// Directories returns an array of directories in a given directory
	Directories(directory string, recursive ...bool) ([]string, error)

	// AllDirectories returns an array of all directories in a given directory (recursive)
	AllDirectories(directory string) ([]string, error)

	// MakeDirectory creates a directory
	MakeDirectory(path string) error

	// DeleteDirectory deletes a directory
	DeleteDirectory(path string) error

	// GetVisibility returns the visibility of a file (public or private)
	GetVisibility(path string) (string, error)

	// SetVisibility sets the visibility of a file
	SetVisibility(path string, visibility string) error

	// Reset clears all files in the disk (use with caution)
	Reset() error

	// HealthCheck verifies the storage connection is alive
	HealthCheck() error

	// Close closes the storage connection
	Close() error
}
