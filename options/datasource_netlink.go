package options

import CST "github.com/duakc/lightddns/constant"

type OptionDataSourceNetlink struct {
	AbstractProviderOption `yaml:",inline"`

	Interface string `yaml:"interface"`
	Index     int    `yaml:"index"`
}

func (o *OptionDataSourceNetlink) Type() string {
	return CST.DataSourceTypeNetlink
}
