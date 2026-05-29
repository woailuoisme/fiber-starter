package backup

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"lfiber/configs"
	databaseContracts "lfiber/internal/providers/database/contracts"
	notificationContracts "lfiber/internal/providers/notification/contracts"
	storageContracts "lfiber/internal/providers/storage/contracts"
)

type Service struct {
	cfg          Config
	appConfig    *configs.Config
	database     databaseContracts.Manager
	storage      storageContracts.StorageManager
	notification notificationContracts.Dispatcher
	runner       Runner
	now          func() time.Time
}

func NewService(cfg Config, appConfig *configs.Config, database databaseContracts.Manager, storage storageContracts.StorageManager, notification notificationContracts.Dispatcher, opts ...Option) *Service {
	s := &Service{
		cfg:          normalizeConfig(cfg),
		appConfig:    appConfig,
		database:     database,
		storage:      storage,
		notification: notification,
		runner:       ExecRunner{},
		now:          time.Now,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type Option func(*Service)

func WithRunner(runner Runner) Option {
	return func(s *Service) {
		if runner != nil {
			s.runner = runner
		}
	}
}

func WithClock(now func() time.Time) Option {
	return func(s *Service) {
		if now != nil {
			s.now = now
		}
	}
}

func (s *Service) Run(ctx context.Context, opts RunOptions) (Result, error) {
	connectionName, connCfg, err := s.connectionConfig(opts.ConnectionName)
	if err != nil {
		s.notify(ctx, "Backup failed", err.Error(), opts.DisableNotifications, false)
		return Result{}, err
	}

	tempDir, err := s.tempDir()
	if err != nil {
		s.notify(ctx, "Backup failed", err.Error(), opts.DisableNotifications, false)
		return Result{}, err
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	sqlPath := filepath.Join(tempDir, connectionName+".sql")
	// #nosec G304 -- sqlPath is generated under the configured backup temp directory.
	sqlFile, err := os.Create(sqlPath)
	if err != nil {
		return Result{}, err
	}
	cmd, err := dumpCommand(connCfg, s.cfg.Binaries)
	if err == nil {
		cmd.Stdout = sqlFile
		err = s.runner.Run(ctx, cmd)
	}
	closeErr := sqlFile.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		s.notify(ctx, "Backup failed", err.Error(), opts.DisableNotifications, false)
		return Result{}, err
	}

	compressed, err := gzipFile(sqlPath)
	if err != nil {
		s.notify(ctx, "Backup failed", err.Error(), opts.DisableNotifications, false)
		return Result{}, err
	}

	backupPath := s.backupPath(connectionName)
	if err := s.storage.Disk(s.cfg.Disk).Put(backupPath, compressed); err != nil {
		s.notify(ctx, "Backup failed", err.Error(), opts.DisableNotifications, false)
		return Result{}, err
	}

	result := Result{Path: backupPath, Connection: connectionName, Size: int64(len(compressed))}
	s.notify(ctx, "Backup completed", fmt.Sprintf("Database backup stored at %s", backupPath), opts.DisableNotifications, true)
	return result, nil
}

func (s *Service) Restore(ctx context.Context, opts RestoreOptions) error {
	connectionName, connCfg, err := s.connectionConfig(opts.ConnectionName)
	if err != nil {
		s.notify(ctx, "Restore failed", err.Error(), opts.DisableNotifications, false)
		return err
	}
	if strings.TrimSpace(opts.Path) == "" {
		return fmt.Errorf("backup path is required")
	}

	compressed, err := s.storage.Disk(s.cfg.Disk).Get(opts.Path)
	if err != nil {
		s.notify(ctx, "Restore failed", err.Error(), opts.DisableNotifications, false)
		return err
	}
	sqlBytes, err := gunzipBytes(compressed)
	if err != nil {
		s.notify(ctx, "Restore failed", err.Error(), opts.DisableNotifications, false)
		return err
	}

	if normalizedDriver(connCfg.Driver) == "sqlite" {
		if err := os.Remove(connCfg.Database); err != nil && !os.IsNotExist(err) {
			s.notify(ctx, "Restore failed", err.Error(), opts.DisableNotifications, false)
			return err
		}
	}

	cmd, err := restoreCommand(connCfg, s.cfg.Binaries, bytes.NewReader(sqlBytes))
	if err == nil {
		err = s.runner.Run(ctx, cmd)
	}
	if err != nil {
		s.notify(ctx, "Restore failed", err.Error(), opts.DisableNotifications, false)
		return err
	}
	s.notify(ctx, "Restore completed", fmt.Sprintf("Database backup %s restored to %s", opts.Path, connectionName), opts.DisableNotifications, true)
	return nil
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]Backup, error) {
	_ = ctx
	prefix := s.listPrefix(opts.ConnectionName)
	files, err := s.storage.Disk(s.cfg.Disk).AllFiles(prefix)
	if err != nil {
		return nil, err
	}
	backups := make([]Backup, 0, len(files))
	for _, file := range files {
		if !strings.HasSuffix(file, ".sql.gz") {
			continue
		}
		size, _ := s.storage.Disk(s.cfg.Disk).Size(file)
		modified, _ := s.storage.Disk(s.cfg.Disk).LastModified(file)
		backups = append(backups, Backup{
			Path:         file,
			Connection:   connectionFromPath(file),
			Size:         size,
			LastModified: modified,
		})
	}
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Path > backups[j].Path
	})
	return backups, nil
}

