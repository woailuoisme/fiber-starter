package Contracts

// Manager defines the auth manager contract.
type Manager interface {
	Guard(name ...string) Guard
}
