package support

import (
	logging "fiber-starter/internal/providers/logging"
	"fiber-starter/internal/providers/logging/Contracts"
	"fiber-starter/internal/support/appctx"

	"go.uber.org/zap"
)

// This file provides a bridge/proxy for legacy code that still uses the support package
// for logging. It delegates all calls to the new Logging provider and facade.

// Logger is a global logger instance for backward compatibility.
// It will be updated when Init is called, or you can use the methods below which use the Facade.
var Logger = zap.NewNop()

// Init initializes the legacy support logger by setting the provider
func Init(p Contracts.Logger) {
	Logger = p.GetZapLogger()
}

// Channel returns a logger for the specified channel
func Channel(name string) Contracts.Logger {
	return logging.Facade().Channel(name)
}

// Debug logs a message at debug level
func Debug(msg string, fields ...zap.Field) {
	if appctx.App() == nil {
		Logger.Debug(msg, fields...)
		return
	}
	logging.Facade().Debug(msg, fields...)
}

// Info logs a message at info level
func Info(msg string, fields ...zap.Field) {
	if appctx.App() == nil {
		Logger.Info(msg, fields...)
		return
	}
	logging.Facade().Info(msg, fields...)
}

// Warn logs a message at warn level
func Warn(msg string, fields ...zap.Field) {
	if appctx.App() == nil {
		Logger.Warn(msg, fields...)
		return
	}
	logging.Facade().Warn(msg, fields...)
}

// Error logs a message at error level
func Error(msg string, fields ...zap.Field) {
	if appctx.App() == nil {
		Logger.Error(msg, fields...)
		return
	}
	logging.Facade().Error(msg, fields...)
}

// LogError is an alias for Error
func LogError(msg string, fields ...zap.Field) {
	Error(msg, fields...)
}

// Fatal logs a message at fatal level and exits
func Fatal(msg string, fields ...zap.Field) {
	if appctx.App() == nil {
		Logger.Fatal(msg, fields...)
		return
	}
	logging.Facade().Fatal(msg, fields...)
}

// Panic logs a message at panic level and panics
func Panic(msg string, fields ...zap.Field) {
	if appctx.App() == nil {
		Logger.Panic(msg, fields...)
		return
	}
	logging.Facade().Panic(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if appctx.App() == nil {
		return Logger.Sync()
	}
	return logging.Facade().Sync()
}
