package options

import (
	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"

	goyaml "github.com/goccy/go-yaml"
)

type ProviderOption struct {
	AbstractProviderOption `yaml:",inline"`

	Option any `json:"-" yaml:"-"`
}

type _ProviderOption ProviderOption

func (po *ProviderOption) UnmarshalYAML(bs []byte) error {
	err := goyaml.Unmarshal(bs, (*_ProviderOption)(po))
	if err != nil {
		return err
	}
	po.Option, err = adapter.ProviderRegister.CreateOption(po.Type)
	if err != nil {
		return err
	}
	return badyaml.Unmarshal(bs, po.Option)
}

func (po *ProviderOption) setName(name string) {
	po.AbstractProviderOption.setName(name)
	if setter, canSetName := po.Option.(nameSetter); canSetName {
		setter.setName(name)
	}
}

var (
	_ nameSetter    = (*AbstractProviderOption)(nil)
	_ VariantOption = (*AbstractProviderOption)(nil)
)

type AbstractProviderOption struct {
	Type string `json:"type"           yaml:"type"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

func (O *AbstractProviderOption) MajorType() string {
	return "provider"
}

func (O *AbstractProviderOption) UsedType() string {
	return "abstract_provider"
}

func (O *AbstractProviderOption) setName(name string) { O.Name = name }
func (O *AbstractProviderOption) getName() string     { return O.Name }
