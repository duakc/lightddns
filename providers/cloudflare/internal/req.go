package internal

import (
	"net/http"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"go.uber.org/zap"
)

type CloudflareHTTPRequester struct {
	httpx.HTTPRequester

	Logger *zap.Logger
	Token  string
}

func (c *CloudflareHTTPRequester) Do(req *http.Request) (resp *http.Response, err error) {
	defer httpx.NewHTTPRequestRecorder(c.Logger, req, &resp, &err).Record()

	req.Header.Set("Authorization", "Bearer "+c.Token)
	return c.HTTPRequester.Do(req)
}
