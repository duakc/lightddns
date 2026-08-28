package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type TencentCloudProviderOption struct {
	AbstractProviderOption `yaml:",inline"`

	Connect ConnectOption `json:"connect,omitempty" yaml:"connect,omitempty"`
	HTTP    HTTPOption    `json:"http,omitempty"    yaml:"http,omitempty"`
	DNS     DNSOption     `json:"dns,omitempty"     yaml:"dns,omitempty"`

	SecretId  string                   `json:"secretId"        yaml:"secretId"`
	SecretKey string                   `json:"secretKey"       yaml:"secretKey"`
	Lines     badyaml.Listable[string] `json:"lines,omitempty" yaml:"lines,omitempty"`
}

func (TencentCloudProviderOption) UsedType() string {
	return constpkg.ProviderTypeTencentCloud
}
