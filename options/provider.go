package options

import "encoding/json"

type OptionProvider struct {
	AbstractProviderOption `yaml:",inline"`

	Option any             `yaml:"-"`
	Raw    json.RawMessage `yaml:"-"`
}

type AbstractProviderOption struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}
