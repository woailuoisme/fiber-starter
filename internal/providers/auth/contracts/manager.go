package contracts

// Manager defines the auth manager contract.
type Manager interface {
	Guard(name ...string) Guard
	SetModelCreator(provider string, creator func() any)
}
