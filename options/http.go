package options

import (
	"github.com/duakc/lightddns/infra/httpxx"
)

type HTTPOption struct {
	UseSystemProxy bool   `yaml:"use-system-proxy"`
	HTTPProxy      string `yaml:"http-proxy"`
	HTTPSProxy     string `yaml:"https-proxy"`
}

func (ho HTTPOption) Options() ([]httpxx.HTTPClientOption, error) {
	var options []httpxx.HTTPClientOption

	if ho.HTTPProxy != "" || ho.HTTPSProxy != "" {
		if ho.HTTPProxy == "" {
			ho.HTTPProxy = ho.HTTPSProxy
		}
		if ho.HTTPSProxy == "" {
			ho.HTTPSProxy = ho.HTTPProxy
		}
		if ho.HTTPProxy == "" && ho.HTTPSProxy == "" && ho.UseSystemProxy {
			options = append(options, httpxx.ClientOptionEnableProxy())
		} else {
			options = append(options, httpxx.ClientOptionWithProxy(ho.HTTPProxy, ho.HTTPSProxy, ""))
		}
	}
	return options, nil
}
