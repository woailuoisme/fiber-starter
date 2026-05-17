package Contracts

// Manager defines the mail manager contract.
type Manager interface {
	Drive(name ...string) Mailer
	Close() error
}
