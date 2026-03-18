package options

type OptionDomain struct {
	Domain     string `yaml:"domain"`
	TTL        int    `yaml:"TTL"`
	IPv4       bool   `yaml:"ipv4"`
	IPv6       bool   `yaml:"ipv6"`
	Provider   string `yaml:"provider"`
	DataSource string `yaml:"data-source"`
}
