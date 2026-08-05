package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type DatasourceGroupFilterOption struct {
	AbstractDatasourceGroupOption `yaml:",inline"`

	Prefixes badyaml.Listable[badyaml.Prefix] `json:"prefixes"         yaml:"prefixes"`
	Invert   bool                             `json:"invert,omitempty" yaml:"invert,omitempty"`
}

func (DatasourceGroupFilterOption) UsedType() string {
	return constpkg.DatasourceGroupFilter
}
