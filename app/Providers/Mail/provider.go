package mail

import (
	mailContracts "fiber-starter/app/Providers/Mail/Contracts"
	"fiber-starter/configs"
)

// Register initializes and returns the mail manager and the default mailer.
func Register(cfg *configs.Config) (mailContracts.Manager, mailContracts.Mailer, error) {
	if !cfg.Mail.Enabled {
		return nil, nil, nil
	}
	manager := NewManager(cfg)
	return manager, manager.Drive(), nil
}
