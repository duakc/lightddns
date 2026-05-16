package options

import constpkg "github.com/duakc/lightddns/constant"

type DatasourceGroupSumOption struct {
	AbstractDatasourceGroupOption

	FastFail bool `json:"fastFail,omitempty" yaml:"fastFail,omitempty"`
}

func (DatasourceGroupSumOption) UsedType() string {
	return constpkg.DatasourceGroupTypeSum
}
