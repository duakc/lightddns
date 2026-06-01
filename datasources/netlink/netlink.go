package netlink

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/control"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

const DatasourceType = constpkg.DatasourceTypeNetlink

func init() {
	adapter.Register(
		adapter.DatasourceRegister,
		DatasourceType,
		New,
	)
}

func New(ctx context.Context, logger *zap.Logger, option options.NetlinkDatasourceOption) (adapter.Datasource, error) {
	n := &Netlink{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),
		interfaceFinder:     control.NewDefaultInterfaceFinder(),
		interfaceName:       option.IfName,
		interfaceIndex:      option.IfIndex,
		allowBogon:          option.AllowBogon,
		logger:              logger,
	}
	return n, nil
}

type Netlink struct {
	adapter.AbstractManagedType

	logger *zap.Logger

	interfaceFinder control.InterfaceFinder
	interfaceName   string
	interfaceIndex  int
	allowBogon      bool

	finderAccess sync.Mutex
}

func (n *Netlink) IP(ctx context.Context) ([]netip.Addr, error) {
	ip, err := n.ip(ctx)
	if err != nil {
		return nil, err
	}
	return mt.Filter(ip, func(addr netip.Addr) bool {
		return n.allowBogon || !netool.IsBogon(addr)
	}), nil
}

func (n *Netlink) ip(ctx context.Context) ([]netip.Addr, error) {
	logger := n.logger
	interfaceFinder := n.interfaceFinder
	n.finderAccess.Lock()
	if err := interfaceFinder.Update(); err != nil {
		n.finderAccess.Unlock()
		return nil, fmt.Errorf("update interfaceFinder: %w", err)
	}
	n.finderAccess.Unlock()
	if n.interfaceIndex != 0 {
		logger.Debug("use index", zap.Int("index", n.interfaceIndex))
		index, err := interfaceFinder.ByIndex(n.interfaceIndex)
		if err != nil {
			if n.interfaceName != "" {
				logger.Debug("use index failed, fallback to use name", zap.Error(err))
				goto useName
			}
			return nil, fmt.Errorf("by index: %w", err)
		}
		return mt.Map(index.Addresses, netip.Prefix.Addr), nil
	}
useName:
	if n.interfaceName != "" {
		logger.Debug("use interfaceName", zap.String("name", n.interfaceName))
		index, err := interfaceFinder.ByName(n.interfaceName)
		if err != nil {
			return nil, fmt.Errorf("by name: %w", err)
		}
		return mt.Map(index.Addresses, netip.Prefix.Addr), nil
	}
	return nil, fmt.Errorf("not configured for netlink datasource")
}
