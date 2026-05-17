package contracts

// Manager defines the queue manager contract.
type Manager interface {
	Drive(name ...string) Queue
	Close() error
}
