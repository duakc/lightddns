package aliyun

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type requesterFunc func(*http.Request) (*http.Response, error)

func (f requesterFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestAliyunROASignClientSignsContentHashHeader(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, "https://example.com/records", strings.NewReader(`{"name":"home"}`),
	)
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")

	client := &AliyunROASignClient{
		HTTPRequester: requesterFunc(func(request *http.Request) (*http.Response, error) {
			require.NotEmpty(t, request.Header.Get(HeaderContentSha256))
			authorization := request.Header.Get(HeaderAuthorization)
			require.Contains(t, authorization, "SignedHeaders=")
			require.Contains(t, authorization, "x-acs-content-sha256")

			body, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			require.Equal(t, `{"name":"home"}`, string(body))
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}),
		Logger:                zap.NewNop(),
		SecretAccessKeyId:     "key-id",
		SecretAccessKeySecret: "key-secret",
	}

	_, err = client.Do(request)
	require.NoError(t, err)
}

var _ httpx.HTTPRequester = requesterFunc(nil)
