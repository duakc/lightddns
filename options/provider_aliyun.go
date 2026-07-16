package options

import constpkg "github.com/duakc/lightddns/constant"

type AliyunProviderOption struct {
	AbstractProviderOption `yaml:",inline"`

	Connect ConnectOption `json:"connect,omitempty" yaml:"connect,omitempty"`
	HTTP    HTTPOption    `json:"http,omitempty"    yaml:"http,omitempty"`
	DNS     DNSOption     `json:"dns,omitempty"     yaml:"dns,omitempty"`

	AccessKeyId     string `json:"accessKeyId"     yaml:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret" yaml:"accessKeySecret"`
}

func (AliyunProviderOption) UsedType() string {
	return constpkg.ProviderTypeAliyun
}
