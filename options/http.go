package options

import (
	"github.com/duakc/lightddns/infra/netool/httpx"
)

type HTTPOption struct {
	UseSystemProxy bool   `json:"useSystemProxy,omitempty" yaml:"useSystemProxy,omitempty"`
	HTTPProxy      string `json:"httpProxy,omitempty"      yaml:"httpProxy,omitempty"`
	HTTPSProxy     string `json:"httpsProxy,omitempty"     yaml:"httpsProxy,omitempty"`
}

func (ho HTTPOption) Options() ([]httpx.HTTPClientOption, error) {
	var options []httpx.HTTPClientOption

	if ho.HTTPProxy != "" || ho.HTTPSProxy != "" {
		if ho.HTTPProxy == "" {
			ho.HTTPProxy = ho.HTTPSProxy
		}
		if ho.HTTPSProxy == "" {
			ho.HTTPSProxy = ho.HTTPProxy
		}
		if ho.HTTPProxy == "" && ho.HTTPSProxy == "" && ho.UseSystemProxy {
			options = append(options, httpx.ClientOptionEnableProxy())
		} else {
			options = append(options, httpx.ClientOptionWithProxy(ho.HTTPProxy, ho.HTTPSProxy, ""))
		}
	}
	return options, nil
}
