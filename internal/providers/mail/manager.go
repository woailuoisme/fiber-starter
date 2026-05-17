package mail

import (
	"errors"
	"sync"

	"fiber-starter/configs"
	"fiber-starter/internal/providers/mail/contracts"
	"fiber-starter/internal/providers/mail/drivers"
)

// Manager handles the lifecycle and selection of mail drivers
type Manager struct {
	config  *configs.Config
	mailers map[string]contracts.Mailer
	mu      sync.Mutex
}

// NewManager creates a new mail manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config:  cfg,
		mailers: make(map[string]contracts.Mailer),
	}
}

// Drive returns a specific mailer instance
func (m *Manager) Drive(name ...string) contracts.Mailer {
	m.mu.Lock()
	defer m.mu.Unlock()

	driver := m.config.Mail.Default
	if len(name) > 0 && name[0] != "" {
		driver = name[0]
	}

	if mailer, ok := m.mailers[driver]; ok {
		return mailer
	}

	var mailer contracts.Mailer
	switch driver {
	case "resend":
		mailer = drivers.NewResendDriver(m.config)
	case "log":
		mailer = drivers.NewLogDriver()
	case "smtp":
		mailer = drivers.NewSMTPDriver(m.config)
	default:
		mailer = drivers.NewResendDriver(m.config)
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
