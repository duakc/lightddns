package common

import (
	"fmt"
	"os"
)

const (
	colReset  = "\x1b[0m"
	colGreen  = "\x1b[32m"
	colYellow = "\x1b[33m"
	colRed    = "\x1b[31m"
)

var color = colorSupported()

func Infof(format string, a ...any) { logln(colGreen, "(goscript:info): "+format, a...) }

func Warnf(format string, a ...any) { logln(colYellow, "(goscript:warn): "+format, a...) }

func Fatalf(format string, a ...any) {
	logln(colRed, "(goscript:fatal): "+format, a...)
	os.Exit(1)
}

func logln(c, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	if color {
		msg = c + msg + colReset
	}
	_, _ = fmt.Fprintln(os.Stderr, msg)
}

func colorSupported() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stderr.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
