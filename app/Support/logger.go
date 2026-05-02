// Package support 提供各种辅助函数和工具
package support

import (
	"errors"
	"os"
	"strings"

	"fiber-starter/app/Support/logger"
	"fiber-starter/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 全局日志实例
var Logger = zap.NewNop()

// Init 初始化日志配置
// Requirements: 17.1, 17.2, 17.3, 17.5, 17.6, 17.7, 22.11
func Init() error {
	built, err := logger.Build(config.GlobalConfig.Logger)
	if err != nil {
		return err
	}

	Logger = built
	return nil
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
	if err := Logger.Sync(); err != nil && !isIgnorableSyncError(err) {
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
