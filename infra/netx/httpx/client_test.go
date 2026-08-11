package httpx

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestDebugClientForwardsRequest(t *testing.T) {
	var called bool
	client := &DebugClient{
		Logger: zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)),
		HTTPRequester: requesterFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			require.Equal(t, http.MethodGet, req.Method)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("203.0.113.30")),
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.test/ip", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

type requesterFunc func(*http.Request) (*http.Response, error)

func (f requesterFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}
