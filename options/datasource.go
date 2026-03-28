package options

import (
	"github.com/duakc/lightddns/adapter"
	goyaml "github.com/goccy/go-yaml"
)

type OptionDataSource struct {
	AbstractDatasourceOption `yaml:",inline"`

	Option any `yaml:"-"`
}

type _OptionDataSource OptionDataSource

func (O *OptionDataSource) UnmarshalYAML(bs []byte) error {
	err := goyaml.Unmarshal(bs, (*_OptionDataSource)(O))
	if err != nil {
		return err
	}
	O.Option, err = adapter.DataSourceRegister.CreateOption(O.Type)
	if err != nil {
		return err
	}
	return goyaml.Unmarshal(bs, O.Option)
}

type AbstractDatasourceOption struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}
