package netlink

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/common"
	"github.com/duakc/lightddns/infra/netinterface"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"go.uber.org/zap"
)

func New(logger *zap.Logger, option options.OptionDataSourceNetlink) (adapter.DataSource, error) {
	return &netlink{
		logger:         zaplog.ExtendName(logger, option.Name),
		name:           option.Name,
		interfaceName:  option.Name,
		interfaceIndex: option.Index,
	}, nil
}

type netlink struct {
	logger *zap.Logger

	name string

	interfaceName  string
	interfaceIndex int
}

func (n *netlink) Type() string {
	return constpkg.DataSourceTypeNetlink
}

func (n *netlink) Name() string {
	return n.name
}

func (n *netlink) IP(ctx context.Context) ([]netip.Addr, error) {
	logger := n.logger
	if n.interfaceIndex != 0 {
		logger.Debug("use index", zap.Int("index", n.interfaceIndex))
		index, err := netinterface.FindInterfaceByIndex(n.interfaceIndex)
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
		index, err := netinterface.FindInterfaceByName(n.interfaceName)
		if err != nil {
			return nil, fmt.Errorf("by name: %w", err)
		}
		return common.Map(index.Addresses, netip.Prefix.Addr), nil
	}
	return nil, fmt.Errorf("not configured for netlink datasource")
}
