package datasources

import (
	_ "github.com/duakc/lightddns/datasources/command"
	_ "github.com/duakc/lightddns/datasources/groups/failover"
	_ "github.com/duakc/lightddns/datasources/groups/sum"
	_ "github.com/duakc/lightddns/datasources/http"
	_ "github.com/duakc/lightddns/datasources/netlink"
)
