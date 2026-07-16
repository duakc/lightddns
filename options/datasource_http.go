package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type HTTPDatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`

	Connect ConnectOption `json:"connect,omitempty" yaml:"connect,omitempty"`
	HTTP    HTTPOption    `json:"http,omitempty"    yaml:"http,omitempty"`
	DNS     DNSOption     `json:"dns,omitempty"     yaml:"dns,omitempty"`

	URL badyaml.DualStack[badyaml.URL] `json:"url" yaml:"url"`

	JSON  badyaml.DualStack[string] `json:"json,omitempty"  yaml:"json,omitempty"`
	Regex badyaml.DualStack[string] `json:"regex,omitempty" yaml:"regex,omitempty"`

	Method  badyaml.HTTPMethod `json:"method,omitempty"  yaml:"method,omitempty"`
	Headers badyaml.HTTPHeader `json:"headers,omitempty" yaml:"headers,omitempty"`
}

func (HTTPDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeHTTP
}
