package options

import (
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/control"
	"github.com/duakc/lightddns/infra/netool/transport"
)

type ConnectOption struct {
	DialStrategy string `yaml:"dial-strategy"`
	BindAddress4 string `yaml:"bind-address4"`
	BindAddress6 string `yaml:"bind-address6"`

	BindInterface badyaml.StringOrNumber `yaml:"bind-interface"`
	FwManirk      uint                   `yaml:"fwmark"`
	DNS           string                 `yaml:"dns"`
}

func (co ConnectOption) Options() ([]netool.DialerOption, error) {
	var options []netool.DialerOption

	if dialStrategy, ok := netool.DialStrategyFromString(co.DialStrategy); ok {
		options = append(options, netool.DialerOptionWithDialStrategy(dialStrategy))
	}
	if co.BindAddress4 != "" {
		inet4Address, err := netip.ParseAddr(co.BindAddress4)
		if err != nil {
			return nil, fmt.Errorf("inet4Address: %w", err)
		}
		options = append(options, netool.DialerOptionBindAddress4(inet4Address))
	}
	if co.BindAddress6 != "" {
		inet6Address, err := netip.ParseAddr(co.BindAddress6)
		if err != nil {
			return nil, fmt.Errorf("inet6Address: %w", err)
		}
		options = append(options, netool.DialerOptionBindAddress6(inet6Address))
	}
	if co.BindInterface.Num != 0 {
		index := co.BindInterface.Num
		_, err := control.NewDefaultInterfaceFinder().ByIndex(int(index))
		if err != nil {
			return nil, fmt.Errorf("interface index: %d not found: %w", index, err)
		}
		options = append(options, netool.DialerOptionWithBindInterfaceIndex(int(index)))
	} else if co.BindInterface.Str != "" {
		name := co.BindInterface.Str
		_, err := control.NewDefaultInterfaceFinder().ByName(name)
		if err != nil {
			return nil, fmt.Errorf("interface name: %s not found: %w", name, err)
		}
		options = append(options, netool.DialerOptionWithBindInterfaceName(name))
	}

	if co.FwManirk != 0 {
		options = append(options, netool.DialerOptionFwmark(uint32(co.FwManirk)))
	}
	if co.DNS != "" {
		transport, err := transport.NewDNSTransport(co.DNS)
		if err != nil {
			return nil, fmt.Errorf("build dns transport: %w", err)
		}
		options = append(options, netool.DialerOptionWithTransport(transport))
	}

	return options, nil
}
