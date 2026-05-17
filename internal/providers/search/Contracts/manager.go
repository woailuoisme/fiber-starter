package contracts

// Manager defines the search manager contract.
type Manager interface {
	Drive(name ...string) Engine
	Close() error
}
