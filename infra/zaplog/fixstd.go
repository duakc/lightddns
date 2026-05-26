package zaplog

import (
	"errors"
	"io/fs"
	"os"
	"syscall"

	"github.com/duakc/mt/gosys"
)

// fix sync error : https://github.com/uber-go/zap/issues/991

var (
	Stdout = &fixConsoleWs{os.Stdout}
	Stderr = &fixConsoleWs{os.Stderr}
)

type fixConsoleWs struct {
	*os.File
}

func (ws *fixConsoleWs) Sync() error {
	err := ws.File.Sync()
	if err != nil {
		if gosys.IsWindows {
			if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
				// ERROR_INVALID_HANDLE(6) or ERROR_INVALID_FUNCTION(1):
				// handle doesn't support sync (e.g., stdout redirected to pipe)
				if errno, ok := pathErr.Err.(syscall.Errno); ok && (errno == 6 || errno == 1) {
					err = nil
				}
			}
		} else {
			if errors.Is(err, syscall.EBADF) || errors.Is(err, syscall.ENOTTY) {
				err = nil
			}
		}
	}
	return err
}
