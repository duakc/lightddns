package options

import constpkg "github.com/duakc/lightddns/constant"

type NetlinkDatasourceOption struct {
	AbstractDatasourceOption

	Name         string `json:"name,omitempty"         yaml:"name,omitempty"`
	Index        int    `json:"index,omitempty"        yaml:"index,omitempty"`
	AllowPrivate bool   `json:"allowPrivate,omitempty" yaml:"allowPrivate,omitempty"`
}

func (NetlinkDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeNetlink
}
