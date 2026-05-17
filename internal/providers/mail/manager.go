package mail

import (
	"errors"
	"sync"

	"fiber-starter/configs"
	"fiber-starter/internal/providers/mail/Contracts"
	"fiber-starter/internal/providers/mail/Drivers"
)

// Manager handles the lifecycle and selection of mail drivers
type Manager struct {
	config  *configs.Config
	mailers map[string]Contracts.Mailer
	mu      sync.Mutex
}

// NewManager creates a new mail manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config:  cfg,
		mailers: make(map[string]Contracts.Mailer),
	}
}

// Drive returns a specific mailer instance
func (m *Manager) Drive(name ...string) Contracts.Mailer {
	m.mu.Lock()
	defer m.mu.Unlock()

	driver := m.config.Mail.Default
	if len(name) > 0 && name[0] != "" {
		driver = name[0]
	}

	if mailer, ok := m.mailers[driver]; ok {
		return mailer
	}

	var mailer Contracts.Mailer
	switch driver {
	case "resend":
		mailer = Drivers.NewResendDriver(m.config)
	case "log":
		mailer = Drivers.NewLogDriver()
	case "smtp":
		mailer = Drivers.NewSMTPDriver(m.config)
	default:
		mailer = Drivers.NewResendDriver(m.config)
	}

	m.mailers[driver] = mailer
	return mailer
}

// Close closes all cached mailers.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, mailer := range m.mailers {
		if err := mailer.Close(); err != nil {
			errs = append(errs, errors.New(name+": "+err.Error()))
		}
	}

	return errors.Join(errs...)
}
