package Contracts

// StorageManager handles multiple storage disks
type StorageManager interface {
	// Disk returns a storage disk by name
	Disk(name ...string) Disk

	// Close all disks
	Close() error
}
