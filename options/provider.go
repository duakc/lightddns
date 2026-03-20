package options

import "encoding/json"

type OptionProvider struct {
	AbstractProviderOption `yaml:",inline"`

	Raw json.RawMessage `yaml:"-"`
}

type AbstractProviderOption struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}
