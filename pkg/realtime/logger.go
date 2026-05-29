package realtime

import (
	"fmt"
	"log"
)

type noopLogger struct{}

func (noopLogger) Info(msg string, fields ...any)  {}
func (noopLogger) Warn(msg string, fields ...any)  {}
func (noopLogger) Error(msg string, fields ...any) {}

type stdoutLogger struct{}

func (stdoutLogger) Info(msg string, fields ...any) {
	log.Printf("[INFO] %s %v", msg, fmt.Sprint(fields...))
}

func (stdoutLogger) Warn(msg string, fields ...any) {
	log.Printf("[WARN] %s %v", msg, fmt.Sprint(fields...))
}

func (stdoutLogger) Error(msg string, fields ...any) {
	log.Printf("[ERROR] %s %v", msg, fmt.Sprint(fields...))
}

// NewNoopLogger 返回空实现的 logger，防止外部未提供时空指针崩溃
func NewNoopLogger() Logger {
	return noopLogger{}
}

// NewStdoutLogger 返回控制台打印的标准 logger
func NewStdoutLogger() Logger {
	return stdoutLogger{}
}
