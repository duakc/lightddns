package options

import constpkg "github.com/duakc/lightddns/constant"

type TencentCloudProviderOption struct {
	AbstractProviderOption `yaml:",inline"`

	Connect ConnectOption `json:"connect,omitempty" yaml:"connect,omitempty"`
	HTTP    HTTPOption    `json:"http,omitempty"    yaml:"http,omitempty"`
	DNS     DNSOption     `json:"dns,omitempty"     yaml:"dns,omitempty"`

	SecretId  string `json:"secretId"  yaml:"secretId"`
	SecretKey string `json:"secretKey" yaml:"secretKey"`
}

func (TencentCloudProviderOption) UsedType() string {
	return constpkg.ProviderTypeTencentCloud
}
