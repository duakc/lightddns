package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type DatasourceGroupFilterRuleOption struct {
	Prefixes []badyaml.Prefix `json:"prefixes,omitempty" yaml:"prefixes,omitempty"`

	// use "0.0.0.0/0" in Prefixes to include all IPv4 addresses
	// IncludeIPv4 bool `json:"includeIPv4,omitempty" yaml:"includeIPv4,omitempty"`

	// use "::/0" in Prefixes to include all IPv6 addresses
	// IncludeIPv6 bool `json:"includeIPv6,omitempty" yaml:"includeIPv6,omitempty"`

	Invert bool `json:"invert,omitempty" yaml:"invert,omitempty"`
}

type DatasourceGroupFilterOption struct {
	AbstractDatasourceGroupOption `yaml:",inline"`

	Rules []DatasourceGroupFilterRuleOption `yaml:"rules" json:"rules"`
}

func (DatasourceGroupFilterOption) UsedType() string {
	return constpkg.DatasourceGroupFilter
}
