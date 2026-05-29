package backup

import (
	"strings"

	"lfiber/configs"
	databaseContracts "lfiber/internal/providers/database/contracts"
	notificationContracts "lfiber/internal/providers/notification/contracts"
	storageContracts "lfiber/internal/providers/storage/contracts"
	backuppkg "lfiber/pkg/backup"
)

func Register(cfg *configs.Config, database databaseContracts.Manager, storage storageContracts.StorageManager, notification notificationContracts.Dispatcher) (*backuppkg.Service, error) {
	backupCfg := backuppkg.Config{
		AppName:  cfg.App.Name,
		Disk:     cfg.Backup.Disk,
		Path:     cfg.Backup.Path,
		TempPath: cfg.Backup.TempPath,
		Notifications: backuppkg.NotificationConfig{
			Enabled:       cfg.Backup.Notifications.Enabled,
			NotifySuccess: cfg.Backup.Notifications.NotifySuccess,
			Channels:      normalizeChannels(cfg.Backup.Notifications.Channels),
			MailTo:        cfg.Backup.Notifications.MailTo,
		},
		Binaries: backuppkg.BinaryConfig{
			PgDump:  cfg.Backup.Binaries.PgDump,
			Psql:    cfg.Backup.Binaries.Psql,
			SQLite3: cfg.Backup.Binaries.SQLite3,
		},
	}
	return backuppkg.NewService(backupCfg, cfg, database, storage, notification), nil
}

func normalizeChannels(channels []string) []string {
	var out []string
	for _, channel := range channels {
		for _, part := range strings.Split(channel, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	if len(out) == 0 {
		return []string{"mail"}
	}
	return out
}
