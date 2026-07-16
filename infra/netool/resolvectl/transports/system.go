package transports

import (
	"context"
	"fmt"
	"net"

	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
)

const TransportTypeSystem = "system"

type SystemTransport struct {
	logger *zap.Logger

	fixedTTL uint32
}

func NewSystemTransport(logger *zap.Logger, ttl uint32) *SystemTransport {
	return &SystemTransport{
		logger:   createLogger(logger, TransportTypeSystem),
		fixedTTL: ttl,
	}
}

func (t *SystemTransport) Exchange(ctx context.Context, message *mDns.Msg) (*mDns.Msg, error) {
	logger := t.logger
	question := message.Question[0]
	if question.Qtype != mDns.TypeA && question.Qtype != mDns.TypeAAAA {
		logger.Info("")
		return nil, fmt.Errorf("system transports only support `A` and `AAAA` record")
	}
	ip, err := net.DefaultResolver.LookupIP(ctx, "ip", FqdnToDomain(question.Name))
	if err != nil {
		return nil, err
	}
	return FixedResponse(message.Id, question, ip, t.fixedTTL), nil
}
