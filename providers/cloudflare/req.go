package cloudflare

import (
	"net/http"

	"github.com/duakc/lightddns/infra/netx/httpx"

	"go.uber.org/zap"
)

type apiRequester struct {
	httpx.HTTPRequester

	Logger *zap.Logger
	Token  string
}

func (c *apiRequester) Do(req *http.Request) (resp *http.Response, err error) {

	return (&httpx.TokenClient{
		HTTPRequester: c.HTTPRequester,
		Token:         c.Token,
	}).Do(req)
}
