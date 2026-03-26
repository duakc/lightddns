package zaplog

import (
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	EnvEnableStackTrace = "ZAP_STACK_TRACE"
)

type LoggerKey struct{}

var (
	NOP = zap.NewNop()
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
	core := zapcore.NewCore(coreEncoder, zapcore.Lock(zapcore.AddSync(output)), level)
	if stackTraceLevel := os.Getenv(EnvEnableStackTrace); stackTraceLevel != "" {
		parseLevel, err := zapcore.ParseLevel(stackTraceLevel)
		if err != nil {
			panic("wrong stack trace level:" + stackTraceLevel)
		}
		options = append(options, zap.AddStacktrace(parseLevel))
	}
	return zap.New(core, options...)
}

func DoNotPanic(l *zap.Logger) *zap.Logger {
	if l == nil {
		return NOP
	}
	return l
}

func ExtendName(l *zap.Logger, name string) *zap.Logger {
	l = DoNotPanic(l)
	if strings.Contains(name, ".") {
		return l.Named("[" + name + "]")
	}
	return l.Named(name)
}
