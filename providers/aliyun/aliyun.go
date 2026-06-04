package aliyun

import (
	"context"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"

	"go.uber.org/zap"
)

const ProviderType = constpkg.ProviderTypeAliyun

var _ adapter.Provider = (*Aliyun)(nil)

func init() {
	adapter.Register(
		adapter.ProviderRegister,
		ProviderType,
		New,
	)
}

func New(ctx context.Context, logger *zap.Logger, option options.AliyunProviderOption) (adapter.Provider, error) {
	// TODO implement me
	panic("implement me")
}

type Aliyun struct {
	adapter.AbstractManagedType
}

func (a *Aliyun) Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (a *Aliyun) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (changed bool, err error) {
	// TODO implement me
	panic("implement me")
}
