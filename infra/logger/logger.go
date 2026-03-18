package logger

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewDefault(output io.Writer, level zapcore.Level, options []zap.Option) *zap.Logger {
	coreEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
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
	})
	core := zapcore.NewCore(coreEncoder, zapcore.AddSync(output), level)
	if stackTraceLevel := os.Getenv("ZAP_STACK_TRACE"); stackTraceLevel != "" {
		parseLevel, err := zapcore.ParseLevel(stackTraceLevel)
		if err != nil {
			panic("wrong stack trace level:" + stackTraceLevel)
		}
		options = append(options, zap.AddStacktrace(parseLevel))
	}
	return zap.New(core, options...)
}
