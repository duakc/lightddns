package control

import (
	"errors"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func DisableUDPFragment() Func {
	return func(network, address string, conn syscall.RawConn) error {
		if !strings.HasPrefix(network, "udp") {
			return nil
		}
		return Raw(conn, func(fd uintptr) error {
			if network == "udp" || network == "udp4" {
				err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_DONTFRAG, 1)
				if err != nil {
					if errors.Is(err, unix.ENOPROTOOPT) || errors.Is(err, unix.EOPNOTSUPP) {
						return nil
					}
					return os.NewSyscallError("SETSOCKOPT IP_DONTFRAG", err)
				}
			}
			if network == "udp" || network == "udp6" {
				err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_DONTFRAG, 1)
				if err != nil {
					if errors.Is(err, unix.ENOPROTOOPT) || errors.Is(err, unix.EOPNOTSUPP) {
						return nil
					}
					return os.NewSyscallError("SETSOCKOPT IPV6_DONTFRAG", err)
				}
			}
			return nil
		})
	}
}
