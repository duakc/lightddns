package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt/freebuf"
)

type AliyunSignClient struct {
	httpx.HTTPRequester

	SecretSecurityToken   string
	SecretAccessKeyId     string
	SecretAccessKeySecret string
}

func (c *AliyunSignClient) Do(r *http.Request) (resp *http.Response, err error) {
	common := Common{
		SecretSecurityToken: c.SecretSecurityToken,
	}
	httpx.ExtendHeadersOverride(r.Header, common.Headers())

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

	sig := SigContext{
		Method:  r.Method,
		Path:    r.URL.Path,
		Query:   r.URL.Query(),
		Headers: r.Header,

		Body:                  bodyBytes,
		SecretAccessKeyId:     c.SecretAccessKeyId,
		SecretAccessKeySecret: c.SecretAccessKeySecret,
	}

	auth, err := sig.Header()
	if err != nil {
		return nil, fmt.Errorf("aliyun sign: build header: %w", err)
	}

	httpx.ExtendHeadersOverride(r.Header, auth)

	return c.HTTPRequester.Do(r)
}
