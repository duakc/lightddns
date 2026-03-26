package options

import (
	"github.com/duakc/lightddns/infra/badyaml"
)

type OptionDomain struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
	TTL      int    `yaml:"ttl"`
	IPv4     bool   `yaml:"ipv4"`
	IPv6     bool   `yaml:"ipv6"`
	Provider string `yaml:"provider"`

	Domain     badyaml.Listable[string] `yaml:"domain"`
	DataSource badyaml.Listable[string] `yaml:"datasource"`
}
