package httpx

import (
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadAndReplayBody(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest(http.MethodPost, "https://example.com", strings.NewReader("payload"))
	require.NoError(t, err)

	body, err := ReadAndReplayBody(request)
	require.NoError(t, err)
	require.Equal(t, []byte("payload"), body)
	require.Equal(t, int64(len(body)), request.ContentLength)

	replayed, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Equal(t, body, replayed)

	copyBody, err := request.GetBody()
	require.NoError(t, err)
	defer copyBody.Close()
	copied, err := io.ReadAll(copyBody)
	require.NoError(t, err)
	require.Equal(t, body, copied)
}

func TestExtendHeadersOverridePreservesMultipleValues(t *testing.T) {
	t.Parallel()

	source := http.Header{"Accept": {"text/plain"}}
	extended := http.Header{"Accept": {"application/json", "application/problem+json"}}
	ExtendHeadersOverride(source, extended)

	require.Equal(t, extended["Accept"], source["Accept"])
	extended["Accept"][0] = "changed"
	require.Equal(t, "application/json", source["Accept"][0])
}

func TestIsJsonContentTypeAcceptsStructuredSuffix(t *testing.T) {
	t.Parallel()

	require.True(t, IsJsonContentType("application/problem+json; charset=utf-8"))
	require.False(t, IsJsonContentType("text/plain"))
}

func TestRedactQueryAndHeader(t *testing.T) {
	t.Parallel()

	query := urlpkg.Values{
		"Action":        {"DescribeDomains"},
		"AccessKeyId":   {"key"},
		"SecurityToken": {"token"},
		"Signature":     {"signature"},
	}
	redactedQuery := RedactQuery(query)
	require.Equal(t, "DescribeDomains", redactedQuery.Get("Action"))
	require.Equal(t, redactedValue, redactedQuery.Get("AccessKeyId"))
	require.Equal(t, redactedValue, redactedQuery.Get("SecurityToken"))
	require.Equal(t, redactedValue, redactedQuery.Get("Signature"))
	require.Equal(t, "key", query.Get("AccessKeyId"))

	header := http.Header{
		"Authorization": {"Bearer token"},
		"Content-Type":  {"application/json"},
	}
	redactedHeader := RedactHeader(header)
	require.Equal(t, redactedValue, redactedHeader.Get("Authorization"))
	require.Equal(t, "application/json", redactedHeader.Get("Content-Type"))
	require.Equal(t, "Bearer token", header.Get("Authorization"))
}

type requesterFunc func(*http.Request) (*http.Response, error)

func (f requesterFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestTokenClientOverridesExistingAuthorization(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	request.Header.Add(HeaderAuthorization, "old")

	client := &TokenClient{
		HTTPRequester: requesterFunc(func(request *http.Request) (*http.Response, error) {
			require.Equal(t, []string{"Bearer new"}, request.Header.Values(HeaderAuthorization))
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}),
		Token: "new",
	}

	_, err = client.Do(request)
	require.NoError(t, err)
}
