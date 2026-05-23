package zaplog

import (
	"io"
	"os"

	"github.com/duakc/mt/debug"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	EnvEnableStackTrace = "ZAP_STACK_TRACE"
)

var NOP = zap.NewNop()

func NewDefault(output io.Writer, level zapcore.LevelEnabler, options []zap.Option) *zap.Logger {
	coreEncoder := zapcore.NewJSONEncoder(DefaultJSONEncoderConfig())
	smartCore := zapcore.Core(NewSmartCore(coreEncoder, zapcore.Lock(zapcore.AddSync(output)), level))
	if level.Enabled(zapcore.WarnLevel) {
		smartCore = newSplitLevelCore(smartCore, zapcore.WarnLevel)
	}

	if stackTraceLevel := os.Getenv(EnvEnableStackTrace); stackTraceLevel != "" {
		parseLevel, err := zapcore.ParseLevel(stackTraceLevel)
		if err != nil {
			panic("wrong stack trace level:" + stackTraceLevel)
		}
		options = append(options, zap.AddStacktrace(parseLevel))
	}
	if debug.Enabled {
		options = append(options, zap.Development())
	}
	return zap.New(smartCore, options...)
}

func DoNotPanic(logger *zap.Logger) *zap.Logger {
	if logger == nil {
		return NOP
	}
	return logger
}

func DefaultJSONEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
