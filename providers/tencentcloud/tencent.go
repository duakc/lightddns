package tencentcloud

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/providerx"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/options/castoption"

	"go.uber.org/zap"
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

	reconciler *ddnsx.Reconciler[ComparedRecord]
}

func New(ctx context.Context, logger *zap.Logger, option options.TencentCloudProviderOption) (adapter.Provider, error) {
	if option.SecretId == "" {
		return nil, fmt.Errorf("secretId is empty")
	}

	if option.SecretKey == "" {
		return nil, fmt.Errorf("secretKey is empty")
	}

	_, _, httpClient, err := castoption.BuildHTTPClientFromScratch(
		logger, option.Connect, option.DNS, option.HTTP)
	if err != nil {
		return nil, err
	}

	apiClient := NewAPIClient(logger, httpClient, option.SecretId, option.SecretKey)
	client := NewClient(logger, apiClient, option.Lines.Value)
	observed := providerx.NewMetricsClientFromContext[ComparedRecord](
		ctx, option.Name, ProviderType, client,
	)

	return &TencentCloud{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		reconciler: ddnsx.NewReconciler[ComparedRecord](
			logger, observed,
		),
	}, nil
}

func (t *TencentCloud) Close() error { return nil }

func (t *TencentCloud) Diff(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (bool, error) {
	return t.reconciler.Diff(ctx, domain, ttl, addr)
}

func (t *TencentCloud) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (bool, error) {
	return t.reconciler.Update(ctx, domain, ttl, addr)
}
