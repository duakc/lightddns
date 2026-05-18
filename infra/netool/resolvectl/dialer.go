package resolvectl

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strconv"

	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/internal"
	"github.com/duakc/lightddns/infra/netool/resolvectl/transports"

	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

type ResolveDialer struct {
	logger *zap.Logger

	dialer    dialerx.Dialer
	transport transports.Transport
}

func (r *ResolveDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	if len(address) == 0 {
		return nil, dialerx.ErrNoAddressToDialer
	}

	var (
		host     string
		hostPort uint16
		err      error
	)
	if slices.Contains(internal.L4NetworkList, network) {
		var (
			port       string
			hostPort64 uint64
		)
		host, port, err = net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		hostPort64, err = strconv.ParseUint(port, 10, 16)
		if err != nil {
			return nil, err
		}
		hostPort = uint16(hostPort64)
	}
	if !internal.IsDomainName(host) {
		return r.dialer.DialContext(ctx, network, address)
	}

	strategy := ResolveAsis
	if network == "udp4" || network == "tcp4" || network == "ip4" {
		strategy = ResolveIPv4
	} else if network == "udp6" || network == "tcp6" || network == "ip6" {
		strategy = ResolveIPv6
	}
	resolver := services.LookupDefault[ResolveClient](ctx, DefaultResolveClient)
	var addresses []netip.Addr
	addresses, err = resolver.Lookup(ctx, r.transport, host, strategy)
	if err != nil {
		return nil, fmt.Errorf("lookup: %w", err)
	}
	if len(addresses) == 0 {
		return nil, dialerx.ErrNoAddressToDialer
	}
	if parallelDialer, isParallel := r.dialer.(dialerx.ParallelDialer); isParallel {
		return parallelDialer.DialParallel(ctx, addresses, hostPort)
	}
	return dialerx.DialParallel(ctx, r.dialer, addresses, hostPort,
		internal.IsIPv6(addresses[0]) || strategy == ResolveIPv6,
		internal.DefaultHappyEyeballFallbackDelay)
}

func NewDialer(ctx context.Context, dialer dialerx.Dialer, transport transports.Transport) *ResolveDialer {
	logger := services.LookupPtrDefault[zap.Logger](ctx, resolverLogger)
	return &ResolveDialer{logger: logger, dialer: dialer, transport: transport}
}
