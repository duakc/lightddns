package options

import constpkg "github.com/duakc/lightddns/constant"

type DatasourceGroupFailoverOption struct {
	AbstractDatasourceGroupOption
}

func (DatasourceGroupFailoverOption) UsedType() string {
	return constpkg.DatasourceGroupTypeFailover
}
