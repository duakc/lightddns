package lightddns

import (
	// base datasource
	_ "github.com/duakc/lightddns/datasources/command"
	// group datasource
	_ "github.com/duakc/lightddns/datasources/groups/failover"
	_ "github.com/duakc/lightddns/datasources/groups/sum"
	_ "github.com/duakc/lightddns/datasources/http"
	_ "github.com/duakc/lightddns/datasources/netlink"
	// provider
	_ "github.com/duakc/lightddns/providers/cloudflare"
)
