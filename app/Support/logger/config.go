package logger

import (
	"os"
	"strings"

	"fiber-starter/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// DefaultLogDir 默认日志目录
	DefaultLogDir = "./storage/logs"
	// DefaultLogFile 默认单文件日志文件名
	DefaultLogFile = "app.log"
	// DefaultDailyLogPrefix 默认按天日志前缀
	DefaultDailyLogPrefix = "app"
	// DefaultMaxSize 默认日志文件最大大小 (MB)
	DefaultMaxSize = 100
	// DefaultMaxBackups 默认日志文件最大备份数
	DefaultMaxBackups = 30
	// DefaultMaxAge 默认日志文件最大保存时间 (天)
	DefaultMaxAge = 90
	// LogDirPerm 日志目录权限
	LogDirPerm = 0o755
)

const (
	logOutputStdout = "stdout"
	logOutputStderr = "stderr"
	logOutputSingle = "single"
	logOutputDaily  = "daily"
	logOutputStack  = "stack"
)

// Build 初始化并返回 Logger
func Build(logConfig config.LoggerConfig) (*zap.Logger, error) {
	level := getLogLevel(logConfig.Level)
	encoderConfig := createEncoderConfig()

	cores, err := buildLoggerCores(logConfig, encoderConfig, level)
	if err != nil {
		return nil, err
	}

	return zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	), nil
}

func buildLoggerCores(logConfig config.LoggerConfig, encoderConfig zapcore.EncoderConfig, level zapcore.Level) ([]zapcore.Core, error) {
	switch normalizeLogOutput(logConfig.Output) {
	case logOutputStdout:
		return []zapcore.Core{createConsoleCore(encoderConfig, level, os.Stdout)}, nil
	case logOutputStderr:
		return []zapcore.Core{createConsoleCore(encoderConfig, level, os.Stderr)}, nil
	case logOutputSingle:
		core, err := createSingleFileCore(encoderConfig, level, logConfig)
		if err != nil {
			return nil, err
		}
		return []zapcore.Core{core}, nil
	case logOutputDaily:
		core, err := createDailyFileCore(encoderConfig, level, logConfig)
		if err != nil {
			return nil, err
		}
		return []zapcore.Core{core}, nil
	case logOutputStack:
		dailyCore, err := createDailyFileCore(encoderConfig, level, logConfig)
		if err != nil {
			return nil, err
		}
		return []zapcore.Core{
			createConsoleCore(encoderConfig, level, os.Stdout),
			dailyCore,
		}, nil
	default:
		return []zapcore.Core{createConsoleCore(encoderConfig, level, os.Stdout)}, nil
	}
}

func normalizeLogOutput(output string) string {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "", "stack", "both":
		return logOutputStack
	case "stdout":
		return logOutputStdout
	case "stderr":
		return logOutputStderr
	case "single", "file":
		return logOutputSingle
	case "daily":
		return logOutputDaily
	default:
		return logOutputStack
	}
}

func createEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "message"
	encoderConfig.CallerKey = "caller"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	return encoderConfig
}

func getLogLevel(levelStr string) zapcore.Level {
	levelStr = strings.ToLower(strings.TrimSpace(levelStr))
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func positiveOrDefault(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
