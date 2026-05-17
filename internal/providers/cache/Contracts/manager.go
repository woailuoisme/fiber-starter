package contracts

// Manager defines the cache manager contract.
type Manager interface {
	Store(name ...string) Store
	Close() error
}
