package tencentcloud

import (
	"context"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"
)

const ProviderType = constpkg.ProviderTypeTencentCloud

func init() {
	adapter.Register(
		adapter.ProviderRegister,
		ProviderType,
		New,
	)
}

var _ adapter.Provider = (*TencentCloud)(nil)

type TencentCloud struct {
	adapter.AbstractManagedType
}

func New(ctx context.Context, option options.TencentCloudProviderOption) (adapter.Provider, error) {
	panic("implement me")
}

func (t *TencentCloud) Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error) {
	panic("implement me")
}

func (t *TencentCloud) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (changed bool, err error) {
	panic("implement me")
}
