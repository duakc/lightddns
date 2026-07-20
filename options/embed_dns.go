package options

import (
	"fmt"
	urlpkg "net/url"
	"strconv"
	"strings"

	"github.com/duakc/lightddns/infra/netx/resolvectl/transports"

	"github.com/duakc/mt"
)

type DNSOption struct {
	Type   string `json:"type"           yaml:"type"`
	Server string `json:"server"         yaml:"server"`
	Port   uint16 `json:"port,omitempty" yaml:"port,omitempty"`
}

func (do *DNSOption) UnmarshalYAML(data []byte) error {
	unquoted := mt.UnquoteString(string(data))
	lowerUnquoted := strings.ToLower(unquoted)
	if lowerUnquoted == "system" || lowerUnquoted == "" {
		do.Type = transports.TransportTypeSystem
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
		do.Type = transports.TransportTypeTLS
		do.Server = dnsURL.Host
		if dnsURL.Port() == "" {
			do.Port = 853
		} else if numPort, err := strconv.ParseUint(dnsURL.Port(), 10, 16); err != nil {
			return fmt.Errorf("bad port: %w", err)
		} else {
			do.Port = uint16(numPort)
		}
	default:
		return fmt.Errorf("unknown dns: `%s`", unquoted)
	}
	return nil
}
