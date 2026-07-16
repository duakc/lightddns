package options

type HTTPOption struct {
	UseSystemProxy bool   `json:"useSystemProxy,omitempty" yaml:"useSystemProxy,omitempty"`
	HTTPProxy      string `json:"httpProxy,omitempty"      yaml:"httpProxy,omitempty"`
	HTTPSProxy     string `json:"httpsProxy,omitempty"     yaml:"httpsProxy,omitempty"`
}
