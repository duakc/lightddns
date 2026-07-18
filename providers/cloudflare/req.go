package cloudflare

import (
	"net/http"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"go.uber.org/zap"
)

type apiRequester struct {
	httpx.HTTPRequester

	Logger *zap.Logger
	Token  string
}

func (c *apiRequester) Do(req *http.Request) (resp *http.Response, err error) {
	defer httpx.NewHTTPRequestRecorder(c.Logger, req, &resp, &err).Record()

	return (&httpx.TokenClient{
		HTTPRequester: c.HTTPRequester,
		Token:         c.Token,
	}).Do(req)
}
