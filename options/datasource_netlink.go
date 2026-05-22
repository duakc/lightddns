package options

import constpkg "github.com/duakc/lightddns/constant"

type NetlinkDatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`

	IfName       string `json:"ifName,omitempty"       yaml:"ifName,omitempty"`
	IfIndex      int    `json:"ifIndex,omitempty"      yaml:"ifIndex,omitempty"`
	AllowPrivate bool   `json:"allowPrivate,omitempty" yaml:"allowPrivate,omitempty"`
}

func (NetlinkDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeNetlink
}
