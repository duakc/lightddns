package zaplog

import (
	"time"

	"github.com/duakc/mt/debug"

	"go.uber.org/multierr"
	"go.uber.org/zap/zapcore"
)

var _ zapcore.Core = (*SmartCore)(nil)

type SmartCore struct {
	zapcore.Core

	ioCore      zapcore.Core
	flushBuffer *zapcore.BufferedWriteSyncer

	outputWriteIsSync bool
}

func NewSmartCore(enc zapcore.Encoder, w zapcore.WriteSyncer,
	lvl zapcore.LevelEnabler,
) *SmartCore {
	smart := &SmartCore{}
	if bufferWriteSyncer, isBuffer := w.(*zapcore.BufferedWriteSyncer); isBuffer {
		smart.flushBuffer = bufferWriteSyncer
	}

	if smart.flushBuffer == nil {
		smart.flushBuffer = &zapcore.BufferedWriteSyncer{
			WS:            w,
			Size:          16 * 1024, // 16k
			FlushInterval: 5 * time.Second,
			Clock:         zapcore.DefaultClock,
		}
		if debug.Enabled {
			// debug need a faster flush speed
			smart.flushBuffer.FlushInterval = 1 * time.Second
		}
	}

	smart.ioCore = zapcore.NewCore(enc, smart.flushBuffer, lvl)
	smart.Core = smart.ioCore
	return smart
}

func (c *SmartCore) Sync() error {
	return c.flushBuffer.Sync()
}

func (c *SmartCore) Close() error {
	return c.flushBuffer.Stop()
}

var _ zapcore.Core = (*splitLevelCore)(nil)

type splitLevelCore struct {
	zapcore.Core

	flushLevel zapcore.Level
}

func newSplitLevelCore(c zapcore.Core, flushLevel zapcore.Level) zapcore.Core {
	return &splitLevelCore{
		Core:       c,
		flushLevel: flushLevel,
	}
}

func (c *splitLevelCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	err := c.Core.Write(entry, fields)

	if err != nil || c.flushLevel.Enabled(entry.Level) || entry.Level > zapcore.ErrorLevel {
		err = multierr.Append(err, c.Core.Sync())
	}
	return err
}
