package options

import (
	"github.com/duakc/lightddns/adapter"
	goyaml "github.com/goccy/go-yaml"
)

type OptionProvider struct {
	AbstractProviderOption `yaml:",inline"`

	Option any `yaml:"-"`
}

type _OptionProvider OptionProvider

func (O *OptionProvider) UnmarshalYAML(bs []byte) error {
	err := goyaml.Unmarshal(bs, (*_OptionProvider)(O))
	if err != nil {
		return err
	}
	O.Option, err = adapter.ProviderRegister.CreateOption(O.Type)
	if err != nil {
		return err
	}
	return goyaml.Unmarshal(bs, O.Option)
}

type AbstractProviderOption struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}
