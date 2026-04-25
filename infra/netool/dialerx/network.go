package dialerx

import (
	"context"
	"net"
)

type NetworkDialer struct {
	Network string
	Dialer  Dialer
}

func (n *NetworkDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	return n.Dialer.DialContext(ctx, n.Network, address)
}
