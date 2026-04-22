package dialerx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/duakc/lightddns/infra/netool/internal"
	"github.com/duakc/mt"
)

const defaultFallbackDelay = 300 * time.Millisecond

func DialParallel(ctx context.Context, dialer Dialer,
	network string, addresses []netip.Addr, port uint16,
	preferIPv6 bool, fallbackDelay time.Duration,
) (net.Conn, error) {
	if fallbackDelay == 0 {
		fallbackDelay = defaultFallbackDelay
	}

	returned := make(chan struct{})
	defer close(returned)

	addresses4 := mt.Filter(addresses, internal.IsIPv4)
	addresses6 := mt.Filter(addresses, internal.IsIPv6)

	if len(addresses4) == 0 || len(addresses6) == 0 {
		return DialSerial(ctx, dialer, network, addresses, port)
	}

	var primaries, fallbacks []netip.Addr
	if preferIPv6 {
		primaries = addresses6
		fallbacks = addresses4
	} else {
		primaries = addresses4
		fallbacks = addresses6
	}
	type dialResult struct {
		net.Conn
		error

		primary bool
		done    bool
	}
	results := make(chan dialResult)
	startRacer := func(ctx context.Context, primary bool) {
		ras := primaries
		if !primary {
			ras = fallbacks
		}
		c, err := DialSerial(ctx, dialer, network, ras, port)
		select {
		case results <- dialResult{Conn: c, error: err, primary: primary, done: true}:
		case <-returned:
			if c != nil {
				c.Close()
			}
		}
	}
	var primary, fallback dialResult
	primaryCtx, primaryCancel := context.WithCancel(ctx)
	defer primaryCancel()
	go startRacer(primaryCtx, true)
	fallbackTimer := time.NewTimer(fallbackDelay)
	defer fallbackTimer.Stop()
	for {
		select {
		case <-fallbackTimer.C:
			fallbackCtx, fallbackCancel := context.WithCancel(ctx)
			defer fallbackCancel()
			go func() {
				startRacer(fallbackCtx, false)
			}()

		case res := <-results:
			if res.error == nil {
				return res.Conn, nil
			}
			if res.primary {
				primary = res
			} else {
				fallback = res
			}
			if primary.done && fallback.done {
				return nil, primary.error
			}
			if res.primary && fallbackTimer.Stop() {
				fallbackTimer.Reset(0)
			}
		}
	}
}

func DialSerial(ctx context.Context, this Dialer, network string, address []netip.Addr, port uint16) (net.Conn, error) {
	var errs error
	if len(address) == 0 {
		return nil, fmt.Errorf("no address to dial")
	}
	for _, addr := range address {
		var conn net.Conn
		var err error
		addrPort := netip.AddrPortFrom(addr.Unmap(), port)
		if addrPortDialer, isAddrPortDialer := this.(AddrPortDialer); isAddrPortDialer {
			conn, err = addrPortDialer.DialContextAddrPort(ctx, network, addrPort)
		} else {
			conn, err = this.DialContext(ctx, network, addrPort.String())
		}
		if err == nil {
			return conn, nil
		}
		errs = errors.Join(errs, err)
	}

	return nil, fmt.Errorf("all addresses dial failed : %w", errs)
}
