package options

import (
	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"

	goyaml "github.com/goccy/go-yaml"
)

type ProviderOption struct {
	AbstractProviderOption

	Option any `json:"-" yaml:"-"`
}

type _ProviderOption ProviderOption

func (O *ProviderOption) UnmarshalYAML(bs []byte) error {
	err := badyaml.Unmarshal(bs, (*_ProviderOption)(O))
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
	Type string `json:"type" yaml:"type"`
	Name string `json:"name" yaml:"name"`
}

func (AbstractProviderOption) UsedType() string {
	return "abstract_provider"
}
