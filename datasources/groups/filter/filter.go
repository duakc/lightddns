package filter

import (
	"context"
	"fmt"
	"hash/maphash"
	"net/netip"
	"sync"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/datasourcex"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netx"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services"

	"github.com/elastic/go-freelru"
	"go.uber.org/zap"
)

const DatasourceType = constpkg.DatasourceGroupFilter

func init() {
	adapter.Register(
		adapter.DatasourceRegister,
		DatasourceType,
		New,
	)
}

var _ adapter.DatasourceDualStack = (*Filter)(nil)

type Filter struct {
	adapter.AbstractManagedType

	logger      *zap.Logger
	datasources []adapter.Datasource

	rules []Rule

	ruleMatchCache  freelru.Cache[netip.Addr, int]
	ruleMatchAccess sync.Mutex
}

func New(ctx context.Context, logger *zap.Logger, option options.DatasourceGroupFilterOption) (adapter.Datasource, error) {
	if len(option.Rules) == 0 {
		return nil, fmt.Errorf("empty rules")
	}

	var (
		datasources []adapter.Datasource
		err         error
	)

	datasources, err = datasourcex.Lookup(
		services.Lookup[adapter.DatasourceManager](ctx),
		option.Datasources...)
	if err != nil {
		return nil, err
	}

	var rules []Rule
	for ruleOptionIndex, ruleOption := range option.Rules {
		var rule Rule
		if len(ruleOption.Prefixes) > 0 {
			var buildIpSetErr error
			rule.Prefixes, buildIpSetErr = netx.BuildIPSetFromPrefixes(
				mt.Map(ruleOption.Prefixes, func(s badyaml.Prefix) netip.Prefix {
					return netip.Prefix(s)
				}))
			if buildIpSetErr != nil {
				return nil, fmt.Errorf("build rule[%d] prefixes: %w",
					ruleOptionIndex, buildIpSetErr)
			}
		}
		rule.Invert = ruleOption.Invert
		rules = append(rules, rule)
	}

	maphashSeed := maphash.MakeSeed()
	ruleMatchCache := mt.Must(freelru.New[netip.Addr, int](1024, func(addr netip.Addr) uint32 {
		return uint32(maphash.Comparable(maphashSeed, addr))
	}))

	return &Filter{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),

		logger:      logger,
		datasources: datasources,
		rules:       rules,

		ruleMatchCache: ruleMatchCache,
	}, nil
}

func (f *Filter) IP(ctx context.Context) ([]netip.Addr, error) {
	collectedIP, err := datasourcex.MergeDatasources(ctx, f.datasources, true, true, true)
	if err != nil {
		return nil, err
	}
	return f.matchRules(collectedIP), nil
}

func (f *Filter) IPv4(ctx context.Context) ([]netip.Addr, error) {
	collectedIP, err := datasourcex.MergeDatasources(ctx, f.datasources, true, false, true)
	if err != nil {
		return nil, err
	}
	return f.matchRules(collectedIP), nil
}

func (f *Filter) IPv6(ctx context.Context) ([]netip.Addr, error) {
	collectedIP, err := datasourcex.MergeDatasources(ctx, f.datasources, false, true, true)
	if err != nil {
		return nil, err
	}
	return f.matchRules(collectedIP), nil
}

func (f *Filter) matchRules(ips []netip.Addr) []netip.Addr {
	var matchIP []netip.Addr

	for _, ip := range ips {
		f.ruleMatchAccess.Lock()
		matchRuleIndex, isCacheRule := f.ruleMatchCache.Get(ip)
		if !isCacheRule {
			matchRuleIndex = matchRulesSlow(f.rules, ip)
			f.ruleMatchCache.Add(ip, matchRuleIndex)
		}
		f.ruleMatchAccess.Unlock()

		if matchRuleIndex >= 0 {
			matchIP = append(matchIP, ip)
		}
	}

	return matchIP
}

func matchRulesSlow(rules []Rule, ip netip.Addr) int {
	for ruleIndex, rule := range rules {
		if rule.Match(ip) {
			return ruleIndex
		}
	}
	return -1
}
