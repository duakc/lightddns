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

	URL     badyaml.URL        `json:"url" yaml:"url"`
	Method  badyaml.HTTPMethod `json:"method,omitempty"  yaml:"method,omitempty"`
	Headers badyaml.HTTPHeader `json:"headers,omitempty" yaml:"headers,omitempty"`

	JQ    badyaml.JQ    `json:"jq,omitempty" yaml:"jq,omitempty"`
	Regex badyaml.Regex `json:"regex,omitempty" yaml:"regex,omitempty"`
}

func (HTTPDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeHTTP
}
