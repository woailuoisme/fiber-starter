package backup

import (
	"fmt"
	"io"
	"strings"

	"lfiber/configs"
)

func dumpCommand(cfg configs.DBConnection, binaries BinaryConfig) (Command, error) {
	switch normalizedDriver(cfg.Driver) {
	case "postgres":
		args := []string{
			"--clean",
			"--if-exists",
			"--no-owner",
			"--no-privileges",
			"-h", cfg.Host,
			"-p", cfg.Port,
			"-U", cfg.Username,
			"-d", cfg.Database,
		}
		return Command{Name: binaryOrDefault(binaries.PgDump, "pg_dump"), Args: args, Env: passwordEnv(cfg)}, nil
	case "sqlite":
		return Command{Name: binaryOrDefault(binaries.SQLite3, "sqlite3"), Args: []string{cfg.Database, ".dump"}}, nil
	default:
		return Command{}, fmt.Errorf("unsupported backup database driver: %s", cfg.Driver)
	}
}

func restoreCommand(cfg configs.DBConnection, binaries BinaryConfig, input io.Reader) (Command, error) {
	switch normalizedDriver(cfg.Driver) {
	case "postgres":
		args := []string{"-h", cfg.Host, "-p", cfg.Port, "-U", cfg.Username, "-d", cfg.Database}
		return Command{Name: binaryOrDefault(binaries.Psql, "psql"), Args: args, Env: passwordEnv(cfg), Stdin: input}, nil
	case "sqlite":
		return Command{Name: binaryOrDefault(binaries.SQLite3, "sqlite3"), Args: []string{cfg.Database}, Stdin: input}, nil
	default:
		return Command{}, fmt.Errorf("unsupported restore database driver: %s", cfg.Driver)
	}
}

func normalizedDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "postgresql", "pgsql":
		return "postgres"
	case "sqlite", "sqlite3":
		return "sqlite"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func binaryOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func passwordEnv(cfg configs.DBConnection) []string {
	if strings.TrimSpace(cfg.Password) == "" {
		return nil
	}
	return []string{"PGPASSWORD=" + cfg.Password}
}
