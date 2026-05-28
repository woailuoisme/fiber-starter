package mail

import (
	"lfiber/configs"
	mailContracts "lfiber/internal/providers/mail/contracts"
)

// Register initializes and returns the mail manager and the default mailer.
func Register(cfg *configs.Config) (mailContracts.Manager, mailContracts.Mailer, error) {
	manager := NewManager(cfg)
	if !cfg.Mail.Enabled {
		return manager, manager.Drive("noop"), nil
	}
	return manager, manager.Drive(), nil
}
