package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"go.uber.org/zap"
)

type AliyunROASignClient struct {
	httpx.HTTPRequester

	Logger *zap.Logger

	SecretAccessKeyId     string
	SecretAccessKeySecret string
	SecretSecurityToken   string
}

func (c *AliyunROASignClient) Do(r *http.Request) (resp *http.Response, err error) {
	defer httpx.NewHTTPRequestRecorder(c.Logger, r, &resp, &err).Record()

	common := ROACommon{SecretSecurityToken: c.SecretSecurityToken}
	httpx.ExtendHeadersOverride(r.Header, common.Headers())

	// V3 signing requires Host in the canonical header set. net/http carries
	// Host on the request struct rather than in Header, so backfill it.
	if r.Header.Get("Host") == "" {
		host := r.Host
		if host == "" {
			host = r.URL.Host
		}
		r.Header.Set("Host", host)
	}

	// Drain the request body into a Go-owned []byte for the V3 signature.
	// We must not back r.Body / GetBody with a pooled buffer: the HTTP
	// transport may still be writing the body in a goroutine after Do
	// returns (HTTP/2 streams, retries via GetBody), and recycling the
	// backing slice underneath the transport produces a corrupted request
	// → server resets the stream → caller sees io.EOF on resp.Body. See
	// tencentcloud's sig_client.go for the same fix.
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("aliyun sign: read body: %w", err)
		}
		_ = r.Body.Close()
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	r.ContentLength = int64(len(bodyBytes))
	r.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}

	sig := ROASigContext{
		Method:  r.Method,
		Path:    r.URL.Path,
		Query:   r.URL.Query(),
		Headers: r.Header,

		Body:                  bodyBytes,
		SecretAccessKeyId:     c.SecretAccessKeyId,
		SecretAccessKeySecret: c.SecretAccessKeySecret,
	}

	canonicalRequest, signedHeaders, hashedRequestPayload, err := sig.CanonicalRequest()
	if err != nil {
		return nil, fmt.Errorf("aliyun sign: canonical request: %w", err)
	}

	// x-acs-content-sha256 is part of the signed header set. It must be set
	// on the request before computing the signature, otherwise the server's
	// recomputed canonical headers (which include it) won't match ours.
	r.Header.Set(HeaderContentSha256, hashedRequestPayload)

	stringToSign := sig.StringToSign(canonicalRequest)
	r.Header.Set(HeaderAuthorization, sig.BuildAuthorization(stringToSign, signedHeaders))

	return c.HTTPRequester.Do(r)
}
