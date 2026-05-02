package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"
	"fiber-starter/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerStdoutMode(t *testing.T) {
	testkit.UseTempWorkDir(t)
	testkit.SetLoggerConfig(t, config.LoggerConfig{
		Level:  "info",
		Output: "stdout",
	})

	output := testkit.CaptureOutput(t, "stdout", func() {
		require.NoError(t, helpers.Init())
		helpers.Info("stdout-channel-test")
		assert.NoError(t, helpers.Sync())
	})

	assert.Contains(t, output, "stdout-channel-test")
	testkit.AssertNoLogDir(t)
}

func TestLoggerStderrMode(t *testing.T) {
	testkit.UseTempWorkDir(t)
	testkit.SetLoggerConfig(t, config.LoggerConfig{
		Level:  "info",
		Output: "stderr",
	})

	output := testkit.CaptureOutput(t, "stderr", func() {
		require.NoError(t, helpers.Init())
		helpers.Info("stderr-channel-test")
		assert.NoError(t, helpers.Sync())
	})

	assert.Contains(t, output, "stderr-channel-test")
	testkit.AssertNoLogDir(t)
}

func TestLoggerSingleMode(t *testing.T) {
	dir := testkit.UseTempWorkDir(t)
	testkit.SetLoggerConfig(t, config.LoggerConfig{
		Level:      "info",
		Output:     "single",
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   false,
	})

	require.NoError(t, helpers.Init())
	helpers.Info("single-channel-test")
	require.NoError(t, helpers.Sync())

	logFile := filepath.Join(dir, "storage", "logs", "app.log")
	data, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "single-channel-test")

	files := testkit.ListLogFiles(t, filepath.Join(dir, "storage", "logs"))
	assert.Equal(t, []string{"app.log"}, files)
}

func TestLoggerDailyMode(t *testing.T) {
	dir := testkit.UseTempWorkDir(t)
	now := time.Now()
	oldFile := filepath.Join(dir, "storage", "logs", fmt.Sprintf("app-%s.log", now.AddDate(0, 0, -2).Format("2006-01-02")))
	require.NoError(t, os.MkdirAll(filepath.Dir(oldFile), 0o755))
	require.NoError(t, os.WriteFile(oldFile, []byte("old log"), 0o600))

	testkit.SetLoggerConfig(t, config.LoggerConfig{
		Level:      "info",
		Output:     "daily",
		MaxAge:     1,
		MaxBackups: 3,
		Compress:   false,
	})

	require.NoError(t, helpers.Init())
	helpers.Info("daily-channel-test")
	require.NoError(t, helpers.Sync())

	dailyFile := filepath.Join(dir, "storage", "logs", fmt.Sprintf("app-%s.log", now.Format("2006-01-02")))
	data, err := os.ReadFile(dailyFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "daily-channel-test")

	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err))
}

func TestLoggerStackModeUsesStdoutAndDaily(t *testing.T) {
	dir := testkit.UseTempWorkDir(t)
	testkit.SetLoggerConfig(t, config.LoggerConfig{
		Level:      "info",
		Output:     "",
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   false,
	})

	output := testkit.CaptureOutput(t, "stdout", func() {
		require.NoError(t, helpers.Init())
		helpers.Info("stack-channel-test")
		assert.NoError(t, helpers.Sync())
	})

	assert.Contains(t, output, "stack-channel-test")

	dailyFile := filepath.Join(dir, "storage", "logs", fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
	data, err := os.ReadFile(dailyFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "stack-channel-test")
}
