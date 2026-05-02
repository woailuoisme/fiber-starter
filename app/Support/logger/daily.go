package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type dailyLogWriter struct {
	dir           string
	prefix        string
	suffix        string
	maxAge        int
	maxBackups    int
	clock         func() time.Time
	mu            sync.Mutex
	file          *os.File
	currentDay    string
	lastCleanedAt string
}

func newDailyLogWriter(dir, prefix, suffix string, maxAge, maxBackups int) (*dailyLogWriter, error) {
	if err := os.MkdirAll(dir, LogDirPerm); err != nil {
		return nil, err
	}

	writer := &dailyLogWriter{
		dir:        dir,
		prefix:     prefix,
		suffix:     suffix,
		maxAge:     maxAge,
		maxBackups: maxBackups,
		clock:      time.Now,
	}

	if err := writer.cleanupLocked(); err != nil {
		return nil, err
	}

	return writer, nil
}

func (w *dailyLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateLocked(); err != nil {
		return 0, err
	}

	return w.file.Write(p)
}

func (w *dailyLogWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	return w.file.Sync()
}

func (w *dailyLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}

func (w *dailyLogWriter) rotateLocked() error {
	currentDay := w.clock().Format("2006-01-02")
	if w.file != nil && w.currentDay == currentDay {
		return nil
	}

	if w.file != nil {
		_ = w.file.Sync()
		_ = w.file.Close()
		w.file = nil
	}

	if err := os.MkdirAll(w.dir, LogDirPerm); err != nil {
		return err
	}

	filePath := filepath.Join(w.dir, w.dailyFilename(currentDay))
	// #nosec G304 -- 日志文件路径由固定前缀和日期组成，目录受控且不来自外部输入。
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	w.file = file
	w.currentDay = currentDay

	if w.lastCleanedAt != currentDay {
		if err := w.cleanupLocked(); err != nil {
			return err
		}
		w.lastCleanedAt = currentDay
	}

	return nil
}

func (w *dailyLogWriter) dailyFilename(day string) string {
	return fmt.Sprintf("%s-%s%s", w.prefix, day, w.suffix)
}

func (w *dailyLogWriter) cleanupLocked() error {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	type logFile struct {
		path string
		day  time.Time
	}

	var files []logFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, w.prefix+"-") || !strings.HasSuffix(name, w.suffix) {
			continue
		}

		dayPart := strings.TrimSuffix(strings.TrimPrefix(name, w.prefix+"-"), w.suffix)
		day, err := time.Parse("2006-01-02", dayPart)
		if err != nil {
			continue
		}

		files = append(files, logFile{
			path: filepath.Join(w.dir, name),
			day:  day,
		})
	}

	if len(files) == 0 {
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].day.After(files[j].day)
	})

	cutoff := time.Time{}
	if w.maxAge > 0 {
		cutoff = w.clock().AddDate(0, 0, -w.maxAge)
	}

	for i, file := range files {
		if w.maxAge > 0 && file.day.Before(cutoff) {
			_ = os.Remove(file.path)
			continue
		}

		if w.maxBackups > 0 && i >= w.maxBackups {
			_ = os.Remove(file.path)
		}
	}

	return nil
}
