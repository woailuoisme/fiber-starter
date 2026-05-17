package Contracts

// Manager defines the schedule manager contract.
type Manager interface {
	Scheduler() Scheduler
	Close() error
}
