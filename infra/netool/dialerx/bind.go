package dialerx

import (
	"net"
	"net/netip"

	"github.com/duakc/lightddns/infra/netool/control"
	"github.com/duakc/lightddns/infra/netool/internal"
)

type bindDialer struct {
	useInterface    bool
	interfaceDialer net.Dialer

	v4TcpDialer net.Dialer
	v6TcpDialer net.Dialer
	v4UdpDialer net.Dialer
	v6UdpDialer net.Dialer
	v4L3Dialer  net.Dialer
	v6L3Dialer  net.Dialer
}

func newInterfaceBindDialer(this net.Dialer,
	networkInfName string, networkInfIndex int,
	finder control.InterfaceFinder,
) bindDialer {
	if finder == nil {
		finder = control.NewDefaultInterfaceFinder()
	}
	this.Control = control.Append(this.Control, control.BindToInterface(finder, networkInfName, networkInfIndex))
	return bindDialer{useInterface: true, interfaceDialer: this}
}

func newAddressesBindDialer(this net.Dialer,
	address4 netip.Addr, address6 netip.Addr,
) bindDialer {
	b := bindDialer{
		v4TcpDialer: this, v6TcpDialer: this,
		v4UdpDialer: this, v6UdpDialer: this,
		v4L3Dialer: this, v6L3Dialer: this,
	}

	if internal.IsIPv4(address4) {
		address4 = address4.Unmap()
		b.v4TcpDialer.LocalAddr = &net.TCPAddr{IP: address4.AsSlice(), Port: 0, Zone: ""}
		b.v4UdpDialer.LocalAddr = &net.UDPAddr{IP: address4.AsSlice(), Port: 0, Zone: ""}
		b.v4L3Dialer.LocalAddr = &net.IPAddr{IP: address4.AsSlice(), Zone: ""}
	}
	if internal.IsIPv6(address6) {
		b.v6TcpDialer.LocalAddr = &net.TCPAddr{IP: address6.AsSlice(), Port: 0, Zone: address6.Zone()}
		b.v6UdpDialer.LocalAddr = &net.UDPAddr{IP: address6.AsSlice(), Port: 0, Zone: address6.Zone()}
		b.v6L3Dialer.LocalAddr = &net.IPAddr{IP: address6.AsSlice(), Zone: address6.Zone()}
	}
	return b
}
