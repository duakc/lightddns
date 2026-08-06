package tencentcloud

import (
	"fmt"
	"net/http"
	"time"

	"github.com/duakc/lightddns/infra/netx/httpx"

	"go.uber.org/zap"
)

// TencentSignHTTPRequester signs outgoing requests with TC3-HMAC-SHA256 before
// forwarding them to the wrapped HTTPRequester.
//
// The caller is expected to have already set X-TC-Action and X-TC-Version
// on the request (typically via Client.newRequest). X-TC-Timestamp, Host,
// and Authorization are injected here.
//
// The request body is buffered fully because TC3 signs the SHA256 of the
// body. This is fine for DNSPod payloads (tiny JSON) but is the wrong
// shape for streaming APIs.
type TencentSignHTTPRequester struct {
	httpx.HTTPRequester

	Logger *zap.Logger

	SecretId  string
	SecretKey string
	Service   string
}

func (tc *TencentSignHTTPRequester) Do(r *http.Request) (resp *http.Response, err error) {
	defer httpx.NewHTTPRequestRecorder(tc.Logger, r, &resp, &err).Record()

	common := Common{
		Timestamp: time.Now().UTC().Unix(),
	}

	httpx.ExtendHeadersOverride(r.Header, common.Headers())

	bodyBytes, err := httpx.ReadAndReplayBody(r)
	if err != nil {
		return nil, fmt.Errorf("tencentcloud sign: %w", err)
	}

	sig := SigContext{
		Method:    r.Method,
		Headers:   r.Header,
		Body:      bodyBytes,
		Timestamp: common.Timestamp,
		SecretId:  tc.SecretId,
		SecretKey: tc.SecretKey,
		Service:   tc.Service,
	}
	auth, err := sig.Authorization()
	if err != nil {
		return nil, fmt.Errorf("tencentcloud sign: authorization: %w", err)
	}
	r.Header.Set(httpx.HeaderAuthorization, auth)

	return tc.HTTPRequester.Do(r)
}
