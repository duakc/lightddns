package transports

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strings"

	"github.com/duakc/lightddns/infra/netx/internal"
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"
	"github.com/duakc/mt/freebuf"

	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
)

type Transport interface {
	Exchange(ctx context.Context, message *mDns.Msg) (*mDns.Msg, error)
}

type FuncTransport func(ctx context.Context, message *mDns.Msg) (*mDns.Msg, error)

func (f FuncTransport) Exchange(ctx context.Context, message *mDns.Msg) (*mDns.Msg, error) {
	return f(ctx, message)
}

func CalculateTTL(message *mDns.Msg) (ttl uint32) {
	for _, rrs := range [][]mDns.RR{message.Answer, message.Ns, message.Extra} {
		for _, rr := range rrs {
			if ttl == 0 || ttl > rr.Header().Ttl {
				ttl = rr.Header().Ttl
			}
		}
	}
	return ttl
}

func OverwriteTTL(message *mDns.Msg, ttl uint32) {
	for _, rrs := range [][]mDns.RR{message.Answer, message.Ns, message.Extra} {
		for _, rr := range rrs {
			rr.Header().Ttl = ttl
		}
	}
}

func EdnsBackwards(req *mDns.Msg, resp *mDns.Msg) *mDns.Msg {
	requestEdns0 := req.IsEdns0()
	responseEdns0 := resp.IsEdns0()
	if responseEdns0 != nil && (requestEdns0 == nil || requestEdns0.Version() < responseEdns0.Version()) {
		resp.Extra = mt.Filter(resp.Extra, func(it mDns.RR) bool {
			return it.Header().Rrtype != mDns.TypeOPT
		})
		if requestEdns0 != nil {
			resp.SetEdns0(requestEdns0.UDPSize(), responseEdns0.Do())
		}
	}
	return resp
}

func FqdnToDomain(fqdn string) string {
	if fqdn[len(fqdn)-1] == '.' {
		return fqdn[:len(fqdn)-1]
	}
	return fqdn
}

func FixedResponse(id uint16, question mDns.Question, addresses []net.IP, timeToLive uint32) *mDns.Msg {
	if timeToLive == 0 {
		timeToLive = internal.DefaultDNSTTL
	}

	response := mDns.Msg{
		MsgHdr: mDns.MsgHdr{
			Id:                 id,
			Response:           true,
			Authoritative:      true,
			RecursionDesired:   true,
			RecursionAvailable: true,
			Rcode:              mDns.RcodeSuccess,
		},
		Question: []mDns.Question{question},
	}
	for _, address := range addresses {
		if question.Qtype == mDns.TypeA && address.To4() != nil {
			response.Answer = append(response.Answer, &mDns.A{
				Hdr: mDns.RR_Header{
					Name:   question.Name,
					Rrtype: mDns.TypeA,
					Class:  mDns.ClassINET,
					Ttl:    timeToLive,
				},
				A: address.To4(),
			})
		} else if question.Qtype == mDns.TypeAAAA && address.To16() != nil {
			response.Answer = append(response.Answer, &mDns.AAAA{
				Hdr: mDns.RR_Header{
					Name:   question.Name,
					Rrtype: mDns.TypeAAAA,
					Class:  mDns.ClassINET,
					Ttl:    timeToLive,
				},
				AAAA: address.To16(),
			})
		}
	}
	return &response
}

func MessageToAddresses(response *mDns.Msg) (addresses []netip.Addr) {
	addresses = make([]netip.Addr, 0, len(response.Answer))
	ipToAddr := func(ip net.IP) {
		if netipAddr, ok := netip.AddrFromSlice(ip); ok {
			addresses = append(addresses, netipAddr.Unmap())
		}
	}
	for _, rawAnswer := range response.Answer {
		switch answer := rawAnswer.(type) {
		case *mDns.A:
			ipToAddr(answer.A)
		case *mDns.AAAA:
			ipToAddr(answer.AAAA)
		case *mDns.HTTPS:
			for _, value := range answer.SVCB.Value {
				if value.Key() == mDns.SVCB_IPV4HINT || value.Key() == mDns.SVCB_IPV6HINT {
					addresses = append(addresses, mt.Map[string, netip.Addr](strings.Split(value.String(), ","), func(it string) netip.Addr {
						a, _ := netip.ParseAddr(it)
						return a
					})...)
				}
			}
		}
	}
	return addresses
}

func WriteMessage(w io.Writer, messageID uint16, message *mDns.Msg) error {
	exMessage := *message
	exMessage.Id = messageID
	exMessage.Compress = true

	packedMessage, err := exMessage.Pack()
	if err != nil {
		return err
	}
	if len(packedMessage) > mDns.MaxMsgSize {
		return fmt.Errorf("dns message too large: %d", len(packedMessage))
	}

	var messageLen [2]byte
	binary.BigEndian.PutUint16(messageLen[:], uint16(len(packedMessage)))
	if err := writeFull(w, messageLen[:]); err != nil {
		return err
	}
	return writeFull(w, packedMessage)
}

func ReadMessage(r io.Reader) (*mDns.Msg, error) {
	var responseLen uint16
	err := binary.Read(r, binary.BigEndian, &responseLen)
	if err != nil {
		return nil, err
	}
	if responseLen < 12 {
		return nil, mDns.ErrShortRead
	}
	buffer := freebuf.NewSerial()
	defer buffer.FreeMe()
	buffer.Grow(int(responseLen))

	_, err = freebuf.ReadFull(r, buffer, int(responseLen))
	if err != nil {
		return nil, err
	}

	var message mDns.Msg
	err = message.Unpack(buffer.Bytes())
	return &message, err
}

func writeFull(w io.Writer, p []byte) error {
	for len(p) > 0 {
		nn, err := w.Write(p)
		if err != nil {
			return err
		}
		if nn == 0 {
			return io.ErrNoProgress
		}
		p = p[nn:]
	}
	return nil
}

func createLogger(logger *zap.Logger, transportType string) *zap.Logger {
	if logger == nil || logger == zaplog.NOP {
		return logger
	}
	return logger.WithLazy(zap.String("transport_type", transportType))
}
