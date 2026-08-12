package httpx

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/duakc/lightddns/infra/netx/resolvectl"
	"github.com/duakc/lightddns/infra/netx/resolvectl/transports"

	mDns "github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestDebugClientForwardsRequest(t *testing.T) {
	var called bool
	client := &DebugClient{
		Logger: zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)),
		HTTPRequester: requesterFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			require.Equal(t, http.MethodGet, req.Method)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("203.0.113.30")),
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.test/ip", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProxyRequestOnlyResolvesProxyHost(t *testing.T) {
	proxyDialer := newPipeProxyDialer(t)

	resolver := newStaticResolveClient(map[string]netip.Addr{
		"proxy.test": netip.MustParseAddr("127.0.0.1"),
	})
	dnsTransport := transports.FuncTransport(func(context.Context, *mDns.Msg) (*mDns.Msg, error) {
		return nil, errors.New("static resolver should not exchange")
	})
	resolveDialer := resolvectl.NewDialer(proxyDialer, dnsTransport, resolver)
	proxyURL := "http://proxy.test:8080"

	client := NewClient(resolveDialer, ClientOptionWithProxy(proxyURL, proxyURL, ""))
	req, err := http.NewRequest(http.MethodGet, "http://target.test/ip", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)

	require.Equal(t, "GET http://target.test/ip HTTP/1.1", <-proxyDialer.requestLine)
	require.Equal(t, 1, resolver.lookupCount("proxy.test"))
	require.Equal(t, 0, resolver.lookupCount("target.test"))
}

type requesterFunc func(*http.Request) (*http.Response, error)

func (f requesterFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newPipeProxyDialer(t *testing.T) *pipeProxyDialer {
	t.Helper()
	return &pipeProxyDialer{
		t:           t,
		requestLine: make(chan string, 1),
	}
}

type pipeProxyDialer struct {
	t           *testing.T
	requestLine chan string
}

func (d *pipeProxyDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return nil, errors.New("unexpected domain dial " + host)
	}
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, err
	}
	return d.DialAddrPort(ctx, network, addr, uint16(portNum))
}

func (d *pipeProxyDialer) DialAddrPort(ctx context.Context, _ string, addr netip.Addr, port uint16) (net.Conn, error) {
	require.Equal(d.t, netip.MustParseAddr("127.0.0.1"), addr)
	require.Equal(d.t, uint16(8080), port)

	clientConn, serverConn := net.Pipe()
	go func() {
		defer serverConn.Close()

		reader := bufio.NewReader(serverConn)
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return
		}
		for {
			headerLine, headerErr := reader.ReadString('\n')
			if headerErr != nil || headerLine == "\r\n" {
				break
			}
		}
		d.requestLine <- strings.TrimRight(line, "\r\n")

		_, _ = serverConn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\nok"))
	}()
	return clientConn, nil
}

type staticResolveClient struct {
	mu        sync.Mutex
	addresses map[string]netip.Addr
	lookups   map[string]int
}

func newStaticResolveClient(addresses map[string]netip.Addr) *staticResolveClient {
	return &staticResolveClient{
		addresses: addresses,
		lookups:   make(map[string]int),
	}
}

func (r *staticResolveClient) Lookup(
	_ context.Context,
	_ transports.Transport,
	domain string,
	_ resolvectl.LookupStrategy,
) ([]netip.Addr, error) {
	domain = strings.TrimSuffix(domain, ".")

	r.mu.Lock()
	defer r.mu.Unlock()

	r.lookups[domain]++
	addr, ok := r.addresses[domain]
	if !ok {
		return nil, errors.New("unexpected lookup " + domain)
	}
	return []netip.Addr{addr}, nil
}

func (r *staticResolveClient) Exchange(
	context.Context,
	transports.Transport,
	*mDns.Msg,
) (*mDns.Msg, error) {
	return nil, errors.New("static resolver does not exchange")
}

func (r *staticResolveClient) lookupCount(domain string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lookups[domain]
}
