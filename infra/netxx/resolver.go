package netxx

import (
	"context"
	"errors"
	"fmt"
	"hash/maphash"
	"net/netip"
	"strings"
	"time"

	"github.com/duakc/lightddns/infra/common"
	"github.com/duakc/lightddns/infra/generic"
	"github.com/duakc/lightddns/infra/gos"
	"github.com/duakc/lightddns/infra/netxx/dnstransport"
	"github.com/elastic/go-freelru"
	mDns "github.com/miekg/dns"
)

const defaultCacheSize = 1024

type RcodeError struct {
	Code     int
	Excepted int
}

func (e *RcodeError) Error() string {
	return fmt.Sprintf("bad rcode: %d, excepted: %d", e.Code, e.Excepted)
}

type ResolveClient interface {
	Lookup(ctx context.Context, transport dnstransport.DNSTransport, domain string, strategy ResolveStrategy) (addresses []netip.Addr, err error)
}

type dnsCacheMessage struct {
	message *mDns.Msg
	expire  time.Time
}
type defaultResolveClient struct {
	cacheMaxItem int

	cache freelru.Cache[mDns.Question, dnsCacheMessage]
	sf    generic.SingleFlight[mDns.Question, *mDns.Msg]
}

func NewDefaultDNSResolver() ResolveClient {
	seed := maphash.MakeSeed()
	return &defaultResolveClient{
		cache: common.Must(freelru.NewSharded[mDns.Question, dnsCacheMessage](defaultCacheSize, func(question mDns.Question) uint32 {
			return uint32(hashQuestion(seed, question))
		})),
		cacheMaxItem: defaultCacheSize,
	}
}

func (r *defaultResolveClient) Lookup(ctx context.Context, transport dnstransport.DNSTransport,
	domain string, strategy ResolveStrategy) (addresses []netip.Addr, err error) {
	//
	domain = strings.TrimSpace(domain)
	if len(domain) == 0 || domain == "." {
		return nil, fmt.Errorf("empty name")
	}
	fqdn := mDns.Fqdn(domain)
	var (
		group         gos.Group
		ipv4Addresses []netip.Addr
		ipv6Addresses []netip.Addr
	)
	if strategy != ResolveIPv6 {
		group.Append("exchangeA", func(ctx context.Context) error {
			exchange, err := r.lookupToExchange(ctx, transport, fqdn, mDns.Type(mDns.TypeA))
			ipv4Addresses = exchange
			return err
		})
	}
	if strategy != ResolveIPv4 {
		group.Append("exchangeAAAA", func(ctx context.Context) error {
			exchange, err := r.lookupToExchange(ctx, transport, fqdn, mDns.Type(mDns.TypeAAAA))
			ipv6Addresses = exchange
			return err
		})
	}
	err = group.Run(ctx)
	if err != nil {
		rcodeError := &RcodeError{}
		if errors.As(err, &rcodeError) || len(ipv6Addresses)+len(ipv4Addresses) != 0 {
			// one query succeed, but an error return.
			err = nil
		} else {
			return nil, err
		}
	}
	return append(ipv4Addresses, ipv6Addresses...), nil
}

func (r *defaultResolveClient) lookupToExchange(ctx context.Context, transport dnstransport.DNSTransport,
	fqdn string, typ mDns.Type) (addresses []netip.Addr, err error) {
	question := mDns.Question{
		Name:   fqdn,
		Qtype:  uint16(typ),
		Qclass: mDns.ClassINET,
	}

	message := &mDns.Msg{
		MsgHdr: mDns.MsgHdr{
			Id:               mDns.Id(),
			RecursionDesired: true,
		},
		Compress: true,
		Question: []mDns.Question{question},
	}
	response, err := r.Exchange(ctx, transport, message)
	if err != nil {
		return nil, err
	}
	if response.Rcode != mDns.RcodeSuccess {
		return nil, &RcodeError{Code: response.Rcode, Excepted: mDns.RcodeSuccess}
	}
	return dnstransport.MessageToAddresses(response), nil
}

func (r *defaultResolveClient) Exchange(ctx context.Context, transport dnstransport.DNSTransport,
	message *mDns.Msg) (response *mDns.Msg, err error) {

	question := message.Question[0]
	cacheMessage, cached := r.cache.Get(question)
	now := time.Now()
	if cached {
		expired := cacheMessage.expire.After(now)
		if expired {
			r.cache.Remove(question)
			goto exchange
		}
		ttl := cacheMessage.expire.Sub(now) / time.Second
		copiedMessage := cacheMessage.message.Copy()
		copiedMessage.Id = message.Id
		dnstransport.OverwriteTTL(copiedMessage, uint32(ttl))
		return dnstransport.EdnsBackwards(message, copiedMessage), nil
	}
exchange:
	exchangedMessage, err, _ := r.sf.Do(question, func() (*mDns.Msg, error) {
		responseMessage, err := transport.Exchange(ctx, message)
		if err != nil {
			return nil, err
		}
		ttl := dnstransport.CalculateTTL(responseMessage)
		if r.cache.Len() > r.cacheMaxItem {
			r.cache.RemoveOldest()
		}
		r.cache.Add(question, dnsCacheMessage{
			message: responseMessage,
			expire:  time.Now().Add(time.Duration(ttl)),
		})
		return responseMessage, nil
	})
	if err != nil {
		return nil, err
	}
	// Copy on write, so here doesn't need a lock
	copiedMessage := exchangedMessage.Copy()
	copiedMessage.Id = message.Id
	return dnstransport.EdnsBackwards(message, copiedMessage), nil
}

func hashQuestion(seed maphash.Seed, q mDns.Question) uint32 {
	return uint32(maphash.Comparable[mDns.Question](seed, q))
}
