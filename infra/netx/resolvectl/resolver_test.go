package resolvectl

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/duakc/lightddns/infra/netx/resolvectl/transports"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func TestExchangeUpstreamRetriesRetryableTransportError(t *testing.T) {
	transport := &retryTransport{
		errs: []error{
			transports.Retryable(io.EOF),
			transports.Retryable(io.ErrClosedPipe),
		},
	}

	response, err := exchangeUpstream(context.Background(), transport, testResolverMessage(1))
	require.NoError(t, err)
	require.Equal(t, uint16(1), response.Id)
	require.Equal(t, 3, transport.calls)
}

func TestExchangeUpstreamStopsOnNonRetryableTransportError(t *testing.T) {
	transportErr := errors.New("bad response")
	transport := &retryTransport{
		errs: []error{transportErr},
	}

	_, err := exchangeUpstream(context.Background(), transport, testResolverMessage(2))
	require.ErrorIs(t, err, transportErr)
	require.Equal(t, 1, transport.calls)
}

func TestExchangeUpstreamReturnsContextAndLastRetryableError(t *testing.T) {
	transport := &retryTransport{
		errs:      []error{transports.Retryable(io.EOF)},
		repeatErr: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := exchangeUpstream(ctx, transport, testResolverMessage(3))
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorIs(t, err, io.EOF)
	require.Greater(t, transport.calls, 1)
}

type retryTransport struct {
	errs      []error
	calls     int
	repeatErr bool
}

func (t *retryTransport) Exchange(context.Context, *dns.Msg) (*dns.Msg, error) {
	t.calls++
	if len(t.errs) > 0 {
		err := t.errs[0]
		if !t.repeatErr || len(t.errs) > 1 {
			t.errs = t.errs[1:]
		}
		return nil, err
	}
	return testResolverMessage(1), nil
}

func testResolverMessage(id uint16) *dns.Msg {
	message := new(dns.Msg)
	message.SetQuestion("example.test.", dns.TypeA)
	message.Id = id
	return message
}
