package options

import (
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netool/control"
	"github.com/duakc/lightddns/infra/netool/dialerx"
)

// ConnectOption @Shared
// @LANG.EN_US Connect Option
// @LANG.ZH_CN 连接选项
type ConnectOption struct {
	// @LANG.EN_US
	// Firewall Mark
	//
	// @LANG.ZH_CN
	// 防火墙标记
	//
	// @Platform linux
	Fwmark uint `json:"fwmark,omitempty" yaml:"fwmark,omitempty"`

	// @LANG.EN_US
	// DNS configuration
	//
	// @LANG.ZH_CN
	// DNS 配置
	DNS DNSOption `json:"dns,omitempty" yaml:"dns,omitempty"`

	// @LANG.EN_US
	// The IPv4 address to bind when establishing outgoing connections.
	//
	// @LANG.ZH_CN
	// 建立出站连接时绑定的 IPv4 地址。
	BindAddress4 string `json:"bindAddress4,omitempty" yaml:"bindAddress4,omitempty"`

	// @LANG.EN_US
	// The IPv6 address to bind when establishing outgoing connections.
	//
	// @LANG.ZH_CN
	// 建立出站连接时绑定的 IPv6 地址。
	BindAddress6 string `json:"bindAddress6,omitempty" yaml:"bindAddress6,omitempty"`

	// @LANG.EN_US
	// control underlay dialer's priority of establish a new connection, based on Happy Eyeball
	// "ipv4_only": Only attempt IPv4 connections. All AAAA records are ignored, no IPv6 fallback.
	// "prefer_ipv4": Prefer IPv4 and race with IPv6. If IPv4 connectivity fails, fall back to IPv6.
	// "ipv6_only": Only attempt IPv6 connections. All A records are ignored, no IPv4 fallback.
	// "prefer_ipv6": Prefer IPv6 and race with IPv4. If IPv6 connectivity fails, fall back to IPv4.
	//
	// @LANG.ZH_CN
	// 控制底层拨号器建立连接的优先级，基于 Happy Eyeball.
	// "ipv4_only": 仅尝试 IPv4 连接，忽略所有 AAAA 记录，不回退到 IPv6。
	// "ipv6_only": 仅尝试 IPv6 连接，忽略所有 A 记录，不回退到 IPv4。
	// "prefer_ipv4": 优先尝试 IPv4，同时发起 IPv6 竞速。若 IPv4 连接不可达，则回退到 IPv6。
	// "prefer_ipv6": 优先尝试 IPv6，同时发起 IPv4 竞速。若 IPv6 连接不可达，则回退到 IPv4。
	//
	// @Values ipv4_only, prefer_ipv4, ipv6_only, prefer_ipv6
	DialStrategy dialerx.DialStrategy `json:"dialStrategy,omitempty" yaml:"dialStrategy,omitempty"`

	// @LANG.EN_US
	// The network interface to bind when establishing connections, can be an interface name or index.
	//
	// @LANG.ZH_CN
	// 建立连接时绑定的网络接口，可以是接口名称或索引。
	BindInterface badyaml.StringOrNumber `json:"bindInterface,omitempty" yaml:"bindInterface,omitempty"`
}

func (co ConnectOption) Options() ([]dialerx.DialerOption, error) {
	var options []dialerx.DialerOption
	options = append(options, dialerx.WithDialStrategy(co.DialStrategy))
	if co.BindAddress4 != "" || co.BindAddress6 != "" {
		var (
			address4 netip.Addr
			address6 netip.Addr
			err      error
		)

		if co.BindAddress4 != "" {
			address4, err = netip.ParseAddr(co.BindAddress4)
			if err != nil {
				return nil, fmt.Errorf("ConnectOption.BindAddress4: %w", err)
			}
		}
		if co.BindAddress6 != "" {
			address6, err = netip.ParseAddr(co.BindAddress6)
			if err != nil {
				return nil, fmt.Errorf("ConnectOption.BindAddress6: %w", err)
			}
		}
		options = append(options, dialerx.WithBindAddress(address4, address6))
	}
	if co.BindInterface.Num != 0 {
		index := co.BindInterface.Num
		_, err := control.NewDefaultInterfaceFinder().ByIndex(int(index))
		if err != nil {
			return nil, fmt.Errorf("ConnectOption.BindInterface(index): %w", err)
		}
		options = append(options, dialerx.WithBindInterfaceIndex(int(index)))
	} else if co.BindInterface.Str != "" {
		name := co.BindInterface.Str
		_, err := control.NewDefaultInterfaceFinder().ByName(name)
		if err != nil {
			return nil, fmt.Errorf("ConnectOption.BindInterface(name): %w", err)
		}
		options = append(options, dialerx.WithBindInterfaceName(co.BindInterface.Str))
	}

	if co.Fwmark != 0 {
		options = append(options, dialerx.WithFwmark(uint32(co.Fwmark)))
	}

	return options, nil
}
