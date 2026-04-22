package transport

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/duakc/mt"

	mDns "github.com/miekg/dns"
)

type Transport interface {
	Exchange(ctx context.Context, message *mDns.Msg) (*mDns.Msg, error)
}

func NewDNSTransport(buildString string) (Transport, error) {
	if buildString == "" {
		return &SystemTransport{}, nil
	}
	var schema, value string
	if idx := strings.Index(buildString, "://"); idx > 0 {
		schema = buildString[:idx]
		value = buildString[idx+len("://"):]
	} else {
		if buildString == "system" {
			schema = "system"
		} else if idx == 0 {
			buildString = buildString[len("://"):]
			schema = "udp"
		}
		value = buildString
	}
	_ = value
	switch schema {
	case "system":
		return &SystemTransport{}, nil
	default:
		// TODO: add more transport
		return nil, fmt.Errorf("transport build failed for: %s", buildString)
	}
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
