// Package support 提供各种辅助函数和工具
package support

import (
	"os"
	"path/filepath"
	"strings"

	"fiber-starter/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger = zap.NewNop()

const (
	// DefaultLogDir 默认日志目录
	DefaultLogDir = "./storage/logs"
	// DefaultLogFile 默认日志文件名
	DefaultLogFile = "app.log"
	// DefaultMaxSize 默认日志文件最大大小 (MB)
	DefaultMaxSize = 100
	// DefaultMaxBackups 默认日志文件最大备份数
	DefaultMaxBackups = 30
	// DefaultMaxAge 默认日志文件最大保存时间 (天)
	DefaultMaxAge = 90
	// LogDirPerm 日志目录权限
	LogDirPerm = 0755
)

// Init 初始化日志配置
// Requirements: 17.1, 17.2, 17.3, 17.5, 17.6, 17.7, 22.11
func Init() error {
	// 获取日志配置
	logConfig := config.GlobalConfig.Logger

	// 设置日志级别
	// Requirements: 17.6
	level := getLogLevel(logConfig.Level)

	// 创建编码器配置
	// Requirements: 17.5
	encoderConfig := createEncoderConfig()

	// 创建多个核心以支持不同的输出格式
	var cores []zapcore.Core

	// 输出到文件（带日志轮转）
	// Requirements: 17.7
	if logConfig.Output == "file" || logConfig.Output == "both" {
		fileCore, err := createFileCore(encoderConfig, level)
		if err != nil {
			return err
		}
		cores = append(cores, fileCore)
	}

	// 输出到标准输出
	if logConfig.Output == "stdout" || logConfig.Output == "both" || logConfig.Output == "" {
		cores = append(cores, createConsoleCore(encoderConfig, level))
	}

	// 如果没有配置任何输出，默认输出到标准输出（彩色）
	if len(cores) == 0 {
		cores = append(cores, createConsoleCore(encoderConfig, level))
	}

	// 创建核心
	core := zapcore.NewTee(cores...)

	// 创建Logger实例
	// Requirements: 17.1, 17.2, 17.3
	Logger = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return nil
}

// createEncoderConfig 创建编码器配置
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

// createFileCore 创建文件输出核心
func createFileCore(encoderConfig zapcore.EncoderConfig, level zapcore.Level) (zapcore.Core, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(DefaultLogDir, LogDirPerm); err != nil {
		return nil, err
	}

	// 使用 lumberjack 实现日志轮转
	// Requirements: 17.7
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(DefaultLogDir, DefaultLogFile),
		MaxSize:    DefaultMaxSize,    // MB
		MaxBackups: DefaultMaxBackups, // 保留30个备份
		MaxAge:     DefaultMaxAge,     // 保留90天
		Compress:   true,
		LocalTime:  true,
	}

	// 文件输出使用 JSON 格式
	fileEncoderConfig := encoderConfig
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	return zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(fileWriter),
		level,
	), nil
}

// createConsoleCore 创建控制台输出核心
func createConsoleCore(encoderConfig zapcore.EncoderConfig, level zapcore.Level) zapcore.Core {
	// 控制台输出使用彩色格式
	consoleEncoderConfig := encoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	return zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		level,
	)
}

// getLogLevel 根据字符串获取日志级别
// Requirements: 17.6
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
		return zapcore.InfoLevel // 默认使用info级别
	}
}

// Debug 记录调试日志
func Debug(msg string, fields ...zapcore.Field) {
	Logger.Debug(msg, fields...)
}

// Info 记录信息日志
func Info(msg string, fields ...zapcore.Field) {
	Logger.Info(msg, fields...)
}

// Warn 记录警告日志
func Warn(msg string, fields ...zapcore.Field) {
	Logger.Warn(msg, fields...)
}

// LogError 记录错误日志
func LogError(msg string, fields ...zapcore.Field) {
	Logger.Error(msg, fields...)
}

// Error 记录错误日志（LogError 的别名）
func Error(msg string, fields ...zapcore.Field) {
	Logger.Error(msg, fields...)
}

// DPanic 记录严重错误日志，开发环境会触发panic
func DPanic(msg string, fields ...zapcore.Field) {
	Logger.DPanic(msg, fields...)
}

// Panic 记录严重错误日志并触发panic
func Panic(msg string, fields ...zapcore.Field) {
	Logger.Panic(msg, fields...)
}

// Fatal 记录致命错误日志并退出程序
func Fatal(msg string, fields ...zapcore.Field) {
	Logger.Fatal(msg, fields...)
}

// Sync 刷新日志缓冲区
func Sync() error {
	return Logger.Sync()
}
