package options

import constpkg "github.com/duakc/lightddns/constant"

type DatasourceGroupFailoverOption struct {
	AbstractDatasourceGroupOption `yaml:",inline"`
}

func (DatasourceGroupFailoverOption) UsedType() string {
	return constpkg.DatasourceGroupTypeFailover
}
