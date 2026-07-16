package castoption

import (
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/infra/netool/control"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/options"
)

func BuildDialer(rawOption options.ConnectOption) (dialerx.Dialer, error) {
	dialerOptions, err := ConnectOptionToDialerOption(rawOption)
	if err != nil {
		return nil, fmt.Errorf("build dialer option: %w", err)
	}
	return dialerx.NewDialerWithOption(dialerOptions...), nil
}

func ConnectOptionToDialerOption(rawOption options.ConnectOption) ([]dialerx.DialerOption, error) {
	var dialerOptions []dialerx.DialerOption
	dialerOptions = append(dialerOptions, dialerx.WithDialStrategy(rawOption.DialStrategy))
	if rawOption.BindAddress4 != "" || rawOption.BindAddress6 != "" {
		var (
			address4 netip.Addr
			address6 netip.Addr
			err      error
		)

		if rawOption.BindAddress4 != "" {
			address4, err = netip.ParseAddr(rawOption.BindAddress4)
			if err != nil {
				return nil, fmt.Errorf("ConnectOption.BindAddress4: %w", err)
			}
		}
		if rawOption.BindAddress6 != "" {
			address6, err = netip.ParseAddr(rawOption.BindAddress6)
			if err != nil {
				return nil, fmt.Errorf("ConnectOption.BindAddress6: %w", err)
			}
		}
		dialerOptions = append(dialerOptions, dialerx.WithBindAddress(address4, address6))
	}
	if rawOption.BindInterface.Num != 0 {
		index := rawOption.BindInterface.Num
		_, err := control.NewDefaultInterfaceFinder().ByIndex(int(index))
		if err != nil {
			return nil, fmt.Errorf("ConnectOption.BindInterface(index): %w", err)
		}
		dialerOptions = append(dialerOptions, dialerx.WithBindInterfaceIndex(int(index)))
	} else if rawOption.BindInterface.Str != "" {
		name := rawOption.BindInterface.Str
		_, err := control.NewDefaultInterfaceFinder().ByName(name)
		if err != nil {
			return nil, fmt.Errorf("ConnectOption.BindInterface(name): %w", err)
		}
		dialerOptions = append(dialerOptions, dialerx.WithBindInterfaceName(name))
	}

	if rawOption.Fwmark != 0 {
		dialerOptions = append(dialerOptions, dialerx.WithFwmark(uint32(rawOption.Fwmark)))
	}

	return dialerOptions, nil
}
