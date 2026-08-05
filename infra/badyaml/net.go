package badyaml

import (
	"fmt"
	"net/netip"
	urlpkg "net/url"

	"github.com/duakc/lightddns/infra/netx/domains"
)

type URL struct {
	URL *urlpkg.URL
	Raw string
}

func (m *URL) UnmarshalYAML(data []byte) error {
	s, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	parse, err := urlpkg.Parse(s)
	if err != nil {
		return err
	}
	m.URL = parse
	m.Raw = s
	return nil
}

type DomainName string

func (d *DomainName) UnmarshalYAML(data []byte) error {
	s, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	if !domains.IsDomainName(s) {
		return fmt.Errorf("invalid domain name: %s", s)
	}
	*d = DomainName(s)
	return nil
}

type (
	Prefix   netip.Prefix
	AddrPort netip.AddrPort
	Addr     netip.Addr
)

func (a *Prefix) UnmarshalYAML(data []byte) error {
	prefixString, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	prefix, err := netip.ParsePrefix(prefixString)
	if err != nil {
		return err
	}
	*a = Prefix(prefix)
	return nil
}

func (a *Addr) UnmarshalYAML(data []byte) error {
	addrString, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	addr, err := netip.ParseAddr(addrString)
	if err != nil {
		return err
	}
	*a = Addr(addr)
	return nil
}

func (a *AddrPort) UnmarshalYAML(data []byte) error {
	addrPortString, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	addrPort, err := netip.ParseAddrPort(addrPortString)
	if err != nil {
		return err
	}
	*a = AddrPort(addrPort)
	return nil
}
