package lightddns

import (
	// group datasource
	_ "github.com/duakc/lightddns/datasources/groups/failover"
	_ "github.com/duakc/lightddns/datasources/groups/sum"

	// base datasource
	_ "github.com/duakc/lightddns/datasources/httpds"
	_ "github.com/duakc/lightddns/datasources/netlink"

	// provider
	_ "github.com/duakc/lightddns/providers/cloudflare"
)
