package logging

import (
	"fiber-starter/internal/providers/logging/contracts"
	"fiber-starter/internal/support/appctx"

	"go.uber.org/zap"
)

// Facade returns the logging service from the application container.
// This allows for a more Laravel-like "Log::info()" experience if used correctly.
func Facade() contracts.Logger {
	// We use the raw appctx here to avoid circular dependencies with the providers package
	// if this were to be called from within other providers.
	rt := appctx.App()
	if rt == nil {
		// Fallback for cases where the app container isn't initialized
		return &Service{default_: zap.NewNop()}
	}

	// We assume the Log field exists on the runtime struct (which is duck-typed here via interface)
	type logProvider interface {
		LogService() contracts.Logger
	}

	if lp, ok := rt.(logProvider); ok {
		if service := lp.LogService(); service != nil {
			return service
		}
	}

	return &Service{default_: zap.NewNop()}
}

// Info is a shortcut for Facade().Info()
func Info(msg string, fields ...zap.Field) {
	Facade().Info(msg, fields...)
}

// Error is a shortcut for Facade().Error()
func Error(msg string, fields ...zap.Field) {
	Facade().Error(msg, fields...)
}

// Debug is a shortcut for Facade().Debug()
func Debug(msg string, fields ...zap.Field) {
	Facade().Debug(msg, fields...)
}

// Warn is a shortcut for Facade().Warn()
func Warn(msg string, fields ...zap.Field) {
	Facade().Warn(msg, fields...)
}

// Fatal is a shortcut for Facade().Fatal()
func Fatal(msg string, fields ...zap.Field) {
	Facade().Fatal(msg, fields...)
}

// Panic is a shortcut for Facade().Panic()
func Panic(msg string, fields ...zap.Field) {
	Facade().Panic(msg, fields...)
}

// Channel is a shortcut for Facade().Channel()
func Channel(name string) contracts.Logger {
	return Facade().Channel(name)
}

// With is a shortcut for Facade().With()
func With(fields ...zap.Field) contracts.Logger {
	return Facade().With(fields...)
}
