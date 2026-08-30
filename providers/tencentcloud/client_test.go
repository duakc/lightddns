package tencentcloud

import (
	"context"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"testing"

	"github.com/duakc/lightddns/adapter/ddnsx"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type apiRequesterFunc func(*http.Request) (*http.Response, error)

func (f apiRequesterFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

type testRecordReader struct {
	records []ComparedRecord
}

func (r testRecordReader) Records(context.Context, ddnsx.RecordKey) ([]ComparedRecord, error) {
	return r.records, nil
}

func TestAPIErrorIncludesDetailsURL(t *testing.T) {
	require.Contains(t, (&APIError{Code: DNSPodErrCodeMustAddDefaultLineFirst}).Error(), DNSPodErrorCodeURL)
}

func TestHTTPAPIErrorIncludesResponseDetails(t *testing.T) {
	client := &defaultAPIClient{
		logger: zap.NewNop(),
		requester: apiRequesterFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body: io.NopCloser(strings.NewReader(`{
					"Response": {
						"Error": {
							"Code": "AuthFailure.InvalidSecretId",
							"Message": "invalid secret id"
						},
						"RequestId": "request-123"
					}
				}`)),
			}, nil
		}),
	}

	_, err := doAction[CreateRecordResponse](
		context.Background(), client, DNSPodActionCreateRecord, CreateRecordRequest{},
	)
	require.Error(t, err)
	require.ErrorContains(t, err, "status=400")
	require.ErrorContains(t, err, "AuthFailure.InvalidSecretId")
	require.ErrorContains(t, err, "invalid secret id")
	require.ErrorContains(t, err, "request-123")
	require.ErrorContains(t, err, DNSPodErrorCodeURL)
}

func TestBuildDiffsCreatesOneRecordPerAddressAndLine(t *testing.T) {
	client := NewClient(nil, nil, []string{"移动", "电信", "默认"})
	existing := []ComparedRecord{
		{Addr: netip.MustParseAddr("192.0.2.1"), Line: DefaultRecordLine, TTL: 300},
		{Addr: netip.MustParseAddr("192.0.2.2"), Line: DefaultRecordLine, TTL: 300},
	}
	diffs, err := client.BuildDiffs(context.Background(), ddnsx.RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("192.0.2.2"),
	}, 300, testRecordReader{records: existing})
	require.NoError(t, err)
	require.Len(t, diffs, 4)
	for _, diff := range diffs {
		require.Equal(t, ddnsx.DDNSActionCreate, diff.Action)
	}
}

func TestBuildDiffsRequiresDefaultDuringInitialization(t *testing.T) {
	client := NewClient(nil, nil, []string{"移动"})
	_, err := client.BuildDiffs(context.Background(),
		ddnsx.RecordKey{FQDN: "host.example.com"}, []netip.Addr{
			netip.MustParseAddr("192.0.2.1"),
		}, 300, testRecordReader{})
	require.ErrorIs(t, err, ErrDefaultRecordLineRequired)

	client = NewClient(nil, nil, []string{DefaultRecordLine, "移动"})
	diffs, err := client.BuildDiffs(context.Background(),
		ddnsx.RecordKey{FQDN: "host.example.com"},
		[]netip.Addr{
			netip.MustParseAddr("192.0.2.1"),
		}, 300, testRecordReader{})
	require.NoError(t, err)
	require.Len(t, diffs, 2)
}

func TestBuildDiffsRejectsExistingRecordsWithoutDefault(t *testing.T) {
	client := NewClient(nil, nil, []string{DefaultRecordLine})
	_, err := client.BuildDiffs(context.Background(), ddnsx.RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
	}, 300, testRecordReader{records: []ComparedRecord{{
		Addr: netip.MustParseAddr("192.0.2.1"), Line: "移动", TTL: 300,
	}}})
	require.NoError(t, err)
}

func TestBuildDiffsIgnoresUnconfiguredLines(t *testing.T) {
	client := NewClient(nil, nil, []string{"移动"})
	existing := []ComparedRecord{
		{Addr: netip.MustParseAddr("192.0.2.1"), Line: DefaultRecordLine, TTL: 300},
		{Addr: netip.MustParseAddr("192.0.2.9"), Line: "电信", TTL: 300},
	}
	diffs, err := client.BuildDiffs(context.Background(), ddnsx.RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
	}, 300, testRecordReader{records: existing})
	require.NoError(t, err)
	require.Len(t, diffs, 1)
	require.Equal(t, ddnsx.DDNSActionCreate, diffs[0].Action)
	require.Equal(t, "移动", diffs[0].Target.Line)
}

func TestConfiguredLines(t *testing.T) {
	lines := configuredLines(nil)
	require.Equal(t, []string{DefaultRecordLine}, lines)

	lines = configuredLines([]string{"移动"})
	require.Equal(t, []string{"移动"}, lines)

	lines = configuredLines([]string{"默认", "移动"})
	require.Equal(t, []string{"默认", "移动"}, lines)

	lines = configuredLines([]string{"默认", "移动", "移动", ""})
	require.Equal(t, []string{"默认", "移动"}, lines)
}

func TestComparedRecordSortsByLineThenAddress(t *testing.T) {
	addr := netip.MustParseAddr("192.0.2.1")
	other := netip.MustParseAddr("192.0.2.2")
	require.Less(t, (ComparedRecord{Line: DefaultRecordLine, Addr: addr}).Compare(ComparedRecord{Line: "移动", Addr: addr}), 0)
	require.Less(t, (ComparedRecord{Line: "移动", Addr: addr}).Compare(ComparedRecord{Line: "移动", Addr: other}), 0)
}

func TestComparedRecordIgnoresProviderDefaultTTL(t *testing.T) {
	addr := netip.MustParseAddr("192.0.2.1")
	local := ComparedRecord{Addr: addr, TTL: 0}
	remote := ComparedRecord{Addr: addr, TTL: 600}
	require.Equal(t, 0, local.Compare(remote))
	require.Equal(t, 0, remote.Compare(local))
}
