package options

type OptionProviderCloudflare struct {
	AbstractProviderOption `yaml:",inline"`

	Token string `yaml:"token"`
	Proxy string `yaml:"Proxy"`
}
