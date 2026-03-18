package options

import CST "github.com/duakc/lightddns/constant"

type OptionDataSourceNetlink struct {
	Interface string `yaml:"interface"`
}

func (o *OptionDataSourceNetlink) Type() string {
	return CST.DataSourceTypeNetlink
}
