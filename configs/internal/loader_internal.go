package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"gopkg.in/yaml.v3"
)

var loadEnvOnce sync.Once

func LoadEnvFile() {
	loadEnvOnce.Do(func() {
		appEnv := os.Getenv("APP_ENV")
		if appEnv == "" {
			appEnv = "development"
		}

		files := []string{fmt.Sprintf(".env.%s.local", appEnv)}
		if appEnv != "test" {
			files = append(files, ".env.local")
		}
		files = append(files, fmt.Sprintf(".env.%s", appEnv), ".env")

		loaded := false
		for _, file := range files {
			for _, path := range EnvFileCandidates(file) {
				if FileExists(path) {
					if err := godotenv.Load(path); err == nil {
						log.Printf("Successfully loaded environment file: %s", path) //nolint:gosec // environment file path is controlled by local project config
						loaded = true
						break
					}
				}
			}
		}

		if !loaded && strings.ToLower(strings.TrimSpace(os.Getenv("CONFIG_WARN_MISSING_ENV_FILE"))) != "false" {
			log.Printf("Environment file not found, will use environment variables and default configuration")
		}
	})
}

func FileExists(path string) bool {
	_, err := os.Stat(path) //nolint:gosec // path is derived from local config/env file lookup
	return !os.IsNotExist(err)
}

func EnvFileCandidates(file string) []string {
	return []string{
		file,
		filepath.Join("config", file),
		filepath.Join("configs", file),
		filepath.Join("..", file),
	}
}

func LoadConfigFiles(k *koanf.Koanf) error {
	dir := filepath.Join("configs", "yml")

	// Try to find the configs/yml directory by walking up to 3 levels (for tests)
	found := false
	for i := 0; i < 3; i++ {
		if FileExists(dir) {
			found = true
			break
		}
		dir = filepath.Join("..", dir)
	}

	if !found {
		return nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		path := filepath.Join(dir, file.Name())
		// nolint:gosec // path is derived from local configs directory
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// We use a simple environment variable expansion that supports ${VAR:default}
		expanded := ExpandEnv(string(data))

		var m map[string]any
		if err := yaml.Unmarshal([]byte(expanded), &m); err != nil {
			return fmt.Errorf("failed to parse %s: %w", file.Name(), err)
		}

		if err := k.Load(confmap.Provider(m, "."), nil); err != nil {
			return err
		}
	}
	return nil
}

func ExpandEnv(s string) string {
	// Regex for ${VAR:default}
	re := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		envVar := submatches[1]
		defaultVal := ""
		if len(submatches) > 2 {
			defaultVal = submatches[2]
		}

		val := os.Getenv(envVar)
		if val == "" {
			return defaultVal
		}
		return val
	})
}
