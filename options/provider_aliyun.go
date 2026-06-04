package options

import constpkg "github.com/duakc/lightddns/constant"

type AliyunProviderOption struct {
	AbstractProviderOption `yaml:",inline"`
	ConnectOption          `yaml:",inline"`
	HTTPOption             `yaml:",inline"`

	AccessKeyId     string `json:"accessKeyId"     yaml:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret" yaml:"accessKeySecret"`
}

func (AliyunProviderOption) UsedType() string {
	return constpkg.ProviderTypeAliyun
}
