package backup

import "time"

type Config struct {
	AppName       string
	Disk          string
	Path          string
	TempPath      string
	Notifications NotificationConfig
	Binaries      BinaryConfig
}

type NotificationConfig struct {
	Enabled       bool
	NotifySuccess bool
	Channels      []string
	MailTo        string
}

type BinaryConfig struct {
	PgDump  string
	Psql    string
	SQLite3 string
}

type RunOptions struct {
	ConnectionName       string
	DisableNotifications bool
}

type RestoreOptions struct {
	ConnectionName       string
	Path                 string
	DisableNotifications bool
}

type ListOptions struct {
	ConnectionName string
}

type Backup struct {
	Path         string
	Connection   string
	Size         int64
	LastModified time.Time
}

type Result struct {
	Path       string
	Connection string
	Size       int64
}
