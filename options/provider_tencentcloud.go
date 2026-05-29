package options

import constpkg "github.com/duakc/lightddns/constant"

type TencentCloudProviderOption struct {
	AbstractProviderOption `yaml:",inline"`
	ConnectOption          `yaml:",inline"`
	HTTPOption             `yaml:",inline"`

	SecretId  string `json:"secretId"  yaml:"secretId"`
	SecretKey string `json:"secretKey" yaml:"secretKey"`
}

func (TencentCloudProviderOption) UsedType() string {
	return constpkg.ProviderTypeTencentCloud
}
