package castoption

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/duakc/lightddns/infra/netx/httpx"
	"github.com/duakc/lightddns/infra/netx/resolvectl"
	"github.com/duakc/lightddns/infra/netx/resolvectl/transports"
	"github.com/duakc/lightddns/options"

	"github.com/stretchr/testify/require"
)

func TestHTTPOptionToHTTPXOptionsUsesSystemProxyWithoutExplicitProxy(t *testing.T) {
	transport := newHTTPTransportFromOptions(t, options.HTTPOption{
		UseSystemProxy: true,
	})

	require.NotNil(t, transport.Proxy)
}

func TestHTTPOptionToHTTPXOptionsPrefersExplicitProxyOverSystemProxy(t *testing.T) {
	transport := newHTTPTransportFromOptions(t, options.HTTPOption{
		UseSystemProxy: true,
		HTTPProxy:      "http://explicit-proxy.test:8080",
		HTTPSProxy:     "http://explicit-proxy.test:8080",
	})

	req, err := http.NewRequest(http.MethodGet, "https://example.test/ip", nil)
	require.NoError(t, err)

	proxyURL, err := transport.Proxy(req)
	require.NoError(t, err)
	require.Equal(t, "http://explicit-proxy.test:8080", proxyURL.String())
}

func TestBuildHTTPClientFromScratchSkipsResolveDialerWhenDNSDisabled(t *testing.T) {
	rawDialer, resolveDialer, requester, err := BuildHTTPClientFromScratch(
		nil,
		options.ConnectOption{},
		options.DNSOption{},
		options.HTTPOption{
			HTTPProxy: "http://proxy.test:8080",
		},
	)
	require.NoError(t, err)
	require.Same(t, rawDialer, resolveDialer)

	client := requester.(*httpx.Client)
	transport := client.Transport.(*http.Transport)
	require.NotNil(t, transport.Proxy)
}

func TestBuildHTTPClientFromScratchUsesResolveDialerWhenDNSEnabled(t *testing.T) {
	rawDialer, resolveDialer, requester, err := BuildHTTPClientFromScratch(
		nil,
		options.ConnectOption{},
		options.DNSOption{
			Enabled: true,
			Type:    transports.TransportTypeSystem,
		},
		options.HTTPOption{
			HTTPProxy: "http://proxy.test:8080",
		},
	)
	require.NoError(t, err)
	require.NotSame(t, rawDialer, resolveDialer)
	require.IsType(t, &resolvectl.ResolveDialer{}, resolveDialer)

	client := requester.(*httpx.Client)
	transport := client.Transport.(*http.Transport)
	require.NotNil(t, transport.Proxy)
}

func newHTTPTransportFromOptions(t *testing.T, option options.HTTPOption) *http.Transport {
	t.Helper()

	httpOptions, err := HTTPOptionToHTTPXOptions(option)
	require.NoError(t, err)

	client := httpx.NewClient(noopDialer{}, httpOptions...)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	return transport
}

type noopDialer struct{}

func (noopDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	return nil, errors.New("noop dialer should not be used")
}
