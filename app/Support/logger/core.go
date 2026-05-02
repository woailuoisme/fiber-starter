package logger

import (
	"io"
	"os"
	"path/filepath"

	"fiber-starter/config"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func createConsoleCore(encoderConfig zapcore.EncoderConfig, level zapcore.Level, stream io.Writer) zapcore.Core {
	consoleEncoderConfig := encoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	return zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(stream),
		level,
	)
}

func createSingleFileCore(encoderConfig zapcore.EncoderConfig, level zapcore.Level, logConfig config.LoggerConfig) (zapcore.Core, error) {
	if err := ensureLogDir(); err != nil {
		return nil, err
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(DefaultLogDir, DefaultLogFile),
		MaxSize:    positiveOrDefault(logConfig.MaxSize, DefaultMaxSize),
		MaxBackups: positiveOrDefault(logConfig.MaxBackups, DefaultMaxBackups),
		MaxAge:     positiveOrDefault(logConfig.MaxAge, DefaultMaxAge),
		Compress:   logConfig.Compress,
		LocalTime:  true,
	}

	return newJSONCore(encoderConfig, level, fileWriter), nil
}

func createDailyFileCore(encoderConfig zapcore.EncoderConfig, level zapcore.Level, logConfig config.LoggerConfig) (zapcore.Core, error) {
	writer, err := newDailyLogWriter(
		DefaultLogDir,
		DefaultDailyLogPrefix,
		".log",
		positiveOrDefault(logConfig.MaxAge, DefaultMaxAge),
		positiveOrDefault(logConfig.MaxBackups, DefaultMaxBackups),
	)
	if err != nil {
		return nil, err
	}

	return newJSONCore(encoderConfig, level, writer), nil
}

func newJSONCore(encoderConfig zapcore.EncoderConfig, level zapcore.Level, writer io.Writer) zapcore.Core {
	fileEncoderConfig := encoderConfig
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	return zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(writer),
		level,
	)
}

func ensureLogDir() error {
	return os.MkdirAll(DefaultLogDir, LogDirPerm)
}
