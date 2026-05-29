package providers_test

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"lfiber/configs"
	notificationContracts "lfiber/internal/providers/notification/contracts"
	storageProvider "lfiber/internal/providers/storage"
	"lfiber/pkg/backup"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type backupRunner struct {
	commands []backup.Command
	stdin    string
}

func (r *backupRunner) Run(_ context.Context, cmd backup.Command) error {
	r.commands = append(r.commands, cmd)
	if cmd.Stdin != nil {
		data, _ := io.ReadAll(cmd.Stdin)
		r.stdin = string(data)
	}
	if cmd.Stdout != nil {
		_, _ = cmd.Stdout.Write([]byte("CREATE TABLE users (id integer);\n"))
	}
	return nil
}

type backupDispatcher struct {
	sent []notificationContracts.Notification
}

func (d *backupDispatcher) Send(_ interface{}, n notificationContracts.Notification) error {
	d.sent = append(d.sent, n)
	return nil
}

func newBackupTestService(t *testing.T, runner *backupRunner, dispatcher *backupDispatcher, notifySuccess bool) (*backup.Service, *configs.Config) {
	t.Helper()
	cfg := &configs.Config{}
	cfg.App.Name = "lfiber"
	cfg.Database.Default = "sqlite"
	cfg.Database.Connections = map[string]configs.DBConnection{
		"sqlite": {
			Driver:   "sqlite",
			Database: filepath.Join(t.TempDir(), "database.sqlite"),
		},
	}
	cfg.Storage.Driver = "local"
	cfg.Storage.Local = &configs.LocalStorageConfig{Root: t.TempDir(), URL: "/storage"}
	storage, err := storageProvider.Register(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = storage.Close() })

	service := backup.NewService(backup.Config{
		AppName:  "lfiber",
		Disk:     "local",
		Path:     "backups",
		TempPath: filepath.Join(t.TempDir(), "backup-temp"),
		Notifications: backup.NotificationConfig{
			Enabled:       true,
			NotifySuccess: notifySuccess,
			Channels:      []string{"mail"},
			MailTo:        "ops@example.com",
		},
		Binaries: backup.BinaryConfig{SQLite3: "sqlite3"},
	}, cfg, nil, storage, dispatcher, backup.WithRunner(runner), backup.WithClock(func() time.Time {
		return time.Date(2026, 5, 29, 16, 30, 0, 0, time.UTC)
	}))
	return service, cfg
}

func TestBackupService_RunStoresGzippedSQL(t *testing.T) {
	runner := &backupRunner{}
	dispatcher := &backupDispatcher{}
	service, _ := newBackupTestService(t, runner, dispatcher, true)

	result, err := service.Run(context.Background(), backup.RunOptions{})
	require.NoError(t, err)

	assert.Equal(t, "backups/lfiber/sqlite/20260529163000.sql.gz", result.Path)
	require.Len(t, runner.commands, 1)
	assert.Equal(t, "sqlite3", runner.commands[0].Name)
	require.Len(t, runner.commands[0].Args, 2)
	assert.Equal(t, ".dump", runner.commands[0].Args[1])
	assert.NotZero(t, result.Size)
	assert.Len(t, dispatcher.sent, 1)

	backups, err := service.List(context.Background(), backup.ListOptions{})
	require.NoError(t, err)
	require.Len(t, backups, 1)
	assert.Equal(t, result.Path, backups[0].Path)
	assert.Equal(t, "sqlite", backups[0].Connection)
}

func TestBackupService_RestoreReadsGzippedBackup(t *testing.T) {
	runner := &backupRunner{}
	dispatcher := &backupDispatcher{}
	service, _ := newBackupTestService(t, runner, dispatcher, false)

	result, err := service.Run(context.Background(), backup.RunOptions{DisableNotifications: true})
	require.NoError(t, err)
	runner.commands = nil

	err = service.Restore(context.Background(), backup.RestoreOptions{Path: result.Path, DisableNotifications: true})
	require.NoError(t, err)

	require.Len(t, runner.commands, 1)
	assert.Equal(t, "sqlite3", runner.commands[0].Name)
	assert.Equal(t, "CREATE TABLE users (id integer);\n", runner.stdin)
	assert.Empty(t, dispatcher.sent)
}

func TestBackupNotification_ViaFallsBackToMail(t *testing.T) {
	n := backup.Notification{Title: "Backup failed", Message: "dump failed"}
	assert.Equal(t, []string{"mail"}, n.Via(nil))
	msg := n.ToGotify(nil)
	assert.Equal(t, "Backup failed", msg.Title)
	assert.Equal(t, "dump failed", msg.Message)
}

func TestBackupNotification_UsesConfiguredChannels(t *testing.T) {
	n := backup.Notification{Channels: []string{"gotify", "telegram"}}
	assert.Equal(t, []string{"gotify", "telegram"}, n.Via(nil))
}
