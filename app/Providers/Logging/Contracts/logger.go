package Contracts

import (
	"go.uber.org/zap"
)

// Logger defines the interface for the logging service
type Logger interface {
	// Channel returns a logger for the specified channel
	Channel(name string) Logger

	// Default returns the default logger
	Default() Logger

	// Methods for standard logging
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)

	// With adds structured context to the logger
	With(fields ...zap.Field) Logger

	// GetZapLogger returns the underlying zap.Logger
	GetZapLogger() *zap.Logger

	// Sync flushes any buffered log entries
	Sync() error
}
