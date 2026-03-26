package options

type OptionProviderCloudflare struct {
	abstractProviderOption `yaml:",inline"`

	Token string `yaml:"token"`
	Proxy bool   `yaml:"Proxy"`
}
