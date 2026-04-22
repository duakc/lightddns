package control

import (
	"fmt"
	"syscall"
)

func BindToInterface(finder InterfaceFinder, interfaceName string, interfaceIndex int) Func {
	return func(network, address string, conn syscall.RawConn) error {
		return BindToInterface0(finder, conn, network, address, interfaceName, interfaceIndex, false)
	}
}

func BindToInterfaceFunc(finder InterfaceFinder, block func(network string, address string) (interfaceName string, interfaceIndex int, err error)) Func {
	return func(network, address string, conn syscall.RawConn) error {
		interfaceName, interfaceIndex, err := block(network, address)
		if err != nil {
			return err
		}
		return BindToInterface0(finder, conn, network, address, interfaceName, interfaceIndex, false)
	}
}

func BindToInterface0(finder InterfaceFinder, conn syscall.RawConn, network string, address string, interfaceName string, interfaceIndex int, preferInterfaceName bool) error {
	if interfaceName == "" && interfaceIndex == -1 {
		return fmt.Errorf("interface not found: %s", interfaceName)
	}
	return bindToInterface(conn, network, address, finder, interfaceName, interfaceIndex, preferInterfaceName)
}
