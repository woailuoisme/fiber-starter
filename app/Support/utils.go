package support

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/fatih/color"
)

// UtcNow returns the current time in UTC using carbon
func UtcNow() time.Time {
	return carbon.Now(carbon.UTC).StdTime()
}

// NormalizePagination normalizes pagination parameters (page, limit) and returns page, limit, offset
func NormalizePagination(page, limit int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return page, limit, (page - 1) * limit
}

// SplitAndTrim splits a string by a separator and trims each element
func SplitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// IsConnectionError checks if an error is a connection-related error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "dial tcp") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "operation not permitted")
}

// HandleServerStartError provides user-friendly error reporting for server startup failures
func HandleServerStartError(err error, port string) {
	if err == nil {
		return
	}

	errStr := err.Error()
	if strings.Contains(errStr, "server_port_in_use") {
		fmt.Println()
		_, _ = color.New(color.FgRed, color.Bold).Printf("  Error: Port %s is already in use!\n", port)
		_, _ = color.New(color.FgWhite).Printf("  The application failed to start because the configured port is occupied.\n\n")

		_, _ = color.New(color.FgYellow).Printf("  Suggestions:\n")
		fmt.Printf("  1. Kill the process using the port:\n")
		_, _ = color.New(color.FgCyan).Printf("     kill -9 $(lsof -t -i:%s)\n\n", port)
		fmt.Printf("  2. Or change the APP_PORT in your .env file.\n\n")

		os.Exit(1)
	}

	// For other errors, fall back to standard fatal logging (handled by caller if preferred, or here)
	// But let's keep it consistent and just handle the port one here as requested.
}
