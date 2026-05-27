package support

import (
	"strings"

	logging "lfiber/internal/providers/logging"
	"lfiber/internal/providers/logging/contracts"
	"lfiber/internal/support/appctx"

	"go.uber.org/zap"
)

// This file provides a bridge/proxy for legacy code that still uses the support package
// for logging. It delegates all calls to the new Logging provider and facade.

// Logger is a global logger instance for backward compatibility.
// It will be updated when Init is called, or you can use the methods below which use the Facade.
var Logger = zap.NewNop()

// Init initializes the legacy support logger by setting the provider
func Init(p contracts.Logger) {
	Logger = p.GetZapLogger()
}

// Channel returns a logger for the specified channel
func Channel(name string) contracts.Logger {
	return logging.Facade().Channel(name)
}

func writeLog(logFunc func(string, ...zap.Field), facadeFunc func(string, ...zap.Field), msg string, fields ...zap.Field) {
	if appctx.App() == nil {
		logFunc(msg, fields...)
		return
	}
	facadeFunc(msg, fields...)
}

// Debug logs a message at debug level
func Debug(msg string, fields ...zap.Field) {
	writeLog(Logger.Debug, logging.Facade().Debug, msg, fields...)
}

// Info logs a message at info level
func Info(msg string, fields ...zap.Field) {
	writeLog(Logger.Info, logging.Facade().Info, msg, fields...)
}

// Warn logs a message at warn level
func Warn(msg string, fields ...zap.Field) {
	writeLog(Logger.Warn, logging.Facade().Warn, msg, fields...)
}

// Error logs a message at error level
func Error(msg string, fields ...zap.Field) {
	writeLog(Logger.Error, logging.Facade().Error, msg, fields...)
}

// LogError is an alias for Error
func LogError(msg string, fields ...zap.Field) {
	Error(msg, fields...)
}

// Fatal logs a message at fatal level and exits
func Fatal(msg string, fields ...zap.Field) {
	writeLog(Logger.Fatal, logging.Facade().Fatal, msg, fields...)
}

// Panic logs a message at panic level and panics
func Panic(msg string, fields ...zap.Field) {
	writeLog(Logger.Panic, logging.Facade().Panic, msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	var err error
	if appctx.App() == nil {
		err = Logger.Sync()
	} else {
		err = logging.Facade().Sync()
	}

	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "bad file descriptor") ||
			strings.Contains(msg, "invalid argument") ||
			strings.Contains(msg, "inappropriate ioctl for device") ||
			strings.Contains(msg, "sync /dev/stdout") ||
			strings.Contains(msg, "sync /dev/stderr") {
			return nil
		}
		return err
	}
	return nil
}
