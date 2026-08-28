package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type AliyunProviderOption struct {
	AbstractProviderOption `yaml:",inline"`

	Connect ConnectOption `json:"connect,omitempty" yaml:"connect,omitempty"`
	HTTP    HTTPOption    `json:"http,omitempty"    yaml:"http,omitempty"`
	DNS     DNSOption     `json:"dns,omitempty"     yaml:"dns,omitempty"`

	AccessKeyId     string                   `json:"accessKeyId"     yaml:"accessKeyId"`
	AccessKeySecret string                   `json:"accessKeySecret" yaml:"accessKeySecret"`
	Lines           badyaml.Listable[string] `json:"lines,omitempty" yaml:"lines,omitempty"`
}

func (AliyunProviderOption) UsedType() string {
	return constpkg.ProviderTypeAliyun
}
