package options

import constpkg "github.com/duakc/lightddns/constant"

type DatasourceGroupSumOption struct {
	AbstractDatasourceGroupOption

	IgnoreDownstreamError bool `json:"ignoreDownstreamError,omitempty" yaml:"ignoreDownstreamError,omitempty"`
}

func (DatasourceGroupSumOption) UsedType() string {
	return constpkg.DatasourceGroupTypeSum
}
