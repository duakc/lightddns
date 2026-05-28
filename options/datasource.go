package options

import (
	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"

	goyaml "github.com/goccy/go-yaml"
)

type DatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`

	Option any `json:"-" yaml:"-"`
}

type _DatasourceOption DatasourceOption

func (do *DatasourceOption) UnmarshalYAML(bs []byte) error {
	err := goyaml.Unmarshal(bs, (*_DatasourceOption)(do))
	if err != nil {
		return err
	}
	do.Option, err = adapter.DatasourceRegister.CreateOption(do.Type)
	if err != nil {
		return err
	}
	return badyaml.Unmarshal(bs, do.Option)
}

type AbstractDatasourceOption struct {
	Type string `json:"type"           yaml:"type"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

func (AbstractDatasourceOption) UsedType() string {
	return "abstract_datasource"
}
