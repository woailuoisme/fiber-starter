package logging

import (
	"errors"
	"os"
	"strings"

	"fiber-starter/configs"
	loggingContracts "fiber-starter/internal/providers/logging/contracts"

	"go.uber.org/zap"
)

type Service struct {
	loggers  map[string]*zap.Logger
	default_ *zap.Logger
	config   configs.LoggerConfig
}

// Register initializes the logging service and returns the logger contract.
func Register(cfg configs.LoggerConfig) (loggingContracts.Logger, error) {
	built, err := Build(cfg)
	if err != nil {
		// Fallback to a basic logger if build fails
		built = zap.NewNop()
	}

	service := &Service{
		loggers:  make(map[string]*zap.Logger),
		default_: built,
		config:   cfg,
	}
	service.loggers["default"] = built
	return service, nil
}

// Default returns the default logger
func (s *Service) Default() loggingContracts.Logger {
	return s
}

// Channel returns a logger for the specified channel (currently returns default or creates a new one)
func (s *Service) Channel(name string) loggingContracts.Logger {
	if l, ok := s.loggers[name]; ok {
		return &Service{
			loggers:  s.loggers,
			default_: l,
			config:   s.config,
		}
	}

	return s
}

// Debug logs a message at debug level
func (s *Service) Debug(msg string, fields ...zap.Field) {
	s.default_.Debug(msg, fields...)
}

// Info logs a message at info level
func (s *Service) Info(msg string, fields ...zap.Field) {
	s.default_.Info(msg, fields...)
}

// Warn logs a message at warn level
func (s *Service) Warn(msg string, fields ...zap.Field) {
	s.default_.Warn(msg, fields...)
}

// Error logs a message at error level
func (s *Service) Error(msg string, fields ...zap.Field) {
	s.default_.Error(msg, fields...)
}

// Fatal logs a message at fatal level
func (s *Service) Fatal(msg string, fields ...zap.Field) {
	s.default_.Fatal(msg, fields...)
}

// Panic logs a message at panic level
func (s *Service) Panic(msg string, fields ...zap.Field) {
	s.default_.Panic(msg, fields...)
}

// With adds structured context to the logger
func (s *Service) With(fields ...zap.Field) loggingContracts.Logger {
	return &Service{
		loggers:  s.loggers,
		default_: s.default_.With(fields...),
		config:   s.config,
	}
}

// GetZapLogger returns the underlying zap.Logger
func (s *Service) GetZapLogger() *zap.Logger {
	return s.default_
}

// Sync flushes any buffered log entries
func (s *Service) Sync() error {
	if err := s.default_.Sync(); err != nil && !isIgnorableSyncError(err) {
		return err
	}
	return nil
}

func isIgnorableSyncError(err error) bool {
	if err == nil {
		return true
	}

	if errors.Is(err, os.ErrInvalid) {
		return true
	}

	msg := err.Error()
	return strings.Contains(msg, "sync /dev/stdout") ||
		strings.Contains(msg, "sync /dev/stderr") ||
		strings.Contains(msg, "bad file descriptor")
}
