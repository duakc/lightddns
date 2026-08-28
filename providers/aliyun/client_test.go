package aliyun

import (
	"context"
	"net/netip"
	"testing"

	"github.com/duakc/lightddns/adapter/ddnsx"

	"github.com/stretchr/testify/require"
)

type testRecordReader struct {
	records []ComparedRecord
}

func (r testRecordReader) Records(context.Context, ddnsx.RecordKey) ([]ComparedRecord, error) {
	return r.records, nil
}

func TestAPIErrorIncludesDetailsURL(t *testing.T) {
	require.Contains(t, (&APIError{Code: AlidnsErrCodeInvalidParameterLine}).Error(), AlidnsErrorCodeURL)
}

func TestBuildDiffsCreatesOneRecordPerAddressAndLine(t *testing.T) {
	client := NewClient(nil, nil, []string{"mobile", "telecom", "default"})
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
	client := NewClient(nil, nil, []string{"mobile"})
	_, err := client.BuildDiffs(context.Background(), ddnsx.RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
	}, 300, testRecordReader{})
	require.ErrorIs(t, err, ErrDefaultRecordLineRequired)

	client = NewClient(nil, nil, []string{"default", "mobile"})
	_, err = client.BuildDiffs(context.Background(), ddnsx.RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
	}, 300, testRecordReader{})
	require.ErrorIs(t, err, ErrDefaultRecordLineRequired)
}

func TestBuildDiffsRejectsExistingRecordsWithoutDefault(t *testing.T) {
	client := NewClient(nil, nil, []string{"default"})
	_, err := client.BuildDiffs(context.Background(), ddnsx.RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
	}, 300, testRecordReader{records: []ComparedRecord{{
		Addr: netip.MustParseAddr("192.0.2.1"), Line: "mobile", TTL: 300,
	}}})
	require.NoError(t, err)
}

func TestBuildDiffsIgnoresUnconfiguredLines(t *testing.T) {
	client := NewClient(nil, nil, []string{"mobile"})
	existing := []ComparedRecord{
		{Addr: netip.MustParseAddr("192.0.2.1"), Line: DefaultRecordLine, TTL: 300},
		{Addr: netip.MustParseAddr("192.0.2.9"), Line: "telecom", TTL: 300},
	}
	diffs, err := client.BuildDiffs(context.Background(), ddnsx.RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
	}, 300, testRecordReader{records: existing})
	require.NoError(t, err)
	require.Len(t, diffs, 1)
	require.Equal(t, ddnsx.DDNSActionCreate, diffs[0].Action)
	require.Equal(t, "mobile", diffs[0].Target.Line)
}

func TestConfiguredLines(t *testing.T) {
	lines := configuredLines(nil)
	require.Equal(t, []string{DefaultRecordLine}, lines)

	lines = configuredLines([]string{"telecom"})
	require.Equal(t, []string{"telecom"}, lines)

	lines = configuredLines([]string{"default", "telecom"})
	require.Equal(t, []string{"default", "telecom"}, lines)

	lines = configuredLines([]string{"default", "telecom", "telecom", ""})
	require.Equal(t, []string{"default", "telecom"}, lines)
}

func TestComparedRecordSortsByLineThenAddress(t *testing.T) {
	addr := netip.MustParseAddr("192.0.2.1")
	other := netip.MustParseAddr("192.0.2.2")
	require.Less(t, (ComparedRecord{Line: DefaultRecordLine, Addr: addr}).Compare(ComparedRecord{Line: "mobile", Addr: addr}), 0)
	require.Less(t, (ComparedRecord{Line: "mobile", Addr: addr}).Compare(ComparedRecord{Line: "mobile", Addr: other}), 0)
}
