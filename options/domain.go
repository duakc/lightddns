package options

import (
	"github.com/duakc/lightddns/infra/badyaml"
)

type DomainOption struct {
	Domain badyaml.DomainName `json:"domain" yaml:"domain"`

	Enabled    bool   `json:"enabled"    yaml:"enabled"`
	Provider   string `json:"provider"   yaml:"provider"`
	DataSource string `json:"datasource" yaml:"datasource"`

	TTL  uint32 `json:"ttl,omitempty"  yaml:"ttl,omitempty"`
	IPv4 bool   `json:"ipv4,omitempty" yaml:"ipv4,omitempty"`
	IPv6 bool   `json:"ipv6,omitempty" yaml:"ipv6,omitempty"`

	Interval badyaml.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`
	Timeout  badyaml.Duration `json:"timeout,omitempty"  yaml:"timeout,omitempty"`
}
