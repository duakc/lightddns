package aliyun

import (
	"fmt"
	"net/http"

	"github.com/duakc/lightddns/infra/netx/httpx"

	"go.uber.org/zap"
)

type RoaSignRequester struct {
	httpx.HTTPRequester

	Logger *zap.Logger

	SecretAccessKeyId     string
	SecretAccessKeySecret string
	SecretSecurityToken   string
}

func (c *RoaSignRequester) Do(r *http.Request) (resp *http.Response, err error) {
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

	bodyBytes, err := httpx.ReadAndReplayBody(r)
	if err != nil {
		return nil, fmt.Errorf("aliyun sign: %w", err)
	}

	// X-Acs-Content-Sha256 belongs to the signed header set, so it must be
	// present before CanonicalRequest builds the canonical headers.
	r.Header.Set(HeaderContentSha256, sha256Hex(bodyBytes))

	sig := ROASigContext{
		Method:  r.Method,
		Path:    r.URL.Path,
		Query:   r.URL.Query(),
		Headers: r.Header,

		Body:                  bodyBytes,
		SecretAccessKeyId:     c.SecretAccessKeyId,
		SecretAccessKeySecret: c.SecretAccessKeySecret,
	}

	canonicalRequest, signedHeaders, _, err := sig.CanonicalRequest()
	if err != nil {
		return nil, fmt.Errorf("aliyun sign: canonical request: %w", err)
	}

	stringToSign := sig.StringToSign(canonicalRequest)
	r.Header.Set(HeaderAuthorization, sig.BuildAuthorization(stringToSign, signedHeaders))

	return c.HTTPRequester.Do(r)
}