func (s *Service) connectionConfig(name string) (string, configs.DBConnection, error) {
	if s.appConfig == nil {
		return "", configs.DBConnection{}, fmt.Errorf("backup config is not initialized")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = s.appConfig.Database.Default
	}
	cfg, ok := s.appConfig.Database.Connections[name]
	if !ok {
		return "", configs.DBConnection{}, fmt.Errorf("database connection %q does not exist", name)
	}
	return name, cfg, nil
}

func (s *Service) tempDir() (string, error) {
	base := strings.TrimSpace(s.cfg.TempPath)
	if base == "" {
		base = filepath.Join(".cache", "backup")
	}
	dir := filepath.Join(base, s.now().Format("20060102150405"))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	return dir, nil
}

func (s *Service) backupPath(connection string) string {
	return path.Join(
		s.cfg.Path,
		s.cfg.AppName,
		connection,
		s.now().Format("20060102150405")+".sql.gz",
	)
}

func (s *Service) listPrefix(connection string) string {
	prefix := path.Join(s.cfg.Path, s.cfg.AppName)
	if strings.TrimSpace(connection) != "" {
		prefix = path.Join(prefix, strings.TrimSpace(connection))
	}
	return prefix
}

func (s *Service) notify(_ context.Context, title, message string, disabled bool, success bool) {
	if disabled || s.notification == nil || !s.cfg.Notifications.Enabled {
		return
	}
	if success && !s.cfg.Notifications.NotifySuccess {
		return
	}
	_ = s.notification.Send(Notifiable{MailTo: s.cfg.Notifications.MailTo}, Notification{
		Channels: notificationChannels(s.cfg.Notifications.Channels),
		Title:    title,
		Message:  message,
		Priority: 5,
	})
}

func normalizeConfig(cfg Config) Config {
	if cfg.AppName == "" {
		cfg.AppName = "lfiber"
	}
	if cfg.Disk == "" {
		cfg.Disk = "local"
	}
	if cfg.Path == "" {
		cfg.Path = "backups"
	}
	if cfg.TempPath == "" {
		cfg.TempPath = filepath.Join(".cache", "backup")
	}
	if cfg.Notifications.Channels == nil {
		cfg.Notifications.Channels = []string{"mail"}
	}
	return cfg
}

func gzipFile(file string) ([]byte, error) {
	// #nosec G304 -- file is the service-generated SQL dump path.
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	writer := gzip.NewWriter(&out)
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func gunzipBytes(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()
	var out bytes.Buffer
	if _, err := out.ReadFrom(reader); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func connectionFromPath(file string) string {
	parts := strings.Split(file, "/")
	if len(parts) < 3 {
		return ""
	}
	return parts[len(parts)-2]
}
