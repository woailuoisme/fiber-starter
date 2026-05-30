package contracts

// Manager defines the schedule manager contract.
type Manager interface {
	Scheduler() (Scheduler, error)
	Close() error
}
