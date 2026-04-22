package httpds

import (
	"context"
	"net/http"
	"testing"

	"github.com/duakc/lightddns/infra/httpxx"
	"github.com/duakc/lightddns/infra/netool"

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
	testCases := []Case{
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://ipinfo.io", nil, httpxx.NewClient(), ".ip", "")),
		},
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://api.ip.sb/ip", nil, httpxx.NewClient(), "", "")),
		},
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://api.ip.sb/ip", nil, httpxx.NewClient(
					httpxx.ClientOptionWithDialer(netool.NewDialerWithOption(
						netool.DialerOptionWithDialStrategy(netool.DialOnlyIPv4)))),
				"", "")),
			ipVersion: "4",
		},
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://myip.ipip.net", nil, httpxx.NewClient(), "", `当前 IP：\s*(.+?)\s*来自于：`)),
		},
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://api64.ipify.org", nil, httpxx.NewClient(), "", "")),
		},
		{
			ctx: context.Background(),
			reqContext: mt.Must(newRequestContext(http.MethodGet,
				"https://api64.ipify.org?format=json", nil, httpxx.NewClient(), ".ip", "")),
		},
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
				(tc.ipVersion == "6" && netool.IsIPv6(addr) || (tc.ipVersion == "4" && netool.IsIPv4(addr))),
				"error[%d]: excepted ipversion=%s , got=%s", testIndex, tc.ipVersion, addr.String())
		}
	}
}
