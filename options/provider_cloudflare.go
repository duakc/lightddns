package options

import (
	constpkg "github.com/duakc/lightddns/constant"
)

type CloudflareProviderOption struct {
	AbstractProviderOption `yaml:",inline"`
	ConnectOption          `yaml:",inline"`
	HTTPOption             `yaml:",inline"`

	Token string `json:"token" yaml:"token"`

	Proxy bool `json:"proxy,omitempty" yaml:"proxy,omitempty"`
}

func (CloudflareProviderOption) UsedType() string {
	return constpkg.ProviderTypeCloudflare
}
