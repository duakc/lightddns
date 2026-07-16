package options

import (
	constpkg "github.com/duakc/lightddns/constant"
)

type CloudflareProviderOption struct {
	AbstractProviderOption `yaml:",inline"`

	Connect ConnectOption `json:"connect,omitempty" yaml:"connect,omitempty"`
	HTTP    HTTPOption    `json:"http,omitempty"    yaml:"http,omitempty"`
	DNS     DNSOption     `json:"dns,omitempty"     yaml:"dns,omitempty"`

	Token string `json:"token" yaml:"token"`

	Proxy bool `json:"proxy,omitempty" yaml:"proxy,omitempty"`
}

func (CloudflareProviderOption) UsedType() string {
	return constpkg.ProviderTypeCloudflare
}
