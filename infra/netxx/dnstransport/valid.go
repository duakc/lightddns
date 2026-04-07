package dnstransport

import (
	"context"
	"fmt"

	"github.com/duakc/lightddns/infra/zaplog"
	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
)

type ValidTransport struct {
	Logger    *zap.Logger
	Transport DNSTransport
}

func (v *ValidTransport) Exchange(ctx context.Context, message *mDns.Msg) (*mDns.Msg, error) {
	logger := zaplog.DoNotPanic(v.Logger)
	defer logger.Sync()
	if message == nil {
		panic("nil message")
	}
	if len(message.Question) == 0 {
		return nil, fmt.Errorf("empty question")
	}
	if len(message.Question) > 1 {
		logger.Warn("the number of `messages.Question` is greater than 1, discard the `message.Question` part index greater than 0")
		message.Question = append([]mDns.Question{}, message.Question[0])
	}
	return v.Transport.Exchange(ctx, message)
}
