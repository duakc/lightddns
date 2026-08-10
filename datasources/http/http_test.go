package http

//func TestRequestContext_Handle(t *testing.T) {
//	type Case struct {
//		ctx        context.Context
//		reqContext *requestContext
//		ipVersion  string
//	}
//	defaultDialer := dialerx.NewDialerWithOption()
//
//	testCases := []Case{
//		{
//			ctx: context.Background(),
//			reqContext: mt.Must(newRequestContext(http.MethodGet,
//				"https://ipinfo.io", nil, httpx.NewClient(defaultDialer), ".ip", "")),
//		},
//		{
//			ctx: context.Background(),
//			reqContext: mt.Must(newRequestContext(http.MethodGet,
//				"https://api.ip.sb/ip", nil, httpx.NewClient(defaultDialer), "", "")),
//		},
//		{
//			ctx: context.Background(),
//			reqContext: mt.Must(newRequestContext(http.MethodGet,
//				"https://api.ip.sb/ip", nil, httpx.NewClient(
//					dialerx.NewDialerWithOption(
//						dialerx.WithDialStrategy(dialerx.DialOnlyIPv4))),
//				"", "")),
//			ipVersion: "4",
//		},
//		{
//			ctx: context.Background(),
//			reqContext: mt.Must(newRequestContext(http.MethodGet,
//				"https://myip.ipip.net", nil, httpx.NewClient(defaultDialer), "", `当前 IP：\s*(.+?)\s*来自于：`)),
//		},
//
//		// unstable.
//		//{
//		//	ctx: context.Background(),
//		//	reqContext: mt.Must(newRequestContext(http.MethodGet,
//		//		"https://api64.ipify.org", nil, httpx.NewClient(defaultDialer), "", "")),
//		//},
//		//{
//		//	ctx: context.Background(),
//		//	reqContext: mt.Must(newRequestContext(http.MethodGet,
//		//		"https://api64.ipify.org?format=json", nil, httpx.NewClient(defaultDialer), ".ip", "")),
//		//},
//	}
//	for testIndex, tc := range testCases {
//		addresses, err := tc.reqContext.Handle(tc.ctx)
//		if err != nil {
//			require.NoErrorf(t, err, "error[%d]: %s",
//				testIndex, err.Error())
//		}
//		assert.Truef(t, len(addresses) > 0, "error[%d]: no address found", testIndex)
//		for _, addr := range addresses {
//			assert.Truef(t, addr.IsValid(), "error[%d]: invalid address", testIndex)
//			assert.Truef(t, tc.ipVersion == "" ||
//				(tc.ipVersion == "6" && netx.IsIPv6(addr) || (tc.ipVersion == "4" && netx.IsIPv4(addr))),
//				"error[%d]: excepted ipversion=%s , got=%s", testIndex, tc.ipVersion, addr.String())
//		}
//	}
//}
//
//func newRequestContext(method string, url string, headers http.Header,
//	requester httpx.HTTPRequester, jq string, re string,
//) (*requestContext, error) {
//	R := new(requestContext)
//	var err error
//	if jq != "" {
//		if R.jsonMatch, err = gojq.Parse(jq); err != nil {
//			return nil, fmt.Errorf("MatchJson: %w", err)
//		}
//	}
//	if re != "" {
//		if R.regexMatch, err = regexp.Compile(re); err != nil {
//			return nil, fmt.Errorf("MatchRegex: %w", err)
//		}
//	}
//	parsedURL, err := urlpkg.Parse(url)
//	if err != nil {
//		return nil, fmt.Errorf("parse url: %w", err)
//	}
//	R.method = method
//	R.url = parsedURL
//	R.headers = headers
//	R.requester = requester
//
//	return R, nil
//}
