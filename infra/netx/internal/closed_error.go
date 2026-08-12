package internal

import (
	"errors"
	"io"
	"net"
	"os"
)

// IsClosedError reports whether err represents an already closed connection.
func IsClosedError(err error) bool {
	return isClosedError(err)
}

func isClosedError(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, os.ErrClosed)
}
