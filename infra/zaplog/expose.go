package zaplog

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger = NewDefault(os.Stderr, zapcore.DebugLevel, []zap.Option{
	zap.AddCallerSkip(1),
	zap.Fields(zap.Bool("logger_default", true)),
}).Named("default")

func Debug(msg string, fields ...zap.Field) {
	defer defaultLogger.Sync()
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	defer defaultLogger.Sync()
	defaultLogger.Info(msg, fields...)
}
func Warn(msg string, fields ...zap.Field) {
	defer defaultLogger.Sync()
	defaultLogger.Warn(msg, fields...)
}
func Error(msg string, fields ...zap.Field) {
	defer defaultLogger.Sync()
	defaultLogger.Error(msg, fields...)
}

func DPanic(msg string, fields ...zap.Field) {
	defer defaultLogger.Sync()
	defaultLogger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	defer defaultLogger.Sync()
	defaultLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	defer defaultLogger.Sync()
	defaultLogger.Fatal(msg, fields...)
}
