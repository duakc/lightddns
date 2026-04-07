package netlink

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	datasourcepkg "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/common"
	"github.com/duakc/lightddns/infra/ctxservice"
	"github.com/duakc/lightddns/infra/netxx/control"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"go.uber.org/zap"
)

func init() {
	adapter.Register(
		adapter.DataSourceRegister,
		constpkg.DatasourceTypeNetlink,
		New,
	)
}

func New(ctx context.Context, option options.NetlinkDatasourceOption) (adapter.DataSource, error) {

	return &Netlink{
		logger: datasourcepkg.NewLogger(
			ctxservice.Lookup[*zap.Logger](ctx, common.Zero[zaplog.LoggerKey]()),
			option.AbstractDatasourceOption,
		),
		interfaceFinder: control.NewDefaultInterfaceFinder(),
		name:            option.Name,
		interfaceName:   option.Interface,
		interfaceIndex:  option.Index,
		allowPrivate:    option.AllowPrivate,
	}, nil
}

type Netlink struct {
	logger *zap.Logger
	name   string

	interfaceFinder control.InterfaceFinder
	interfaceName   string
	interfaceIndex  int
	allowPrivate    bool
}

func (n *Netlink) Type() string {
	return constpkg.DatasourceTypeNetlink
}

func (n *Netlink) Name() string {
	return n.name
}

func (n *Netlink) IP(ctx context.Context) ([]netip.Addr, error) {
	ip, err := n.ip(ctx)
	if err != nil {
		return nil, err
	}
	return common.Filter(ip, func(addr netip.Addr) bool {
		return n.allowPrivate || addr.IsGlobalUnicast()
	}), nil
}

func (n *Netlink) ip(ctx context.Context) ([]netip.Addr, error) {
	logger := n.logger
	defer logger.Sync()
	interfaceFinder := n.interfaceFinder
	if err := interfaceFinder.Update(); err != nil {
		return nil, fmt.Errorf("update interfaceFinder: %w", err)
	}
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
		return common.Map(index.Addresses, netip.Prefix.Addr), nil
	}
useName:
	if n.interfaceName != "" {
		logger.Debug("use interfaceName", zap.String("name", n.interfaceName))
		index, err := interfaceFinder.ByName(n.interfaceName)
		if err != nil {
			return nil, fmt.Errorf("by name: %w", err)
		}
		return common.Map(index.Addresses, netip.Prefix.Addr), nil
	}
	return nil, fmt.Errorf("not configured for netlink datasource")
}
