package options

import (
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netool/control"
	"github.com/duakc/lightddns/infra/netool/dialerx"
)

// ConnectOption @Shared
type ConnectOption struct {
	FwManirk uint `yaml:"fwmark"`

	DNS          DNSOption `yaml:"dns"`
	BindAddress4 string    `yaml:"bind-address4"`
	BindAddress6 string    `yaml:"bind-address6"`

	DialStrategy  dialerx.DialStrategy   `yaml:"dial-strategy"`
	BindInterface badyaml.StringOrNumber `yaml:"bind-interface"`
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

	if co.FwManirk != 0 {
		options = append(options, dialerx.WithFwmark(uint32(co.FwManirk)))
	}

	return options, nil
}
