package zaplog

import (
	"strings"

	"github.com/duakc/mt/debug"
	"github.com/duakc/mt/services/closeme"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultLevel  = zap.NewAtomicLevelAt(zap.DebugLevel)
	defaultLogger = createDefault(defaultLevel)
)

func createDefault(lvl zapcore.LevelEnabler) *zap.Logger {
	if debug.IsTestEnv() || !debug.Enabled {
		// omit log message on test
		defaultLevel.SetLevel(zap.ErrorLevel)
	}

	var options []zap.Option
	if debug.Enabled {
		options = append(options, zap.Development())
	}
	options = append(options, zap.AddCallerSkip(1))

	return NewDefault(closeme.Default, Stderr, lvl, options).
		Named("default")
}

func DefaultLevel(lvl zapcore.Level) {
	defaultLevel.SetLevel(lvl)
}

func NewPackage(pkg ...string) *zap.Logger {
	return defaultLogger.With(zap.String("package", strings.Join(pkg, "/")))
}

func Sugar() *zap.SugaredLogger {
	return defaultLogger.Sugar()
}

func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	_ = defaultLogger.Sync()

	defaultLogger.Error(msg, fields...)
}

func DPanic(msg string, fields ...zap.Field) {
	_ = defaultLogger.Sync()

	defaultLogger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	_ = defaultLogger.Sync()

	defaultLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	_ = defaultLogger.Sync()

	defaultLogger.Fatal(msg, fields...)
}
