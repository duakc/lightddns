package options

import constpkg "github.com/duakc/lightddns/constant"

type DatasourceGroupSumOption struct {
	AbstractDatasourceGroupOption `yaml:",inline"`
}

func (DatasourceGroupSumOption) UsedType() string {
	return constpkg.DatasourceGroupTypeSum
}
