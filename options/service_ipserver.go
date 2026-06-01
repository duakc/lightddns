package options

import constpkg "github.com/duakc/lightddns/constant"

type IPServerServiceOption struct {
	AbstractServiceOption `yaml:",inline"`

	Enabled bool   `json:"enabled"                   yaml:"enabled"`
	Listen  string `json:"listen,omitempty"          yaml:"listen,omitempty"`
	Port    uint16 `json:"port,omitempty"            yaml:"port,omitempty"`
	Path    string `json:"path,omitempty"            yaml:"path,omitempty"`
	Dump    bool   `json:"dump,omitempty"            yaml:"dump,omitempty"`
}

func (S *IPServerServiceOption) UsedType() string {
	return constpkg.ServiceTypeIPServer
}
