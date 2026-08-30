package cloudflare

import (
	"context"
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"
	"testing"

	"github.com/duakc/lightddns/infra/netx/httpx"

	"github.com/stretchr/testify/require"
)

type apiRequesterFunc func(*http.Request) (*http.Response, error)

func (f apiRequesterFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAPIRequestIncludesCloudflareErrorDetails(t *testing.T) {
	endpoint, err := urlpkg.Parse("https://api.cloudflare.test/client/v4")
	require.NoError(t, err)

	client := &defaultAPIClient{
		requester: apiRequesterFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{
					"success": false,
					"errors": [{
						"code": 1004,
						"message": "DNS validation failed",
						"documentation_url": "https://api.cloudflare.com/docs"
					}]
				}`)),
			}, nil
		}),
	}

	_, err = doAPIRequest[CreateDNSRecordResponse](
		context.Background(),
		client,
		httpx.ReqConfig{Method: http.MethodPost, BaseURL: endpoint},
	)
	require.Error(t, err)
	require.ErrorContains(t, err, "unacceptable status code: 400")
	require.ErrorContains(t, err, "code=1004")
	require.ErrorContains(t, err, "DNS validation failed")
	require.ErrorContains(t, err, "https://api.cloudflare.com/docs")
}
