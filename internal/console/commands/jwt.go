package command

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var jwtSecretEnvFile string

var jwtSecretCmd = &cobra.Command{
	Use:   "jwt:secret",
	Short: "Generate and replace JWT_SECRET in the environment file",
	Long: `Generate a new secure JWT secret and replace JWT_SECRET in the environment file.

This command is similar to Laravel Artisan key generation commands:
  fiber-starter jwt:secret
  fiber-starter jwt:secret --env .env.local`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		secret, err := generateJWTSecret()
		if err != nil {
			return err
		}

		if err := updateEnvFile(jwtSecretEnvFile, secret); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "JWT_SECRET updated in %s\n", jwtSecretEnvFile)
		return nil
	},
}

func generateJWTSecret() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

func updateEnvFile(envFile, newSecret string) error {
	envFile = strings.TrimSpace(envFile)
	if envFile == "" {
		envFile = ".env"
	}

	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf("%s file does not exist", envFile)
	}

	content, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if isJWTSecretLine(line) {
			lines[i] = fmt.Sprintf("JWT_SECRET=%s", newSecret)
			found = true
			break
		}
	}

	if !found {
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines[len(lines)-1] = fmt.Sprintf("JWT_SECRET=%s", newSecret)
		} else {
			lines = append(lines, fmt.Sprintf("JWT_SECRET=%s", newSecret))
		}
	}

	if err := os.WriteFile(envFile, []byte(strings.Join(lines, "\n")), 0o600); err != nil { //nolint:gosec
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

func isJWTSecretLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "JWT_SECRET=") || strings.HasPrefix(trimmed, "JWT_SECRET =")
}

func init() {
	jwtSecretCmd.Flags().StringVarP(&jwtSecretEnvFile, "env", "e", ".env", "environment file to update")
	rootCmd.AddCommand(jwtSecretCmd)
}
