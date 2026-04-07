package options

type CloudflareProviderOption struct {
	AbstractProviderOption `yaml:",inline"`
	ConnectOption          `yaml:",inline"`
	HTTPOption             `yaml:",inline"`

	Token string `yaml:"token"`
	Proxy bool   `yaml:"proxy"`
}
