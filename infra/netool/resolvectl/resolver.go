package resolvectl

import (
	"context"
	"fmt"
	"hash/maphash"
	"net/netip"
	"strings"
	"time"

	"github.com/duakc/lightddns/infra/gos"
	"github.com/duakc/lightddns/infra/netool/resolvectl/transports"
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"
	"github.com/duakc/mt/common/generic"
	"github.com/duakc/mt/debug"
	"github.com/duakc/mt/services"

	"github.com/elastic/go-freelru"
	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultCacheSize = 1024

var resolverLogger = zaplog.NewPackage("netool", "resolvectl")

type RcodeError struct {
	Code     int
	Excepted int
}

func (e *RcodeError) Error() string {
	return fmt.Sprintf("bad rcode: %s, excepted: %s", mDns.RcodeToString[e.Code], mDns.RcodeToString[e.Excepted])
}

var DefaultResolveClient = NewResolver(context.Background())

type ResolveClient interface {
	Lookup(ctx context.Context, transport transports.Transport, domain string, strategy ResolveStrategy) (addresses []netip.Addr, err error)
}

type dnsCacheMessage struct {
	message *mDns.Msg
	expire  time.Time
}

type defaultResolveClient struct {
	logger *zap.Logger

	// cache
	cacheMax int
	sf       generic.SingleFlight[mDns.Question, *mDns.Msg]
	cache    freelru.Cache[mDns.Question, dnsCacheMessage]
}

func NewResolver(ctx context.Context) ResolveClient {
	seed := maphash.MakeSeed()
	logger := services.LookupPtrDefault[zap.Logger](ctx, resolverLogger)
	return &defaultResolveClient{
		logger: logger.Named("resolver"),
		cache: mt.Must(freelru.NewSharded[mDns.Question, dnsCacheMessage](defaultCacheSize,
			(&bindSeed[mDns.Question]{seed}).hash())),
		cacheMax: defaultCacheSize,
	}
}

func (r *defaultResolveClient) Lookup(ctx context.Context, transport transports.Transport,
	domain string, strategy ResolveStrategy,
) (addresses []netip.Addr, err error) {
	//
	domain = strings.TrimSpace(domain)
	if len(domain) == 0 || domain == "." {
		return nil, fmt.Errorf("empty name")
	}
	fqdn := mDns.Fqdn(domain)
	if strategy == ResolveIPv4 || strategy == ResolveIPv6 {
		return r.lookupFast(ctx, transport, fqdn, strategy)
	}
	var (
		group         gos.Group
		ipv4Addresses []netip.Addr
		ipv6Addresses []netip.Addr
	)
	group.Append("exchangeA", func(ctx context.Context) error {
		exchange, err := r.lookupToExchange(ctx, transport, fqdn, mDns.TypeA)
		ipv4Addresses = exchange
		return err
	})
	group.Append("exchangeAAAA", func(ctx context.Context) error {
		exchange, err := r.lookupToExchange(ctx, transport, fqdn, mDns.TypeAAAA)
		ipv6Addresses = exchange
		return err
	})
	err = group.Run(ctx)
	if err != nil {
		if len(ipv6Addresses)+len(ipv4Addresses) != 0 {
			r.logger.Error("lookup failed but addresses returned",
				zap.Error(err),
				zap.String("fqdn", fqdn))
		} else {
			return nil, err
		}
	}
	return append(ipv4Addresses, ipv6Addresses...), nil
}

func (r *defaultResolveClient) lookupFast(ctx context.Context,
	transport transports.Transport,
	fqdn string, strategy ResolveStrategy,
) (addresses []netip.Addr, err error) {
	if debug.Enabled && strategy != ResolveIPv4 && strategy != ResolveIPv6 {
		panic("unexcepted")
	}
	if strategy == ResolveIPv6 {
		addresses, err = r.lookupToExchange(ctx, transport, fqdn, mDns.TypeAAAA)
	} else {
		addresses, err = r.lookupToExchange(ctx, transport, fqdn, mDns.TypeA)
	}

	if err != nil {
		msg := "exchangeA"
		if strategy == ResolveIPv6 {
			msg = "exchangeAAAA"
		}
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	return addresses, nil
}

func (r *defaultResolveClient) lookupToExchange(ctx context.Context,
	dnsTransport transports.Transport,
	fqdn string, typ uint16,
) (addresses []netip.Addr, err error) {
	question := mDns.Question{
		Name:   fqdn,
		Qtype:  typ,
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
	response, err := r.Exchange(ctx, dnsTransport, message)
	if err != nil {
		return nil, err
	}
	if response.Rcode != mDns.RcodeSuccess {
		return nil, &RcodeError{Code: response.Rcode, Excepted: mDns.RcodeSuccess}
	}
	return transports.MessageToAddresses(response), nil
}

func (r *defaultResolveClient) Exchange(ctx context.Context, dnsTransport transports.Transport,
	message *mDns.Msg,
) (response *mDns.Msg, err error) {
	if len(message.Question) != 1 {
		r.logger.Error("bad question length", zap.Int("length", len(message.Question)))
		return nil, &RcodeError{Code: mDns.RcodeNameError}
	}
	question := message.Question[0]

	logger := r.logger.WithLazy(
		zap.String("type", mDns.TypeToString[question.Qtype]),
		zap.String("fqdn", question.Name))
	cacheMessage, cached := r.cache.Get(question)
	now := time.Now()
	logger.Debug("new exchange")
	if cached {
		expired := cacheMessage.expire.After(now)
		logger.Debug("cached", zap.Bool("expired", expired))
		if expired {
			r.cache.Remove(question)
			goto exchange
		}
		ttl := cacheMessage.expire.Sub(now) / time.Second
		copiedMessage := cacheMessage.message.Copy()
		copiedMessage.Id = message.Id
		transports.OverwriteTTL(copiedMessage, uint32(ttl))
		return transports.EdnsBackwards(message, copiedMessage), nil
	}
exchange:
	exchangedMessage, err, _ := r.sf.Do(question, func() (*mDns.Msg, error) {
		logger.Debug("start a new exchange from upstream")
		responseMessage, err := dnsTransport.Exchange(ctx, message)
		if err != nil {
			return nil, err
		}
		ttl := transports.CalculateTTL(responseMessage)
		if r.cache.Len() > r.cacheMax {
			r.cache.RemoveOldest()
		}
		if ce := logger.Check(zapcore.DebugLevel, "exchanged"); ce != nil {
			var fields = []zap.Field{
				zap.String("type", mDns.TypeToString[question.Qtype]),
				zap.String("fqdn", question.Name),
				zap.Uint32("ttl", ttl),
			}
			if addresses := transports.MessageToAddresses(responseMessage); len(addresses) > 0 {
				fields = append(fields, zap.Stringers("addresses", addresses))
			}
			ce.Write(fields...)
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
	return transports.EdnsBackwards(message, copiedMessage), nil
}

type bindSeed[T comparable] struct {
	seed maphash.Seed
}

func (qh *bindSeed[T]) hash() func(question T) uint32 {
	return func(question T) uint32 {
		return uint32(maphash.Comparable(qh.seed, question))
	}
}
