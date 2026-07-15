package resolvectl

import (
	"context"
	"errors"
	"fmt"
	"hash/maphash"
	"time"

	"github.com/duakc/lightddns/infra/netool/resolvectl/transports"

	"github.com/duakc/mt"
	"github.com/duakc/mt/common/generic"

	"github.com/elastic/go-freelru"
	mDns "github.com/miekg/dns"
)

var (
	ErrBadMessage   = errors.New("bad message")
	ErrCacheMiss    = errors.New("miss")
	ErrCacheExpired = errors.New("expired")
)

type ResolverCache interface {
	ExchangeCache(msg *mDns.Msg) (*mDns.Msg, error)

	Store(msg *mDns.Msg) bool
	LoadUpstream(ctx context.Context,
		upstream transports.Transport, message *mDns.Msg) (*mDns.Msg, error)

	Clear()
}

type dnsMsgCacheEntry struct {
	expires time.Time
	msg     *mDns.Msg
}

// global cache.
var defaultResolverCache = NewResolverCache()

func NewResolverCache() ResolverCache {
	// value from https://coredns.io/plugins/cache/#capacity-and-eviction
	return NewResolverCacheSize(9984)
}

func NewResolverCacheSize(size uint32) ResolverCache {
	maphashSeed := bindSeed[mDns.Question]{seed: maphash.MakeSeed()}
	return &stubResolverCache{
		cache: mt.Must(freelru.NewSharded[mDns.Question, *dnsMsgCacheEntry](size, maphashSeed.hash())),
	}
}

type stubResolverCache struct {
	cache freelru.Cache[mDns.Question, *dnsMsgCacheEntry]

	sf generic.SingleFlight[mDns.Question, *dnsMsgCacheEntry]
}

func (c *stubResolverCache) ExchangeCache(message *mDns.Msg) (*mDns.Msg, error) {
	if message == nil || len(message.Question) != 1 {
		return nil, ErrBadMessage
	}
	// make sure the time is generated before Get.
	now := time.Now()

	question := message.Question[0]
	cachedMessage, existed := c.cache.Get(question)
	if !existed {
		return nil, ErrCacheMiss
	}

	ttl := cachedMessage.expires.Sub(now) / time.Second
	if ttl <= 0 {
		// expired
		c.cache.Remove(question)
		return nil, ErrCacheExpired
	}

	copiedMessage := cachedMessage.msg.Copy()

	copiedMessage.Id = message.Id
	transports.OverwriteTTL(copiedMessage, uint32(ttl))
	return transports.EdnsBackwards(message, copiedMessage), nil
}

func (c *stubResolverCache) Store(msg *mDns.Msg) bool {
	if msg == nil || len(msg.Question) != 1 {
		return false
	}

	ttl := transports.CalculateTTL(msg)
	if ttl <= 1 {
		// too short
		return false
	}

	question := msg.Question[0]

	expireTime := time.Now().Add(time.Duration(ttl) * time.Second)
	cacheEntry := &dnsMsgCacheEntry{
		expires: expireTime,
		msg:     msg,
	}

	c.cache.Add(question, cacheEntry)
	return true
}

func (c *stubResolverCache) storeMsg(msg *mDns.Msg) *dnsMsgCacheEntry {
	ttl := transports.CalculateTTL(msg)

	question := msg.Question[0]

	expireTime := time.Now().Add(time.Duration(ttl) * time.Second)
	cacheEntry := &dnsMsgCacheEntry{
		expires: expireTime,
		msg:     msg,
	}

	if ttl > 1 {
		c.cache.Add(question, cacheEntry)
	}

	// even the TTL is too short, but we still return a valid dnsMsgCacheEntry
	// for upstream get this time query.
	return cacheEntry
}

func (c *stubResolverCache) Clear() {
	c.cache.Purge()
}

func (c *stubResolverCache) LoadUpstream(ctx context.Context,
	upstream transports.Transport, message *mDns.Msg,
) (*mDns.Msg, error) {
	if message == nil {
		return nil, fmt.Errorf("nil message")
	}
	if len(message.Question) != 1 {
		return nil, fmt.Errorf("bad question size: %d", len(message.Question))
	}
	question := message.Question[0]
	exchangedMessage, err, _ := c.sf.Do(question, func() (*dnsMsgCacheEntry, error) {
		responseMessage, err := upstream.Exchange(ctx, message)
		if err != nil {
			return nil, fmt.Errorf("exchange: %w", err)
		}

		return c.storeMsg(responseMessage), nil
	})

	if err != nil {
		return nil, err
	}

	// Copy on write, so here doesn't need a lock
	copiedMessage := exchangedMessage.msg.Copy()
	copiedMessage.Id = message.Id
	return transports.EdnsBackwards(message, copiedMessage), nil
}

type bindSeed[T comparable] struct {
	seed maphash.Seed
}

func (qh bindSeed[T]) hash() func(question T) uint32 {
	return func(question T) uint32 {
		return uint32(maphash.Comparable(qh.seed, question))
	}
}
