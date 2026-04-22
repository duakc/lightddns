package transport

import (
	"context"
	"fmt"
	"net"

	mDns "github.com/miekg/dns"
)

const defaultTTL = 600

type SystemTransport struct{}

func (t *SystemTransport) Exchange(ctx context.Context, message *mDns.Msg) (*mDns.Msg, error) {
	question := message.Question[0]
	if question.Qtype != mDns.TypeA && question.Qtype != mDns.TypeAAAA {
		return nil, fmt.Errorf("system transport only support `A` and `AAAA` record")
	}
	ip, err := net.DefaultResolver.LookupIP(ctx, "ip", FqdnToDomain(question.Name))
	if err != nil {
		return nil, err
	}
	return FixedResponse(message.Id, question, ip, defaultTTL), nil
}
