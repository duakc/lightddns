package options

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	urlpkg "net/url"
	"strconv"
	"strings"

	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/resolvectl/transports"

	"github.com/duakc/mt"
)

type DNSOption struct {
	Type   string `json:"type"           yaml:"type"`
	Server string `json:"server"         yaml:"server"`
	Port   uint16 `json:"port,omitempty" yaml:"port,omitempty"`
}

func (n *DNSOption) UnmarshalYAML(data []byte) error {
	unquoted := mt.UnquoteString(string(data))
	lowerUnquoted := strings.ToLower(unquoted)
	if lowerUnquoted == "system" || lowerUnquoted == "" {
		n.Type = transports.TransportTypeSystem
		return nil
	}
	if !strings.Contains(unquoted, "://") {
		unquoted = "udp://" + unquoted
	}
	dnsURL, err := urlpkg.Parse(unquoted)
	if err != nil {
		return fmt.Errorf("resolve DNSOption: %w", err)
	}
	switch dnsURL.Scheme {
	case transports.TransportTypeTLS:
		n.Type = transports.TransportTypeTLS
		n.Server = dnsURL.Host
		if dnsURL.Port() == "" {
			n.Port = 853
		} else if numPort, err := strconv.ParseUint(dnsURL.Port(), 10, 16); err != nil {
			return fmt.Errorf("bad port: %w", err)
		} else {
			n.Port = uint16(numPort)
		}
	default:
		return fmt.Errorf("unknown dns: `%s`", unquoted)
	}
	return nil
}

func (n *DNSOption) NewTransport(ctx context.Context, dialer dialerx.Dialer) (transports.Transport, error) {
	switch n.Type {
	case transports.TransportTypeSystem:
		return &transports.SystemTransport{}, nil
	case transports.TransportTypeTLS:
		return transports.NewTLS(ctx, dialer,
			net.JoinHostPort(n.Server, strconv.FormatUint(uint64(n.Port), 10)), &tls.Config{})
	default:
		return nil, fmt.Errorf("unknown dns type: %s", n.Type)
	}
}
