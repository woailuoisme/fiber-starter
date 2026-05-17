package testkit

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// CaptureOutput captures stdout or stderr while fn executes and returns the captured text.
func CaptureOutput(t *testing.T, stream string, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe failed: %v", err)
	}

	switch stream {
	case "stdout":
		prev := os.Stdout
		os.Stdout = w
		defer func() { os.Stdout = prev }()
	case "stderr":
		prev := os.Stderr
		os.Stderr = w
		defer func() { os.Stderr = prev }()
	default:
		t.Fatalf("unsupported stream %q", stream)
	}

	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		outCh <- buf.String()
	}()

	fn()
	_ = w.Close()
	return <-outCh
}

// AssertNoLogDir fails the test if storage/logs exists in the current working directory.
func AssertNoLogDir(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(filepath.Join("storage", "logs")); !os.IsNotExist(err) {
		t.Fatalf("expected no log directory, got err=%v", err)
	}
}

// ListLogFiles returns regular files in a log directory.
func ListLogFiles(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read log dir failed: %v", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files
}
