package commandutil

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"lfiber/configs"
	providers "lfiber/internal/providers"
	database "lfiber/internal/providers/database"
	"lfiber/internal/support"

	"github.com/spf13/cobra"
)

var (
	RuntimeBuilder = providers.Build
	AtlasRunner    = runAtlas
)

func BuildRuntime() (*providers.Runtime, error) {
	return RuntimeBuilder()
}

func CloseRuntime(rt *providers.Runtime) error {
	if rt == nil {
		return nil
	}
	if err := rt.Close(); err != nil {
		return err
	}
	return support.Sync()
}

func WaitForInterrupt() <-chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	return quit
}

func ParsePositiveInt(value string, fallback int) int {
	var parsed int
	if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed); err == nil && parsed > 0 {
		return parsed
	}
	return fallback
}

func GetDefaultConnection() (configs.DBConnection, error) {
	cfg, _, err := configs.LoadConfig()
	if err != nil {
		return configs.DBConnection{}, fmt.Errorf("failed to load config: %w", err)
	}

	dbConfig := &cfg.Database
	defaultConn := dbConfig.Default
	connConfig, exists := dbConfig.Connections[defaultConn]
	if !exists {
		return configs.DBConnection{}, fmt.Errorf("database connection config %q does not exist", defaultConn)
	}

	return connConfig, nil
}

func InitDBWithConfig() (*sql.DB, *configs.Config, error) {
	cfg, _, err := configs.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	conn := database.NewConnection(cfg)
	db, err := conn.GetDB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, cfg, nil
}

func MigrationEnvName(driver string) string {
	if IsSQLiteDriver(driver) {
		return "sqlite"
	}
	return "postgres"
}

func RunAtlasForCurrentConnection(args ...string) error {
	connConfig, err := GetDefaultConnection()
	if err != nil {
		return err
	}
	return RunAtlasForConnection(connConfig, args...)
}

func RunAtlasForConnection(connConfig configs.DBConnection, args ...string) error {
	fullArgs := append([]string{}, args...)
	fullArgs = append(fullArgs, "--env", MigrationEnvName(connConfig.Driver))
	return RunAtlas(fullArgs...)
}

func RunAtlas(args ...string) error {
	return AtlasRunner(args...)
}

func runAtlas(args ...string) error {
	// #nosec G204 -- atlas is an expected local developer tool invoked with controlled args.
	cmd := exec.Command("atlas", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func IsSQLiteDriver(driver string) bool {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "sqlite", "sqlite3":
		return true
	default:
		return false
	}
}

func NoInteraction(cmd *cobra.Command) bool {
	flag := cmd.Flag("no-interaction")
	if flag == nil {
		return false
	}
	value, err := strconv.ParseBool(flag.Value.String())
	return err == nil && value
}
