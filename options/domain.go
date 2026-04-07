package options

import (
	"github.com/duakc/lightddns/infra/badyaml"
)

type DomainOption struct {
	Enabled  bool   `yaml:"enabled"`
	TTL      uint32 `yaml:"ttl"`
	IPv4     bool   `yaml:"ipv4"`
	IPv6     bool   `yaml:"ipv6"`
	Provider string `yaml:"provider"`

	Domain     badyaml.DomainName       `yaml:"domain"`
	Interval   badyaml.Duration         `yaml:"interval"`
	DataSource badyaml.Listable[string] `yaml:"datasource"`
}
