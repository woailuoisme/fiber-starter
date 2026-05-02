package command

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var jwtGenerateCmd = &cobra.Command{
	Use:   "jwt:generate",
	Short: "Generate and replace JWT secret",
	Long: `Generate a new secure JWT secret and automatically replace the JWT_SECRET value in .env file.
This command generates a 32-byte random key and updates it to the .env file.`,
	Run: func(_ *cobra.Command, _ []string) {
		generateAndReplaceJWTSecret()
	},
}

func generateJWTSecret() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

func updateEnvFile(newSecret string) error {
	const envFile = ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf(".env file does not exist")
	}

	content, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "JWT_SECRET=") {
			lines[i] = fmt.Sprintf("JWT_SECRET=%s", newSecret)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("JWT_SECRET not found in .env file")
	}

	if err := os.WriteFile(envFile, []byte(strings.Join(lines, "\n")), 0o600); err != nil { //nolint:gosec
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

func generateAndReplaceJWTSecret() {
	color.Cyan("Generating new JWT secret...")

	newSecret, err := generateJWTSecret()
	if err != nil {
		color.Red("Failed to generate secret: %v", err)
		os.Exit(1)
	}

	color.Green("New JWT secret generated: %s", newSecret)
	color.Yellow("Updating .env file...")

	if err := updateEnvFile(newSecret); err != nil {
		color.Red("Failed to update .env file: %v", err)
		os.Exit(1)
	}

	color.Green("JWT secret successfully updated in .env file")
	color.Yellow("Please restart the application for the new secret to take effect")
}

func init() {
	rootCmd.AddCommand(jwtGenerateCmd)
}
