package logger

import (
	"log/slog"
	"os"
)

// Logger 全局日志实例
var Logger *slog.Logger

// logFile 日志文件句柄（用于关闭）
var logFile *os.File

// Init 初始化日志
func Init(level, format, output, filePath string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if output == "file" && filePath != "" {
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			slog.Error("Failed to open log file", "error", err)
			os.Exit(1)
		}
		logFile = file

		if format == "json" {
			handler = slog.NewJSONHandler(file, opts)
		} else {
			handler = slog.NewTextHandler(file, opts)
		}
	} else {
		if format == "json" {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// Debug 调试日志
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

// Info 信息日志
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// Warn 警告日志
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

// Error 错误日志
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

// Close 关闭日志文件
func Close() {
	if logFile != nil {
		logFile.Sync()
		logFile.Close()
	}
}
