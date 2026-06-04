package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt/freebuf"
)

// TencentSignClient signs outgoing requests with TC3-HMAC-SHA256 before
// forwarding them to the wrapped HTTPRequester.
//
// The caller is expected to have already set X-TC-Action and X-TC-Version
// on the request (typically via Client.newRequest). X-TC-Timestamp, Host,
// and Authorization are injected here.
//
// The request body is buffered fully because TC3 signs the SHA256 of the
// body. This is fine for DNSPod payloads (tiny JSON) but is the wrong
// shape for streaming APIs.
type TencentSignClient struct {
	httpx.HTTPRequester

	SecretId  string
	SecretKey string
	Service   string
}

func (tc *TencentSignClient) Do(r *http.Request) (*http.Response, error) {
	common := Common{
		Timestamp: time.Now().UTC().Unix(),
	}

	httpx.ExtendHeadersOverride(r.Header, common.Headers())

	buffer := freebuf.NewSerial()
	defer buffer.FreeMe()

	if r.Body != nil {
		_, err := buffer.ReadFrom(r.Body)
		if err != nil {
			return nil, fmt.Errorf("tencentcloud sign: read body: %w", err)
		}
		_ = r.Body.Close()
	}
	bodyBytes := buffer.Bytes()

	r.Body = io.NopCloser(buffer)
	r.ContentLength = int64(buffer.Len())
	r.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
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
	r.Header.Set("Authorization", auth)

	return tc.HTTPRequester.Do(r)
}
