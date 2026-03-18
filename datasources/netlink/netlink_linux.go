//go:build linux

package netlink

import (
	"context"
	"net"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	CST "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"
)

func newNetLink(ctx context.Context, option options.OptionDataSourceNetlink) (adapter.DataSource, error) {
	panic("not implemented")
}

type netlink struct {
	name         string
	netInterface net.Interface
}

func (n *netlink) Type() string {
	return CST.DataSourceTypeNetlink
}

func (n *netlink) Name() string {
	return n.name
}

func (n *netlink) GetIPv4(ctx context.Context) ([]netip.Addr, error) {
}

func (n *netlink) GetIPv6(ctx context.Context) ([]netip.Addr, error) {

}
