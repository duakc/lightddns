package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt/freebuf"

	"go.uber.org/zap"
)

type AliyunSignClient struct {
	httpx.HTTPRequester

	Logger *zap.Logger

	SecretSecurityToken   string
	SecretAccessKeyId     string
	SecretAccessKeySecret string
}

func (c *AliyunSignClient) Do(r *http.Request) (resp *http.Response, err error) {
	defer httpx.NewHTTPRequestRecorder(c.Logger, r, &resp, &err).Record()

	common := Common{
		SecretSecurityToken: c.SecretSecurityToken,
	}
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

	buffer := freebuf.NewSerial()
	defer buffer.FreeMe()

	if r.Body != nil {
		_, err := buffer.ReadFrom(r.Body)
		if err != nil {
			return nil, fmt.Errorf("aliyun sign: read body: %w", err)
		}
		_ = r.Body.Close()
	}

	bodyBytes := buffer.Bytes()

	r.Body = io.NopCloser(buffer)
	r.ContentLength = int64(buffer.Len())
	r.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}

	// x-acs-content-sha256 is part of the signed header set. It must be set
	// on the request before computing the signature, otherwise the server's
	// recomputed canonical headers (which include it) won't match ours.
	r.Header.Set(HeaderContentSha256, sha256Hex(bodyBytes))

	sig := SigContext{
		Method:  r.Method,
		Path:    r.URL.Path,
		Query:   r.URL.Query(),
		Headers: r.Header,

		Body:                  bodyBytes,
		SecretAccessKeyId:     c.SecretAccessKeyId,
		SecretAccessKeySecret: c.SecretAccessKeySecret,
	}

	auth, err := sig.Authorization()
	if err != nil {
		return nil, fmt.Errorf("aliyun sign: authorization: %w", err)
	}
	r.Header.Set(HeaderAuthorization, auth)

	return c.HTTPRequester.Do(r)
}
