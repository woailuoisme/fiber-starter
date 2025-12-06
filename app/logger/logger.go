package logger

import (
	"os"
	"path/filepath"
	_ "path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"fiber-starter/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger *zap.Logger

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
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "message"
	encoderConfig.CallerKey = "caller"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	// 根据配置选择编码器
	// Requirements: 17.5
	var encoder zapcore.Encoder
	if logConfig.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建输出目标
	var writers []zapcore.WriteSyncer

	// 输出到文件（带日志轮转）
	// Requirements: 17.7
	if logConfig.Output == "file" || logConfig.Output == "both" {
		// 确保日志目录存在
		logDir := "./storage/logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		//使用 lumberjack 实现日志轮转
		//Requirements: 17.7
		fileWriter := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "app.log"),
			MaxSize:    100, // MB
			MaxBackups: 30,  // 保留30个备份
			MaxAge:     90,  // 保留90天
			Compress:   true,
			LocalTime:  true,
		}
		writers = append(writers, zapcore.AddSync(fileWriter))
	}

	// 输出到标准输出
	if logConfig.Output == "stdout" || logConfig.Output == "both" || logConfig.Output == "" {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 如果没有配置任何输出，默认输出到标准输出
	if len(writers) == 0 {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 创建多输出核心
	writer := zapcore.NewMultiWriteSyncer(writers...)

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		writer,
		level,
	)

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

// Error 记录错误日志
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
