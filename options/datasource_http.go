package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type DualStack[T any] struct {
	IPv4 T `json:"ipv4" yaml:"ipv4"`
	IPv6 T `json:"ipv6" yaml:"ipv6"`
}

type HTTPDatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`
	ConnectOption            `yaml:",inline"`
	HTTPOption               `yaml:",inline"`

	URL badyaml.URL `json:"url" yaml:"url"`

	JSON  badyaml.StringOrObject[DualStack[string]] `json:"json,omitempty"  yaml:"json,omitempty"`
	Regex badyaml.StringOrObject[DualStack[string]] `json:"regex,omitempty" yaml:"regex,omitempty"`

	Method  badyaml.HTTPMethod `json:"method,omitempty"  yaml:"method,omitempty"`
	Headers badyaml.HTTPHeader `json:"headers,omitempty" yaml:"headers,omitempty"`
}

func (HTTPDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeHTTP
}
