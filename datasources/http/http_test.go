package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/duakc/lightddns/infra/netx"
	"github.com/duakc/lightddns/infra/netx/dialerx"
	"github.com/duakc/lightddns/infra/netx/httpx"

	"github.com/duakc/mt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestContext_Handle(t *testing.T) {
	type Case struct {
		ctx        context.Context
		reqContext *requestContext
		ipVersion  string
	}
	defaultDialer := dialerx.NewDialerWithOption()

	testCases := []Case{
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://ipinfo.io", nil, httpx.NewClient(defaultDialer), ".ip", "")),
		},
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://api.ip.sb/ip", nil, httpx.NewClient(defaultDialer), "", "")),
		},
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://api.ip.sb/ip", nil, httpx.NewClient(
					dialerx.NewDialerWithOption(
						dialerx.WithDialStrategy(dialerx.DialOnlyIPv4))),
				"", "")),
			ipVersion: "4",
		},
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://myip.ipip.net", nil, httpx.NewClient(defaultDialer), "", `当前 IP：\s*(.+?)\s*来自于：`)),
		},

		// unstable.
		//{
		//	ctx: context.Background(),
		//	reqContext: mt.Must(newRequestContext(http.MethodGet,
		//		"https://api64.ipify.org", nil, httpx.NewClient(defaultDialer), "", "")),
		//},
		//{
		//	ctx: context.Background(),
		//	reqContext: mt.Must(newRequestContext(http.MethodGet,
		//		"https://api64.ipify.org?format=json", nil, httpx.NewClient(defaultDialer), ".ip", "")),
		//},
	}
	for testIndex, tc := range testCases {
		addresses, err := tc.reqContext.Handle(tc.ctx)
		if err != nil {
			require.NoErrorf(t, err, "error[%d]: %s",
				testIndex, err.Error())
		}
		assert.Truef(t, len(addresses) > 0, "error[%d]: no address found", testIndex)
		for _, addr := range addresses {
			assert.Truef(t, addr.IsValid(), "error[%d]: invalid address", testIndex)
			assert.Truef(t, tc.ipVersion == "" ||
				(tc.ipVersion == "6" && netx.IsIPv6(addr) || (tc.ipVersion == "4" && netx.IsIPv4(addr))),
				"error[%d]: excepted ipversion=%s , got=%s", testIndex, tc.ipVersion, addr.String())
		}
	}
}
