package resolvectl

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/duakc/lightddns/infra/gos"
	"github.com/duakc/lightddns/infra/netool/resolvectl/transports"
	"github.com/duakc/mt/debug"

	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ResolveClient interface {
	Lookup(ctx context.Context, transport transports.Transport, domain string, strategy LookupStrategy) (addresses []netip.Addr, err error)
	Exchange(ctx context.Context, dnsTransport transports.Transport,
		message *mDns.Msg,
	) (response *mDns.Msg, err error)
}

type defaultResolveClient struct {
	logger *zap.Logger

	resolverCache ResolverCache
}

func NewResolverWithCache(logger *zap.Logger, cache ResolverCache) ResolveClient {
	return &defaultResolveClient{
		logger:        logger,
		resolverCache: cache,
	}
}

func NewResolver(logger *zap.Logger) ResolveClient {
	return NewResolverWithCache(logger, defaultResolverCache)
}

func (r *defaultResolveClient) Lookup(ctx context.Context, transport transports.Transport,
	domain string, strategy LookupStrategy,
) (addresses []netip.Addr, err error) {
	logger := r.logger.
		WithLazy(zap.String("lookup_fqdn", domain))
	//
	domain = strings.TrimSpace(domain)
	if len(domain) == 0 || domain == "." {
		return nil, fmt.Errorf("empty name")
	}
	fqdn := mDns.Fqdn(domain)
	if strategy == ResolveIPv4 || strategy == ResolveIPv6 {
		return r.lookupFast(ctx, logger, transport, fqdn, strategy)
	}
	var (
		group         gos.Group
		ipv4Addresses []netip.Addr
		ipv6Addresses []netip.Addr
	)
	group.Append("exchangeA", func(ctx context.Context) error {
		exchange, err := r.lookupToExchange(ctx, logger, transport, fqdn, mDns.TypeA)
		ipv4Addresses = exchange
		return err
	})
	group.Append("exchangeAAAA", func(ctx context.Context) error {
		exchange, err := r.lookupToExchange(ctx, logger, transport, fqdn, mDns.TypeAAAA)
		ipv6Addresses = exchange
		return err
	})
	err = group.Run(ctx)
	if err != nil {
		if len(ipv6Addresses)+len(ipv4Addresses) != 0 {
			logger.Error("lookup failed but addresses returned",
				zap.Error(err))
		} else {
			return nil, err
		}
	}
	return append(ipv4Addresses, ipv6Addresses...), nil
}

func (r *defaultResolveClient) lookupFast(ctx context.Context,
	logger *zap.Logger,
	transport transports.Transport,
	fqdn string, strategy LookupStrategy,
) (addresses []netip.Addr, err error) {
	if debug.Enabled && strategy != ResolveIPv4 && strategy != ResolveIPv6 {
		panic("unexcepted")
	}
	if strategy == ResolveIPv6 {
		addresses, err = r.lookupToExchange(ctx, logger, transport, fqdn, mDns.TypeAAAA)
	} else {
		addresses, err = r.lookupToExchange(ctx, logger, transport, fqdn, mDns.TypeA)
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
	logger *zap.Logger,
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

	response, err := exchangeOnce(ctx, logger, dnsTransport, message, r.resolverCache)
	if err != nil {
		return nil, err
	}

	if response.Rcode != mDns.RcodeSuccess {
		return nil, &transports.RcodeError{Code: response.Rcode, Excepted: mDns.RcodeSuccess}
	}
	return transports.MessageToAddresses(response), nil
}

func (r *defaultResolveClient) Exchange(ctx context.Context, dnsTransport transports.Transport,
	message *mDns.Msg,
) (responseMessage *mDns.Msg, err error) {
	logger := r.logger

	if len(message.Question) != 1 {
		logger.Error("bad question length",
			zap.Int("length", len(message.Question)))
		return nil, &transports.RcodeError{Code: mDns.RcodeNameError}
	}

	question := message.Question[0]

	logger = logger.WithLazy(
		zap.String("exchange_type", mDns.TypeToString[question.Qtype]),
		zap.String("exchange_fqdn", question.Name))

	return exchangeOnce(ctx, logger, dnsTransport, message, r.resolverCache)
}

func exchangeOnce(ctx context.Context,
	logger *zap.Logger,
	dnsTransport transports.Transport,
	message *mDns.Msg,
	cacheFrom ResolverCache,
) (responseMessage *mDns.Msg, err error) {
	var cacheErr error
	responseMessage, cacheErr = cacheFrom.ExchangeCache(message)
	if cacheErr != nil && !errors.Is(cacheErr, ErrCacheMiss) {
		logger.Info("load from cache failed", zap.Error(cacheErr))
	}

	if cacheErr == nil {
		logger.Info("load from cache success")
		return responseMessage, nil
	}

	if ctx == nil {
		ctx = context.Background()
	}
	responseMessage, err = cacheFrom.LoadUpstream(ctx, transports.FuncTransport(func(exchangeCtx context.Context, exchangeMessage *mDns.Msg) (*mDns.Msg, error) {
		exchangedMessage, exchangeErr := dnsTransport.Exchange(exchangeCtx, exchangeMessage)
		if exchangeErr != nil {
			return nil, exchangeErr
		}

		if ce := logger.Check(zapcore.DebugLevel, "exchanged"); ce != nil {
			ttl := transports.CalculateTTL(exchangeMessage)

			fields := []zap.Field{
				zap.Uint32("ttl", ttl),
			}

			if addresses := transports.MessageToAddresses(responseMessage); len(addresses) > 0 {
				fields = append(fields, zap.Stringers("addresses", addresses))
			}

			ce.Write(fields...)
		}
		return exchangedMessage, nil
	}), message)

	// cow
	copiedMessage := responseMessage.Copy()
	copiedMessage.Id = message.Id
	return transports.EdnsBackwards(message, copiedMessage), nil
}
