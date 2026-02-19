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

// jwtGenerateCmd represents the jwt:generate command
var jwtGenerateCmd = &cobra.Command{
	Use:   "jwt:generate",
	Short: "Generate and replace JWT secret",
	Long: `Generate a new secure JWT secret and automatically replace the JWT_SECRET value in .env file.
This command generates a 32-byte random key and updates it to the .env file.`,
	Run: func(_ *cobra.Command, _ []string) {
		generateAndReplaceJWTSecret()
	},
}

// generateJWTSecret Generate a secure JWT secret
func generateJWTSecret() (string, error) {
	// Generate 32 bytes of random data
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// Convert to base64 string
	secret := base64.StdEncoding.EncodeToString(key)
	return secret, nil
}

// updateEnvFile Update JWT_SECRET in .env file
func updateEnvFile(newSecret string) error {
	envFile := ".env"

	// Check if file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf(".env file does not exist")
	}

	// Read file content
	content, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	// Split content by lines
	lines := strings.Split(string(content), "\n")

	// Find and replace JWT_SECRET line
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

	// Write modified content back to file
	newContent := strings.Join(lines, "\n")
	err = os.WriteFile(envFile, []byte(newContent), 0600)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

// generateAndReplaceJWTSecret Generate new JWT secret and replace value in file
func generateAndReplaceJWTSecret() {
	color.Cyan("Generating new JWT secret...")

	// Generate new secret
	newSecret, err := generateJWTSecret()
	if err != nil {
		color.Red("Failed to generate secret: %v", err)
		os.Exit(1)
	}

	color.Green("New JWT secret generated: %s", newSecret)

	// Update .env file
	color.Yellow("Updating .env file...")
	err = updateEnvFile(newSecret)
	if err != nil {
		color.Red("Failed to update .env file: %v", err)
		os.Exit(1)
	}

	color.Green("JWT secret successfully updated in .env file")
	color.Yellow("Please restart the application for the new secret to take effect")
}

func init() {
	// Add jwt:generate command to root command
	rootCmd.AddCommand(jwtGenerateCmd)
}
