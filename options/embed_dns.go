package options

import (
	"fmt"
	urlpkg "net/url"
	"strconv"
	"strings"

	"github.com/duakc/lightddns/infra/netx/resolvectl/transports"

	goyaml "github.com/goccy/go-yaml"
)

type DNSOption struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Type    string `json:"type,omitempty"    yaml:"type,omitempty"`
	Server  string `json:"server,omitempty"  yaml:"server,omitempty"`
	Port    uint16 `json:"port,omitempty"    yaml:"port,omitempty"`
}

type _DNSOption DNSOption

func (do *DNSOption) UnmarshalYAML(data []byte) error {
	// String form: `dns: system` or `dns: tls://8.8.8.8:853`.
	var raw string
	if err := goyaml.Unmarshal(data, &raw); err == nil {
		return do.unmarshalString(raw)
	}
	// Object form: `dns: {type: tls, server: 8.8.8.8, port: 853}`.
	return goyaml.Unmarshal(data, (*_DNSOption)(do))
}

// unmarshalString parses the shorthand string form. The string form is always
// enabled; the default port is applied later, not here.
func (do *DNSOption) unmarshalString(s string) error {
	do.Enabled = true
	if lower := strings.ToLower(strings.TrimSpace(s)); lower == "" || lower == transports.TransportTypeSystem {
		do.Type = transports.TransportTypeSystem
		return nil
	}
	dnsURL, err := urlpkg.Parse(s)
	if err != nil {
		return fmt.Errorf("resolve DNSOption: %w", err)
	}
	switch dnsURL.Scheme {
	case transports.TransportTypeTLS:
		do.Type = transports.TransportTypeTLS
		do.Server = dnsURL.Hostname()
		if dnsURL.Port() != "" {
			numPort, err := strconv.ParseUint(dnsURL.Port(), 10, 16)
			if err != nil {
				return fmt.Errorf("bad port: %w", err)
			}
			do.Port = uint16(numPort)
		}
	default:
		return fmt.Errorf("unknown dns: `%s`", s)
	}
	return nil
}