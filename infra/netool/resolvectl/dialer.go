package resolvectl

import "github.com/duakc/lightddns/infra/netool/dialerx"

type ResolveDialer struct {
	resolve ResolveClient
	dialer  dialerx.Dialer

	strategy ResolveStrategy
}
