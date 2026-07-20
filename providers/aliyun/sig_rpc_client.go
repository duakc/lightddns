package aliyun

import (
	"net/http"
	"time"

	"github.com/duakc/lightddns/infra/netx/httpx"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RpcSignRequester signs outgoing requests with the Aliyun RPC v1
// scheme: all public params (AccessKeyId, Timestamp, SignatureNonce,
// SignatureMethod, SignatureVersion, Format) are injected into the URL
// query, the canonical string is HMAC-SHA1'd with (AccessKeySecret + "&"),
// and the base64'd signature is appended as "Signature" on the URL query.
//
// The caller is expected to have already placed Action and Version on the
// URL query (see client.go's newRequest).
type RpcSignRequester struct {
	httpx.HTTPRequester

	Logger *zap.Logger

	SecretAccessKeyId     string
	SecretAccessKeySecret string
	SecretSecurityToken   string
}

func (c *RpcSignRequester) Do(r *http.Request) (resp *http.Response, err error) {
	defer httpx.NewHTTPRequestRecorder(c.Logger, r, &resp, &err).Record()

	q := r.URL.Query()
	q.Set(ParamAccessKeyId, c.SecretAccessKeyId)
	q.Set(ParamSignatureMethod, RPCSignatureMethod)
	q.Set(ParamSignatureVersion, RPCSignatureVersion)
	q.Set(ParamSignatureNonce, uuid.NewString())
	q.Set(ParamTimestamp, time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	if c.SecretSecurityToken != "" {
		q.Set(ParamSecurityToken, c.SecretSecurityToken)
	}
	if q.Get(ParamFormat) == "" {
		q.Set(ParamFormat, RPCFormatJSON)
	}

	sig := RPCSigContext{
		Method:                r.Method,
		Query:                 q,
		SecretAccessKeySecret: c.SecretAccessKeySecret,
	}
	q.Set(ParamSignature, sig.Signature())

	r.URL.RawQuery = q.Encode()
	return c.HTTPRequester.Do(r)
}
