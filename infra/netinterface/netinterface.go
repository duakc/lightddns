package netinterface

import (
	"net"
	"net/netip"

	"github.com/duakc/lightddns/infra/common"
)

type Interface struct {
	*net.Interface
	Addresses []netip.Prefix
}

func FindInterfaceByName(name string) (*Interface, error) {
	return newInterfaceWithAddresses(net.InterfaceByName(name))
}

func FindInterfaceByIndex(idx int) (*Interface, error) {
	return newInterfaceWithAddresses(net.InterfaceByIndex(idx))
}

func newInterfaceWithAddresses(inf *net.Interface, err error) (*Interface, error) {
	if err != nil {
		return nil, err
	}
	addresses, err := inf.Addrs()
	if err != nil {
		return nil, err
	}
	return &Interface{
		Interface: inf,
		Addresses: common.Map(addresses, prefixFromNetAddr),
	}, nil
}

func prefixFromNetAddr(addr net.Addr) netip.Prefix {
	switch actAddr := addr.(type) {
	case *net.IPNet:
		bits, _ := actAddr.Mask.Size()
		ip, _ := netip.AddrFromSlice(actAddr.IP)
		return netip.PrefixFrom(ip.Unmap(), bits)
	default:
		return netip.Prefix{}
	}
}
